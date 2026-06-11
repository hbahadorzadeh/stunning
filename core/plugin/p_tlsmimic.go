package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// tls-mimic disguises the tunnel as TLS. As a Framer it puts a synthetic TLS
// handshake record first (so a DPI box sees a ClientHello/ServerHello and
// allowlists the flow), then carries each payload inside TLS application_data
// records (type 0x17), whose length fields provide the framing. Its Encode/Decode
// are identity -- the disguise is entirely in the framing, so it pairs with an
// inner aead/flate chain that does the real work.
//
// Place it last in the chain (outermost on the wire):
//
//	flate,aead?key=...,tls-mimic
func init() { Register("tls-mimic", newTLSMimic) }

const (
	tlsHandshake   = 0x16
	tlsAppData     = 0x17
	tlsVerMajor    = 0x03
	tlsVerMinor    = 0x03 // TLS 1.2 on the wire, like most real 1.3 handshakes
	tlsMaxRecord   = 1 << 14
	clientHelloMsg = 0x01
)

type tlsMimic struct {
	wroteHandshake bool
	readHandshake  bool
}

func newTLSMimic(_ Params) (Plugin, error) { return &tlsMimic{}, nil }

func (*tlsMimic) Encode(src []byte) ([]byte, error) { return src, nil }
func (*tlsMimic) Decode(src []byte) ([]byte, error) { return src, nil }
func (*tlsMimic) Close() error                      { return nil }

func tlsRecord(typ byte, body []byte) []byte {
	out := make([]byte, 5+len(body))
	out[0] = typ
	out[1] = tlsVerMajor
	out[2] = tlsVerMinor
	binary.BigEndian.PutUint16(out[3:5], uint16(len(body)))
	copy(out[5:], body)
	return out
}

// syntheticHandshake builds a plausible-looking ClientHello body: handshake type
// 0x01, a 3-byte length, the legacy version, 32 random bytes, and a random
// session id. A DPI box keys on the leading 0x16 0x03 ... 0x01, which this
// satisfies; the remaining bytes only need to look random.
func syntheticHandshake() ([]byte, error) {
	body := make([]byte, 4+2+32+1+32)
	body[0] = clientHelloMsg
	// 3-byte handshake length (excludes the 4-byte header).
	msgLen := len(body) - 4
	body[1] = byte(msgLen >> 16)
	body[2] = byte(msgLen >> 8)
	body[3] = byte(msgLen)
	body[4] = tlsVerMajor
	body[5] = tlsVerMinor
	if _, err := io.ReadFull(rand.Reader, body[6:38]); err != nil { // 32 random
		return nil, err
	}
	body[38] = 32 // session id length
	if _, err := io.ReadFull(rand.Reader, body[39:]); err != nil {
		return nil, err
	}
	return body, nil
}

func (t *tlsMimic) Frame(payload []byte) ([]byte, error) {
	if len(payload) > tlsMaxRecord {
		return nil, fmt.Errorf("tls-mimic: payload %d exceeds record max", len(payload))
	}
	app := tlsRecord(tlsAppData, payload)
	if t.wroteHandshake {
		return app, nil
	}
	t.wroteHandshake = true
	hs, err := syntheticHandshake()
	if err != nil {
		return nil, err
	}
	return append(tlsRecord(tlsHandshake, hs), app...), nil
}

func (t *tlsMimic) Deframe(r io.Reader) ([]byte, error) {
	for {
		var hdr [5]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			return nil, err
		}
		if hdr[1] != tlsVerMajor {
			return nil, fmt.Errorf("tls-mimic: bad record version 0x%02x", hdr[1])
		}
		n := int(binary.BigEndian.Uint16(hdr[3:5]))
		if n > tlsMaxRecord {
			return nil, fmt.Errorf("tls-mimic: record length %d too large", n)
		}
		body := make([]byte, n)
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, err
		}
		switch hdr[0] {
		case tlsHandshake:
			t.readHandshake = true
			continue // synthetic, discard
		case tlsAppData:
			return body, nil
		default:
			return nil, fmt.Errorf("tls-mimic: unexpected record type 0x%02x", hdr[0])
		}
	}
}
