package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"time"
)

// profile shapes traffic to resemble a named real-world protocol so statistical
// DPI classifiers (which key on packet-size and inter-packet-timing
// distributions) see a familiar shape rather than an unknown tunnel. Each frame
// is quantized up to a multiple of the profile's size quantum (self-describing
// trailer, like bucket) and released after a sampled inter-frame delay drawn from
// the profile's timing range.
//
// Note: a plugin can only pad up and delay, not split frames, so size-mimicry is
// quantization rather than exact per-packet shaping. Combine with a small write
// chunk for tighter matching.
//
// Params:
//
//	name   web | video | voip | custom (default web)
//	quantum, min, max  override the preset (required for name=custom)
func init() { Register("profile", newProfile) }

type profilePreset struct {
	quantum  int
	min, max time.Duration
}

var profilePresets = map[string]profilePreset{
	"web":   {quantum: 256, min: 1 * time.Millisecond, max: 30 * time.Millisecond},
	"video": {quantum: 1024, min: 5 * time.Millisecond, max: 40 * time.Millisecond},
	"voip":  {quantum: 160, min: 15 * time.Millisecond, max: 25 * time.Millisecond},
}

type profilePlugin struct {
	quantum  int
	min, max time.Duration
}

func newProfile(p Params) (Plugin, error) {
	name := p.String("name", "web")
	preset, ok := profilePresets[name]
	if !ok && name != "custom" {
		return nil, fmt.Errorf("profile: unknown name %q", name)
	}
	pp := &profilePlugin{quantum: preset.quantum, min: preset.min, max: preset.max}
	if name == "custom" {
		pp.quantum = p.Int("quantum", 256)
		pp.min = p.Duration("min", 0)
		pp.max = p.Duration("max", 10*time.Millisecond)
	} else {
		// allow per-field overrides on a preset
		pp.quantum = p.Int("quantum", pp.quantum)
		pp.min = p.Duration("min", pp.min)
		pp.max = p.Duration("max", pp.max)
	}
	if pp.quantum < 1 || pp.quantum > 0xffff {
		return nil, fmt.Errorf("profile: quantum %d out of range [1,65535]", pp.quantum)
	}
	if pp.min < 0 || pp.max < pp.min {
		return nil, fmt.Errorf("profile: invalid delay bounds min=%s max=%s", pp.min, pp.max)
	}
	return pp, nil
}

func (pp *profilePlugin) Encode(src []byte) ([]byte, error) {
	if span := pp.max - pp.min; span >= 0 && pp.max > 0 {
		d := pp.min + time.Duration(mrand.Int64N(int64(span)+1))
		if d > 0 {
			time.Sleep(d)
		}
	}
	base := len(src) + 2
	pad := (pp.quantum - base%pp.quantum) % pp.quantum
	out := make([]byte, len(src)+pad+2)
	copy(out, src)
	if pad > 0 {
		if _, err := io.ReadFull(rand.Reader, out[len(src):len(src)+pad]); err != nil {
			return nil, fmt.Errorf("profile: rand: %w", err)
		}
	}
	binary.LittleEndian.PutUint16(out[len(src)+pad:], uint16(pad))
	return out, nil
}

func (*profilePlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < 2 {
		return nil, fmt.Errorf("profile: frame too short for trailer")
	}
	pad := int(binary.LittleEndian.Uint16(src[len(src)-2:]))
	if pad+2 > len(src) {
		return nil, fmt.Errorf("profile: trailer claims %d pad bytes, frame has %d", pad, len(src)-2)
	}
	return src[:len(src)-2-pad], nil
}

func (*profilePlugin) Close() error { return nil }
