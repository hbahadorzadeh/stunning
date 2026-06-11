package plugin

import (
	"crypto/subtle"
	"fmt"
	"hash"
	"sync"

	"golang.org/x/crypto/blake2b"
)

// probeguard resists active probing. It prefixes each frame with a keyed BLAKE2b
// tag over the payload. On Decode a frame whose tag does not verify is rejected
// with an error, so the server drops unauthenticated bytes (e.g. a censor's
// probe) and returns nothing distinguishing. Place it outermost (last in Encode)
// so the tag covers the whole encoded frame and is checked before any expensive
// inner decode work.
//
// Params:
//
//	key    hex secret (required).
//	taglen tag bytes, 8..32 (default 16).
func init() { Register("probe-guard", newProbeGuard) }

type probeGuardPlugin struct {
	key    []byte
	taglen int
	hpool  sync.Pool // hash.Hash keyed BLAKE2b instances
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
	return &probeGuardPlugin{key: k[:], taglen: taglen}, nil
}

func (g *probeGuardPlugin) tag(payload []byte) ([]byte, error) {
	h, _ := g.hpool.Get().(hash.Hash)
	if h == nil {
		var err error
		if h, err = blake2b.New(g.taglen, g.key); err != nil {
			return nil, err
		}
	} else {
		h.Reset()
	}
	h.Write(payload)
	sum := h.Sum(nil)
	g.hpool.Put(h)
	return sum, nil
}

func (g *probeGuardPlugin) Encode(src []byte) ([]byte, error) {
	tag, err := g.tag(src)
	if err != nil {
		return nil, err
	}
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
	want, err := g.tag(payload)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return nil, fmt.Errorf("probe-guard: authentication failed")
	}
	return payload, nil
}

func (g *probeGuardPlugin) Close() error { return nil }
