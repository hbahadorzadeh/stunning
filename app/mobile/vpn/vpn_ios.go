//go:build ios
// +build ios

package vpn

import (
	"fmt"
	"sync"
)

// iOSVPNProvider implements VPN setup for iOS using NetworkExtension
type iOSVPNProvider struct {
	mu           sync.RWMutex
	connected    bool
	lastError    string
	activeConfig string
}

// NewiOSVPNProvider creates an iOS VPN provider
func NewiOSVPNProvider() *iOSVPNProvider {
	return &iOSVPNProvider{}
}

// Connect initiates VPN connection on iOS
func (ivp *iOSVPNProvider) Connect(serverAddr, protocol string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	// TODO: Call Swift/Objective-C to start NEPacketTunnelProvider
	// For now, simulate the connection
	ivp.connected = true
	ivp.lastError = ""
	return nil
}

// Disconnect stops VPN connection on iOS
func (ivp *iOSVPNProvider) Disconnect() error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	// TODO: Call Swift/Objective-C to stop NEPacketTunnelProvider
	// For now, simulate the disconnection
	ivp.connected = false
	ivp.lastError = ""
	return nil
}

// IsConnected returns iOS VPN status
func (ivp *iOSVPNProvider) IsConnected() bool {
	ivp.mu.RLock()
	defer ivp.mu.RUnlock()
	return ivp.connected
}

// GetError returns the last error
func (ivp *iOSVPNProvider) GetError() string {
	ivp.mu.RLock()
	defer ivp.mu.RUnlock()
	return ivp.lastError
}

// AddConfiguration adds a new VPN configuration to iOS settings
func (ivp *iOSVPNProvider) AddConfiguration(name, server, protocol string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	// TODO: Call Swift/Objective-C to add NetworkExtension configuration
	// For now, just validate inputs
	if name == "" || server == "" || protocol == "" {
		ivp.lastError = "invalid configuration parameters"
		return fmt.Errorf("iOS config error: %s", ivp.lastError)
	}

	return nil
}

// RemoveConfiguration removes a VPN configuration from iOS settings
func (ivp *iOSVPNProvider) RemoveConfiguration(name string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	// TODO: Call Swift/Objective-C to remove NetworkExtension configuration
	if name == "" {
		ivp.lastError = "invalid configuration name"
		return fmt.Errorf("iOS config error: %s", ivp.lastError)
	}

	return nil
}

// ActivateConfiguration activates a saved VPN configuration
func (ivp *iOSVPNProvider) ActivateConfiguration(name string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	// TODO: Call Swift/Objective-C to activate NetworkExtension configuration
	if name == "" {
		ivp.lastError = "invalid configuration name"
		return fmt.Errorf("iOS activation error: %s", ivp.lastError)
	}

	ivp.activeConfig = name
	ivp.connected = true
	ivp.lastError = ""
	return nil
}

func init() {
	SetProvider(NewiOSVPNProvider())
}
