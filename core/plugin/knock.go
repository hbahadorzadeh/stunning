package plugin

import (
	"fmt"
	"sync"
)

// Knocker is a pre-connection authorization gate (port knocking). The client
// sends an out-of-band authorization (Knock) before dialing the tunnel; the
// server runs a knock listener (Start) and only allows tunnel connections from
// recently-authorized source IPs (Authorized). It does not transform any bytes.
type Knocker interface {
	// Knock sends client-side authorization for the given server host.
	Knock(host string) error
	// Start begins the server-side knock listener. Idempotent.
	Start() error
	// Authorized reports whether the source IP is currently allowed.
	Authorized(ip string) bool
	// Close stops the listener.
	Close() error
}

// KnockFactory builds a knocker from parsed params.
type KnockFactory func(p Params) (Knocker, error)

var (
	knockMu       sync.RWMutex
	knockRegistry = map[string]KnockFactory{}
)

// RegisterKnock registers a named knock factory (called from init()).
func RegisterKnock(name string, f KnockFactory) {
	if name == "" || f == nil {
		panic("plugin: invalid knock registration")
	}
	knockMu.Lock()
	defer knockMu.Unlock()
	if _, dup := knockRegistry[name]; dup {
		panic("plugin: duplicate knock registration for " + name)
	}
	knockRegistry[name] = f
}

// NewKnock builds a knocker by name.
func NewKnock(name string, p Params) (Knocker, error) {
	knockMu.RLock()
	f, ok := knockRegistry[name]
	knockMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin: unknown knocker %q", name)
	}
	return f(p)
}

// ParseKnock builds the knocker described by spec ("name?k=v&..."), or nil for an
// empty spec.
func ParseKnock(spec string) (Knocker, error) {
	if spec == "" {
		return nil, nil
	}
	name, params, err := parseEntry(spec)
	if err != nil {
		return nil, err
	}
	return NewKnock(name, params)
}
