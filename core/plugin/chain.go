package plugin

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Chain is an ordered set of plugin instances applied to tunnel payloads.
// Encode runs the plugins front-to-back; Decode runs them back-to-front, so a
// Chain built from the same spec on both peers is a mutual inverse:
//
//	server.Decode(client.Encode(x)) == x
//
// A Chain holds live per-connection plugin state, so build one Chain per
// connection (call ParseChain again) rather than sharing across connections.
type Chain struct {
	plugins []Plugin
	spec    string
}

// ParseChain builds a Chain from a spec string of the form:
//
//	name1,name2?k=v&k2=v2,name3?k=v
//
// An empty spec yields an empty (pass-through) chain. Each named plugin is
// instantiated immediately, so any constructor error (bad key, unknown plugin)
// is reported here.
func ParseChain(spec string) (*Chain, error) {
	c := &Chain{spec: spec}
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return c, nil
	}
	for _, entry := range strings.Split(spec, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		name, params, err := parseEntry(entry)
		if err != nil {
			c.Close()
			return nil, err
		}
		p, err := New(name, params)
		if err != nil {
			c.Close()
			return nil, err
		}
		c.plugins = append(c.plugins, p)
	}
	return c, nil
}

func parseEntry(entry string) (string, Params, error) {
	name := entry
	params := Params{}
	if i := strings.IndexByte(entry, '?'); i >= 0 {
		name = entry[:i]
		for _, kv := range strings.Split(entry[i+1:], "&") {
			if kv == "" {
				continue
			}
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return "", nil, fmt.Errorf("plugin: malformed param %q in %q", kv, entry)
			}
			params[k] = v
		}
	}
	if name == "" {
		return "", nil, fmt.Errorf("plugin: empty plugin name in %q", entry)
	}
	return name, params, nil
}

// Encode applies every plugin in order. The returned slice may alias plugin
// internal buffers; callers that retain it across calls should copy.
func (c *Chain) Encode(src []byte) ([]byte, error) {
	var err error
	for _, p := range c.plugins {
		src, err = p.Encode(src)
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

// Decode applies every plugin in reverse order.
func (c *Chain) Decode(src []byte) ([]byte, error) {
	var err error
	for i := len(c.plugins) - 1; i >= 0; i-- {
		src, err = c.plugins[i].Decode(src)
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

// Len reports the number of plugins in the chain.
func (c *Chain) Len() int { return len(c.plugins) }

// Chaffer is an optional capability of the innermost (first) plugin: it lets
// FramedConn inject independent decoy frames. The chaff plugin tags real vs decoy
// frames; ChaffPayload returns a fresh, already-tagged decoy inner payload and
// NextDelay paces injection.
type Chaffer interface {
	ChaffPayload() []byte
	NextDelay() time.Duration
}

// Chaffer returns the chaff capability if the innermost plugin provides one.
func (c *Chain) Chaffer() Chaffer {
	if len(c.plugins) == 0 {
		return nil
	}
	if ch, ok := c.plugins[0].(Chaffer); ok {
		return ch
	}
	return nil
}

// EncodeChaff builds one decoy frame: it takes a tagged chaff payload from the
// innermost chaff plugin and runs it through the remaining (outer) plugins, so
// the decoy is encrypted/disguised exactly like real traffic. The peer's Decode
// reconstructs it and the chaff plugin drops it (returns empty).
func (c *Chain) EncodeChaff() ([]byte, error) {
	ch := c.Chaffer()
	if ch == nil {
		return nil, fmt.Errorf("plugin: chain has no chaffer")
	}
	p := ch.ChaffPayload()
	var err error
	for i := 1; i < len(c.plugins); i++ {
		if p, err = c.plugins[i].Encode(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Framer returns the outermost plugin's protocol framing if it provides one.
// FramedConn uses it in place of the default masked-length framing so the wire
// carries a convincing protocol header. Only the last (outermost) plugin is
// consulted; an inner Framer would be hidden behind later transforms.
func (c *Chain) Framer() Framer {
	if len(c.plugins) == 0 {
		return nil
	}
	if fr, ok := c.plugins[len(c.plugins)-1].(Framer); ok {
		return fr
	}
	return nil
}

// Spec returns the original spec string.
func (c *Chain) Spec() string { return c.spec }

// FrameKey derives a deterministic 32-byte key from the spec, shared by both
// peers (identical spec -> identical key). It seeds the frame-length mask. When
// the spec embeds a secret (an aead key), the derived key is secret too; with a
// public spec it provides obfuscation only, which is all framing needs.
func (c *Chain) FrameKey() []byte {
	sum := blake2b.Sum256([]byte("stunning/frame-mask\x00" + c.spec))
	return sum[:]
}

// Close releases every plugin's resources.
func (c *Chain) Close() error {
	var first error
	for _, p := range c.plugins {
		if err := p.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
