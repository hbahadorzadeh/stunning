// +build android

package vpn

// AndroidVPNProvider implements VPN setup for Android
type AndroidVPNProvider struct {
	connected bool
	lastError string
}

// NewAndroidVPNProvider creates an Android VPN provider
func NewAndroidVPNProvider() *AndroidVPNProvider {
	return &AndroidVPNProvider{}
}

// Connect initiates VPN connection on Android
func (avp *AndroidVPNProvider) Connect(serverAddr, protocol string) error {
	// On Android, this would:
	// 1. Call JNI to start VpnService
	// 2. Pass TUN fd to Go tunnel code via bindings.SetTunFd()
	// 3. Wait for VPN to be established
	avp.connected = true
	avp.lastError = ""
	return nil
}

// Disconnect stops VPN connection on Android
func (avp *AndroidVPNProvider) Disconnect() error {
	// On Android, this would:
	// 1. Call JNI to stop VpnService
	// 2. Close the TUN fd
	avp.connected = false
	avp.lastError = ""
	return nil
}

// IsConnected returns Android VPN status
func (avp *AndroidVPNProvider) IsConnected() bool {
	return avp.connected
}

// GetError returns the last error
func (avp *AndroidVPNProvider) GetError() string {
	return avp.lastError
}

// Android-specific helper functions would go here:
// SetTunFd(fd int) - called from Java VPN service with TUN fd
// GetAppContext() - returns Android app context

func init() {
	SetProvider(NewAndroidVPNProvider())
}
