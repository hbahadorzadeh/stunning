// Package vpn provides platform-agnostic VPN interface abstraction
package vpn

import (
	"errors"
	"sync"
)

var ErrNoProvider = errors.New("VPN provider not set")

// VPNProvider handles platform-specific VPN setup
type VPNProvider interface {
	// Connect initiates a VPN connection
	Connect(serverAddr, protocol string) error

	// Disconnect stops the active VPN
	Disconnect() error

	// IsConnected returns connection status
	IsConnected() bool

	// GetError returns the last error
	GetError() string
}

var (
	provider     VPNProvider
	providerMu   sync.RWMutex
	providerOnce sync.Once
)

// SetProvider sets the platform-specific VPN provider (thread-safe)
func SetProvider(p VPNProvider) {
	providerOnce.Do(func() {
		providerMu.Lock()
		provider = p
		providerMu.Unlock()
	})
}

// Connect initiates a VPN connection
func Connect(serverAddr, protocol string) error {
	providerMu.RLock()
	p := provider
	providerMu.RUnlock()

	if p == nil {
		return ErrNoProvider
	}
	return p.Connect(serverAddr, protocol)
}

// Disconnect stops the active VPN
func Disconnect() error {
	providerMu.RLock()
	p := provider
	providerMu.RUnlock()

	if p == nil {
		return ErrNoProvider
	}
	return p.Disconnect()
}

// IsConnected returns connection status
func IsConnected() bool {
	providerMu.RLock()
	p := provider
	providerMu.RUnlock()

	if p == nil {
		return false
	}
	return p.IsConnected()
}

// GetError returns the last error
func GetError() string {
	providerMu.RLock()
	p := provider
	providerMu.RUnlock()

	if p == nil {
		return ""
	}
	return p.GetError()
}
