// +build ios

package vpn

// iOSVPNProvider implements VPN setup for iOS
type iOSVPNProvider struct {
	connected bool
	lastError string
}

// NewiOSVPNProvider creates an iOS VPN provider
func NewiOSVPNProvider() *iOSVPNProvider {
	return &iOSVPNProvider{}
}

// Connect initiates VPN connection on iOS
func (ivp *iOSVPNProvider) Connect(serverAddr, protocol string) error {
	// On iOS, this would:
	// 1. Call Swift to start NEPacketTunnelProvider
	// 2. Wait for VPN to be established
	// 3. Configure NetworkExtension settings
	ivp.connected = true
	ivp.lastError = ""
	return nil
}

// Disconnect stops VPN connection on iOS
func (ivp *iOSVPNProvider) Disconnect() error {
	// On iOS, this would:
	// 1. Call Swift to stop NEPacketTunnelProvider
	// 2. Clean up NetworkExtension settings
	ivp.connected = false
	ivp.lastError = ""
	return nil
}

// IsConnected returns iOS VPN status
func (ivp *iOSVPNProvider) IsConnected() bool {
	return ivp.connected
}

// GetError returns the last error
func (ivp *iOSVPNProvider) GetError() string {
	return ivp.lastError
}

// iOS-specific helper functions would go here:
// SetPacketHandler(handler func([]byte)) - called from Swift tunnel provider

func init() {
	SetProvider(NewiOSVPNProvider())
}
