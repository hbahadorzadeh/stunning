package plugin

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
)

// DefaultMaxFrame caps a single encoded frame. It bounds read buffers and
// rejects corrupt/probe traffic that claims an absurd length.
const DefaultMaxFrame = 16 * 1024

// Framer is an optional capability a plugin can implement to provide its own
// wire framing instead of FramedConn's default masked length prefix. Protocol
// mimicry plugins (TLS, HTTP) implement it so the bytes on the wire begin with a
// convincing protocol header and carry the protocol's native length fields,
// rather than a giveaway random length prefix. When the outermost plugin in a
// chain is a Framer, FramedConn delegates all framing to it.
type Framer interface {
	// Frame wraps one already-encoded message in protocol framing for the wire.
	Frame(payload []byte) ([]byte, error)
	// Deframe reads exactly one message back off the wire.
	Deframe(r io.Reader) ([]byte, error)
}

// ErrFrameTooLarge is returned when a frame exceeds the configured cap.
var ErrFrameTooLarge = errors.New("plugin: frame exceeds max size")

// FramedConn turns a byte-stream net.Conn into a message stream that runs every
// Write through the chain's Encode and every read through Decode. Each frame is
// length-prefixed; the 2-byte length is XOR-masked with a per-direction
// ChaCha20 keystream so the wire shows no plaintext length field. Read presents
// the decoded payloads back as a byte stream, so existing io.Copy-based plumbing
// keeps working.
//
// The two peers must construct FramedConn with opposite isClient values so their
// directional keystreams line up.
type FramedConn struct {
	net.Conn
	chain    *Chain
	maxFrame int
	framer   Framer // non-nil when the chain provides protocol-native framing

	wmu   sync.Mutex
	wMask cipher.Stream

	rmu   sync.Mutex
	rMask cipher.Stream
	rbuf  bytes.Buffer
	rdata []byte // reused frame-read scratch (single reader goroutine)
	hdr   [2]byte
}

// NewFramedConn wraps conn. chain holds live per-connection plugin state and is
// closed by Close. maxFrame <= 0 selects DefaultMaxFrame.
func NewFramedConn(conn net.Conn, chain *Chain, isClient bool, maxFrame int) (*FramedConn, error) {
	if maxFrame <= 0 {
		maxFrame = DefaultMaxFrame
	}
	c2s, err := newMask(chain.FrameKey(), "c2s")
	if err != nil {
		return nil, err
	}
	s2c, err := newMask(chain.FrameKey(), "s2c")
	if err != nil {
		return nil, err
	}
	fc := &FramedConn{Conn: conn, chain: chain, maxFrame: maxFrame, framer: chain.Framer()}
	if isClient {
		fc.wMask, fc.rMask = c2s, s2c
	} else {
		fc.wMask, fc.rMask = s2c, c2s
	}
	return fc, nil
}

func newMask(key []byte, label string) (cipher.Stream, error) {
	sub := blake2b.Sum256(append([]byte("stunning/frame/"+label+"\x00"), key...))
	var nonce [chacha20.NonceSize]byte
	return chacha20.NewUnauthenticatedCipher(sub[:chacha20.KeySize], nonce[:])
}

// WriteChunk bounds the plaintext carried per frame. It is well under both the
// default max frame and the TLS record limit, leaving headroom for plugin
// overhead (aead tag/nonce, padding) so an encoded frame never exceeds maxFrame.
// Larger writes are split across frames; the reader reassembles a byte stream, so
// splitting is transparent.
const WriteChunk = 8192

// Write encodes p across one or more frames and reports how many plaintext bytes
// were accepted. Splitting keeps each encoded frame within maxFrame regardless of
// the caller's write size.
func (f *FramedConn) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		n := len(p)
		if n > WriteChunk {
			n = WriteChunk
		}
		if err := f.writeFrame(p[:n]); err != nil {
			return total, err
		}
		p = p[n:]
		total += n
	}
	return total, nil
}

func (f *FramedConn) writeFrame(p []byte) error {
	enc, err := f.chain.Encode(p)
	if err != nil {
		return err
	}
	if len(enc) > f.maxFrame {
		return ErrFrameTooLarge
	}
	f.wmu.Lock()
	defer f.wmu.Unlock()
	if f.framer != nil {
		wire, err := f.framer.Frame(enc)
		if err != nil {
			return err
		}
		if _, err := f.Conn.Write(wire); err != nil {
			return err
		}
		return nil
	}
	var lenbuf [2]byte
	binary.BigEndian.PutUint16(lenbuf[:], uint16(len(enc)))
	f.wMask.XORKeyStream(lenbuf[:], lenbuf[:])
	// Vectored write: header + payload in one syscall, no concatenation copy.
	if _, err := (&net.Buffers{lenbuf[:], enc}).WriteTo(f.Conn); err != nil {
		return err
	}
	return nil
}

// Read serves decoded payload bytes, reading and decoding the next frame when the
// internal buffer is empty.
func (f *FramedConn) Read(p []byte) (int, error) {
	f.rmu.Lock()
	defer f.rmu.Unlock()
	if f.rbuf.Len() == 0 {
		if err := f.readFrame(); err != nil {
			return 0, err
		}
	}
	return f.rbuf.Read(p)
}

func (f *FramedConn) readFrame() error {
	var buf []byte
	if f.framer != nil {
		raw, err := f.framer.Deframe(f.Conn)
		if err != nil {
			return err
		}
		if len(raw) > f.maxFrame {
			return fmt.Errorf("%w: %d", ErrFrameTooLarge, len(raw))
		}
		buf = raw
	} else {
		if _, err := io.ReadFull(f.Conn, f.hdr[:]); err != nil {
			return err
		}
		f.rMask.XORKeyStream(f.hdr[:], f.hdr[:])
		n := int(binary.BigEndian.Uint16(f.hdr[:]))
		if n > f.maxFrame {
			return fmt.Errorf("%w: %d", ErrFrameTooLarge, n)
		}
		// Reuse scratch across frames; only one goroutine reads, and Decode's
		// result is copied into rbuf below so the scratch is free afterwards.
		if cap(f.rdata) < n {
			f.rdata = make([]byte, n)
		}
		buf = f.rdata[:n]
		if _, err := io.ReadFull(f.Conn, buf); err != nil {
			return err
		}
	}
	dec, err := f.chain.Decode(buf)
	if err != nil {
		return err
	}
	f.rbuf.Write(dec)
	return nil
}

// Close closes the chain then the underlying connection.
func (f *FramedConn) Close() error {
	cerr := f.chain.Close()
	nerr := f.Conn.Close()
	if nerr != nil {
		return nerr
	}
	return cerr
}

// PacketTransform applies a chain to whole datagrams, where the packet boundary
// is the frame boundary (no length prefix). Suitable for connected UDP-style
// tunnels that already preserve message boundaries.
type PacketTransform struct {
	chain *Chain
}

// NewPacketTransform builds a datagram transform from a chain.
func NewPacketTransform(chain *Chain) *PacketTransform {
	return &PacketTransform{chain: chain}
}

// Encode transforms an outgoing datagram.
func (t *PacketTransform) Encode(p []byte) ([]byte, error) { return t.chain.Encode(p) }

// Decode transforms an incoming datagram.
func (t *PacketTransform) Decode(p []byte) ([]byte, error) { return t.chain.Decode(p) }

// Close releases the chain.
func (t *PacketTransform) Close() error { return t.chain.Close() }
