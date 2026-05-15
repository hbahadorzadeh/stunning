// Package vpn provides platform-agnostic VPN interface abstraction
package vpn

import "errors"

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

var provider VPNProvider

// SetProvider sets the platform-specific VPN provider
func SetProvider(p VPNProvider) {
	provider = p
}

// Connect initiates a VPN connection
func Connect(serverAddr, protocol string) error {
	if provider == nil {
		return ErrNoProvider
	}
	return provider.Connect(serverAddr, protocol)
}

// Disconnect stops the active VPN
func Disconnect() error {
	if provider == nil {
		return ErrNoProvider
	}
	return provider.Disconnect()
}

// IsConnected returns connection status
func IsConnected() bool {
	if provider == nil {
		return false
	}
	return provider.IsConnected()
}

// GetError returns the last error
func GetError() string {
	if provider == nil {
		return ""
	}
	return provider.GetError()
}
