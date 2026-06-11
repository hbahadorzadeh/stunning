package plugin

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Authenticator is a connection gate: it runs a handshake right after the
// connection is established (and, when chained with plugins, inside the chain
// framing so the handshake is also disguised). Unlike a Plugin it does not
// transform application bytes -- it proves identity and either lets the
// connection proceed or rejects it.
//
// Handshake methods return the connection to use for subsequent data. Most
// authenticators return the same conn; some (mtls) wrap it in a TLS session.
type Authenticator interface {
	// ClientHandshake authenticates from the client side.
	ClientHandshake(conn net.Conn) (net.Conn, error)
	// ServerHandshake authenticates an accepted connection, returning the conn to
	// continue with and the authenticated identity, or an error to reject.
	ServerHandshake(conn net.Conn) (net.Conn, string, error)
}

// AuthFactory builds an authenticator from parsed params.
type AuthFactory func(p Params) (Authenticator, error)

var (
	authMu       sync.RWMutex
	authRegistry = map[string]AuthFactory{}
)

// RegisterAuth registers a named authenticator factory (called from init()).
func RegisterAuth(name string, f AuthFactory) {
	if name == "" || f == nil {
		panic("plugin: invalid auth registration")
	}
	authMu.Lock()
	defer authMu.Unlock()
	if _, dup := authRegistry[name]; dup {
		panic("plugin: duplicate auth registration for " + name)
	}
	authRegistry[name] = f
}

// NewAuth builds an authenticator by name.
func NewAuth(name string, p Params) (Authenticator, error) {
	authMu.RLock()
	f, ok := authRegistry[name]
	authMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin: unknown authenticator %q", name)
	}
	return f(p)
}

// ParseAuth builds the single authenticator described by spec ("name?k=v&..."),
// or nil for an empty spec.
func ParseAuth(spec string) (Authenticator, error) {
	if spec == "" {
		return nil, nil
	}
	name, params, err := parseEntry(spec)
	if err != nil {
		return nil, err
	}
	return NewAuth(name, params)
}

// HandshakeTimeout bounds each auth handshake.
const HandshakeTimeout = 15 * time.Second

const maxHandshakeMsg = 64 * 1024

// writeMsg writes a length-prefixed message (bounded) to conn.
func writeMsg(conn net.Conn, b []byte) error {
	if len(b) > maxHandshakeMsg {
		return fmt.Errorf("handshake message too large: %d", len(b))
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(b)))
	if _, err := conn.Write(hdr[:]); err != nil {
		return err
	}
	if len(b) > 0 {
		if _, err := conn.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// readMsg reads a length-prefixed message, rejecting oversized frames.
func readMsg(conn net.Conn) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > maxHandshakeMsg {
		return nil, fmt.Errorf("handshake message too large: %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
