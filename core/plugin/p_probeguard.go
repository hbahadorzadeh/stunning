package plugin

import (
	"crypto/subtle"
	"fmt"
	"hash"

	"golang.org/x/crypto/blake2b"
)

// probeguard resists active probing. It prefixes each frame with a keyed BLAKE2b
// tag over the payload. On Decode a frame whose tag does not verify is rejected
// with an error, so the server drops unauthenticated bytes (e.g. a censor's
// probe) and returns nothing distinguishing. Place it outermost (last in Encode)
// so the tag covers the whole encoded frame and is checked before any expensive
// inner decode work.
//
// The plugin is per connection and Encode/Decode run on separate (write/read)
// goroutines, so each direction gets its own reusable hasher -- no allocation or
// pool overhead per frame, and no shared state.
//
// Params:
//
//	key    hex secret (required).
//	taglen tag bytes, 8..32 (default 16).
func init() { Register("probe-guard", newProbeGuard) }

type probeGuardPlugin struct {
	taglen int
	encH   hash.Hash
	decH   hash.Hash
}

func newProbeGuard(p Params) (Plugin, error) {
	key, err := p.Bytes("key", nil)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("probe-guard: missing required param key")
	}
	taglen := p.Int("taglen", 16)
	if taglen < 8 || taglen > 32 {
		return nil, fmt.Errorf("probe-guard: taglen %d out of range [8,32]", taglen)
	}
	// blake2b keyed hashing requires key <= 64 bytes; normalize.
	k := blake2b.Sum512(key)
	encH, err := blake2b.New(taglen, k[:])
	if err != nil {
		return nil, err
	}
	decH, err := blake2b.New(taglen, k[:])
	if err != nil {
		return nil, err
	}
	return &probeGuardPlugin{taglen: taglen, encH: encH, decH: decH}, nil
}

func (g *probeGuardPlugin) Encode(src []byte) ([]byte, error) {
	g.encH.Reset()
	g.encH.Write(src)
	tag := g.encH.Sum(nil)
	out := make([]byte, 0, len(tag)+len(src))
	out = append(out, tag...)
	out = append(out, src...)
	return out, nil
}

func (g *probeGuardPlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < g.taglen {
		return nil, fmt.Errorf("probe-guard: frame shorter than tag")
	}
	got, payload := src[:g.taglen], src[g.taglen:]
	g.decH.Reset()
	g.decH.Write(payload)
	want := g.decH.Sum(nil)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return nil, fmt.Errorf("probe-guard: authentication failed")
	}
	return payload, nil
}

func (*probeGuardPlugin) Close() error { return nil }
