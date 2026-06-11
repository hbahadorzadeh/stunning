package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"time"
)

// jitter shapes traffic timing. Encode waits a uniform random delay in
// [min,max] before releasing each frame, smearing the inter-packet timing
// pattern that timing-analysis DPI relies on. Decode is a no-op. Keep the delay
// small (single-digit ms) -- it is per frame and trades throughput for cover.
//
// Params: min, max -- delay bounds (durations, e.g. min=1ms&max=8ms).
func init() {
	Register("jitter", newJitter)
	Register("bucket", newBucket)
}

type jitterPlugin struct {
	min, max time.Duration
}

func newJitter(p Params) (Plugin, error) {
	minD := p.Duration("min", 0)
	maxD := p.Duration("max", 5*time.Millisecond)
	if minD < 0 || maxD < minD {
		return nil, fmt.Errorf("jitter: invalid bounds min=%s max=%s", minD, maxD)
	}
	return &jitterPlugin{min: minD, max: maxD}, nil
}

func (j *jitterPlugin) Encode(src []byte) ([]byte, error) {
	span := j.max - j.min
	d := j.min
	if span > 0 {
		d += time.Duration(mrand.Int64N(int64(span) + 1))
	}
	if d > 0 {
		time.Sleep(d)
	}
	return src, nil
}

func (j *jitterPlugin) Decode(src []byte) ([]byte, error) { return src, nil }
func (j *jitterPlugin) Close() error                      { return nil }

// bucket normalizes frame sizes to a fixed quantum so every frame on the wire is
// one of a few sizes, defeating exact size fingerprinting (a deterministic cousin
// of pad). It pads each frame with random bytes up to the next multiple of size,
// recording the pad length in a 2-byte trailer so Decode restores the original.
//
// Params: size -- bucket quantum in bytes (default 256).
type bucketPlugin struct {
	size int
}

func newBucket(p Params) (Plugin, error) {
	size := p.Int("size", 256)
	if size < 1 || size > 0xffff {
		return nil, fmt.Errorf("bucket: size %d out of range [1,65535]", size)
	}
	return &bucketPlugin{size: size}, nil
}

func (b *bucketPlugin) Encode(src []byte) ([]byte, error) {
	base := len(src) + 2 // payload + trailer
	pad := (b.size - base%b.size) % b.size
	out := make([]byte, len(src)+pad+2)
	copy(out, src)
	if pad > 0 {
		if _, err := io.ReadFull(rand.Reader, out[len(src):len(src)+pad]); err != nil {
			return nil, fmt.Errorf("bucket: rand: %w", err)
		}
	}
	binary.LittleEndian.PutUint16(out[len(src)+pad:], uint16(pad))
	return out, nil
}

func (b *bucketPlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < 2 {
		return nil, fmt.Errorf("bucket: frame too short for trailer")
	}
	pad := int(binary.LittleEndian.Uint16(src[len(src)-2:]))
	if pad+2 > len(src) {
		return nil, fmt.Errorf("bucket: trailer claims %d pad bytes, frame has %d", pad, len(src)-2)
	}
	return src[:len(src)-2-pad], nil
}

func (b *bucketPlugin) Close() error { return nil }
