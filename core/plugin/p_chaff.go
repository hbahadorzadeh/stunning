package plugin

import (
	crand "crypto/rand"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"time"
)

// chaff injects decoy (cover) traffic to mask volume and timing patterns. It tags
// every frame with a 1-byte type — real or decoy — and, as a Chaffer, supplies
// FramedConn with independent decoy frames it emits on a timer. The peer's Decode
// drops decoy frames. Real and decoy frames are indistinguishable on the wire
// because both pass through the rest of the chain (aead/mimic).
//
// chaff must be the innermost (first) plugin in the chain so the type byte is
// covered by the outer encryption/disguise. Place it first:
//
//	chaff,aead?key=...,tls-mimic
//
// Params:
//
//	min, max  decoy payload size bounds (default 64..512)
//	interval  base delay between decoy frames (default 250ms)
//	jitter    random extra delay added to interval (default 250ms)
func init() { Register("chaff", newChaff) }

const (
	chaffReal  = 0x00
	chaffDecoy = 0x01
)

type chaffPlugin struct {
	min, max         int
	interval, jitter time.Duration
}

func newChaff(p Params) (Plugin, error) {
	c := &chaffPlugin{
		min:      p.Int("min", 64),
		max:      p.Int("max", 512),
		interval: p.Duration("interval", 250*time.Millisecond),
		jitter:   p.Duration("jitter", 250*time.Millisecond),
	}
	if c.min < 0 || c.max < c.min {
		return nil, fmt.Errorf("chaff: invalid size bounds min=%d max=%d", c.min, c.max)
	}
	if c.interval <= 0 {
		return nil, fmt.Errorf("chaff: interval must be > 0")
	}
	if c.jitter < 0 {
		return nil, fmt.Errorf("chaff: jitter must be >= 0")
	}
	return c, nil
}

func (*chaffPlugin) Encode(src []byte) ([]byte, error) {
	out := make([]byte, len(src)+1)
	out[0] = chaffReal
	copy(out[1:], src)
	return out, nil
}

func (*chaffPlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < 1 {
		return nil, fmt.Errorf("chaff: empty frame")
	}
	switch src[0] {
	case chaffReal:
		return src[1:], nil
	case chaffDecoy:
		return nil, nil // drop the decoy; FramedConn.Read fetches the next frame
	default:
		return nil, fmt.Errorf("chaff: bad frame type 0x%02x", src[0])
	}
}

func (*chaffPlugin) Close() error { return nil }

// ChaffPayload returns a fresh decoy inner payload (type byte + random bytes).
func (c *chaffPlugin) ChaffPayload() []byte {
	span := c.max - c.min + 1
	n := c.min
	if span > 1 {
		n += mrand.IntN(span)
	}
	out := make([]byte, n+1)
	out[0] = chaffDecoy
	if _, err := io.ReadFull(crand.Reader, out[1:]); err != nil {
		return []byte{chaffDecoy} // still a valid (minimal) decoy
	}
	return out
}

// NextDelay returns the delay until the next decoy frame.
func (c *chaffPlugin) NextDelay() time.Duration {
	d := c.interval
	if c.jitter > 0 {
		d += time.Duration(mrand.Int64N(int64(c.jitter) + 1))
	}
	return d
}
