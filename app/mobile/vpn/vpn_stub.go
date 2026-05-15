// +build !android,!ios

package vpn

// StubProvider is a no-op VPN provider for non-mobile platforms
type StubProvider struct {
	connected bool
	lastError string
}

// NewStubProvider creates a stub VPN provider
func NewStubProvider() *StubProvider {
	return &StubProvider{}
}

// Connect simulates VPN connection
func (sp *StubProvider) Connect(serverAddr, protocol string) error {
	sp.connected = true
	sp.lastError = ""
	return nil
}

// Disconnect simulates VPN disconnection
func (sp *StubProvider) Disconnect() error {
	sp.connected = false
	sp.lastError = ""
	return nil
}

// IsConnected returns stub connection status
func (sp *StubProvider) IsConnected() bool {
	return sp.connected
}

// GetError returns the last error
func (sp *StubProvider) GetError() string {
	return sp.lastError
}

func init() {
	SetProvider(NewStubProvider())
}
