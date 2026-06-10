// Package plugin provides a composable, in-process plugin system for the
// Stunning tunnel. Plugins are stateful per-connection byte transforms that can
// be combined into a Chain to obfuscate, encrypt, compress, or otherwise shape
// tunnel traffic so it is harder for deep-packet-inspection firewalls to
// classify or block.
//
// Unlike the legacy .so loader in core/common, plugins here are compiled into
// the binary and self-register via init(), so they cross-compile to every
// supported platform (mobile, desktop, Windows) with no CGO.
package plugin

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Plugin is a stateful, per-connection transform. Encode runs in the
// client->wire direction; Decode reverses it in the wire->client direction.
// Implementations must guarantee Decode(peer.Encode(x)) == x for the same
// configuration. Encode and Decode for a single instance are each called from a
// single goroutine (read side and write side may differ), so a plugin that
// shares state across directions must guard it.
type Plugin interface {
	Encode(src []byte) ([]byte, error)
	Decode(src []byte) ([]byte, error)
	Close() error
}

// Factory builds a fresh plugin instance from parsed params. A new instance is
// created per connection so per-flow state (nonces, counters) never leaks
// between connections.
type Factory func(p Params) (Plugin, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]Factory{}
)

// Register adds a named plugin factory. It panics on duplicate or empty names,
// which surfaces wiring mistakes at startup rather than at runtime. Intended to
// be called from init().
func Register(name string, f Factory) {
	if name == "" {
		panic("plugin: empty name")
	}
	if f == nil {
		panic("plugin: nil factory for " + name)
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		panic("plugin: duplicate registration for " + name)
	}
	registry[name] = f
}

// New builds a plugin instance by name.
func New(name string, p Params) (Plugin, error) {
	registryMu.RLock()
	f, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin: unknown plugin %q", name)
	}
	return f(p)
}

// Registered returns the sorted list of registered plugin names.
func Registered() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Params holds plugin configuration parsed from a chain spec. Typed getters keep
// plugin constructors terse and consistent.
type Params map[string]string

// String returns the value for key, or def when absent/empty.
func (p Params) String(key, def string) string {
	if v, ok := p[key]; ok && v != "" {
		return v
	}
	return def
}

// Int returns the integer value for key, or def when absent or unparseable.
func (p Params) Int(key string, def int) int {
	v, ok := p[key]
	if !ok || v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// Bool returns the boolean value for key. Accepts 1/t/true/yes/on.
func (p Params) Bool(key string, def bool) bool {
	v, ok := p[key]
	if !ok || v == "" {
		return def
	}
	switch v {
	case "1", "t", "true", "yes", "on", "y":
		return true
	case "0", "f", "false", "no", "off", "n":
		return false
	}
	return def
}

// Duration returns a time.Duration parsed from key, or def on failure.
func (p Params) Duration(key string, def time.Duration) time.Duration {
	v, ok := p[key]
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

// Bytes returns the hex-decoded value for key. Returns an error if the key is
// present but not valid hex; returns def when the key is absent.
func (p Params) Bytes(key string, def []byte) ([]byte, error) {
	v, ok := p[key]
	if !ok || v == "" {
		return def, nil
	}
	b, err := hex.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("param %q: invalid hex: %w", key, err)
	}
	return b, nil
}

// Stateless adapts pure byte transforms into a Plugin. enc/dec must be inverses.
type Stateless struct {
	enc func([]byte) ([]byte, error)
	dec func([]byte) ([]byte, error)
}

// NewStateless builds a stateless plugin from an encode/decode pair.
func NewStateless(enc, dec func([]byte) ([]byte, error)) *Stateless {
	return &Stateless{enc: enc, dec: dec}
}

func (s *Stateless) Encode(src []byte) ([]byte, error) { return s.enc(src) }
func (s *Stateless) Decode(src []byte) ([]byte, error) { return s.dec(src) }
func (s *Stateless) Close() error                      { return nil }
