package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// pad defeats packet/frame-size fingerprinting by appending a random amount of
// random padding to each frame. The layout is:
//
//	payload || padbytes || uint16(len(padbytes))
//
// The 2-byte little-endian trailer lets Decode strip the padding exactly, so a
// single instance is its own inverse even though the pad length varies per call.
// Place pad after aead so the padding falls on ciphertext and is itself
// indistinguishable from the encrypted body.
//
// Params: min, max — inclusive byte bounds for the random pad length.
func init() { Register("pad", newPad) }

type padPlugin struct {
	min, max int
}

func newPad(p Params) (Plugin, error) {
	minN := p.Int("min", 0)
	maxN := p.Int("max", 256)
	if minN < 0 || maxN < minN {
		return nil, fmt.Errorf("pad: invalid bounds min=%d max=%d", minN, maxN)
	}
	if maxN > 0xffff {
		return nil, fmt.Errorf("pad: max %d exceeds uint16 trailer", maxN)
	}
	return &padPlugin{min: minN, max: maxN}, nil
}

func (pp *padPlugin) Encode(src []byte) ([]byte, error) {
	span := pp.max - pp.min + 1
	n := pp.min
	if span > 1 {
		var b [2]byte
		if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
			return nil, fmt.Errorf("pad: rand: %w", err)
		}
		n = pp.min + int(binary.LittleEndian.Uint16(b[:]))%span
	}
	out := make([]byte, len(src)+n+2)
	copy(out, src)
	if n > 0 {
		if _, err := io.ReadFull(rand.Reader, out[len(src):len(src)+n]); err != nil {
			return nil, fmt.Errorf("pad: rand fill: %w", err)
		}
	}
	binary.LittleEndian.PutUint16(out[len(src)+n:], uint16(n))
	return out, nil
}

func (*padPlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < 2 {
		return nil, fmt.Errorf("pad: frame too short for trailer")
	}
	n := int(binary.LittleEndian.Uint16(src[len(src)-2:]))
	if n+2 > len(src) {
		return nil, fmt.Errorf("pad: trailer claims %d pad bytes, frame has %d", n, len(src)-2)
	}
	return src[:len(src)-2-n], nil
}

func (*padPlugin) Close() error { return nil }
