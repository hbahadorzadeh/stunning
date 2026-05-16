// +build android

package vpn

import (
	"fmt"
	"sync"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/jni"
)

// AndroidVPNProvider implements VPN setup for Android using VpnService
type AndroidVPNProvider struct {
	mu           sync.RWMutex
	connected    bool
	lastError    string
	activeConfig string
}

// NewAndroidVPNProvider creates an Android VPN provider
func NewAndroidVPNProvider() *AndroidVPNProvider {
	return &AndroidVPNProvider{}
}

// Connect initiates VPN connection on Android
func (avp *AndroidVPNProvider) Connect(serverAddr, protocol string) error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	env := jni.NewEnv()
	if env == nil {
		avp.lastError = "JNI environment not available"
		return fmt.Errorf("Android VPN error: %s", avp.lastError)
	}

	// Call Java method to start VpnService
	// This would use reflection to call the VPN service
	// For now, we set connected flag
	avp.connected = true
	avp.lastError = ""
	return nil
}

// Disconnect stops VPN connection on Android
func (avp *AndroidVPNProvider) Disconnect() error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	env := jni.NewEnv()
	if env == nil {
		avp.lastError = "JNI environment not available"
		return fmt.Errorf("Android VPN error: %s", avp.lastError)
	}

	// Call Java method to stop VpnService
	avp.connected = false
	avp.lastError = ""
	return nil
}

// IsConnected returns Android VPN status
func (avp *AndroidVPNProvider) IsConnected() bool {
	avp.mu.RLock()
	defer avp.mu.RUnlock()
	return avp.connected
}

// GetError returns the last error
func (avp *AndroidVPNProvider) GetError() string {
	avp.mu.RLock()
	defer avp.mu.RUnlock()
	return avp.lastError
}

// AddConfiguration adds a new VPN configuration to Android settings
func (avp *AndroidVPNProvider) AddConfiguration(name, server, protocol string) error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	env := jni.NewEnv()
	if env == nil {
		avp.lastError = "JNI environment not available"
		return fmt.Errorf("Android config error: %s", avp.lastError)
	}

	// Call Java method to save VPN configuration
	// This would store config in Android's VPN settings
	return nil
}

// RemoveConfiguration removes a VPN configuration from Android settings
func (avp *AndroidVPNProvider) RemoveConfiguration(name string) error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	env := jni.NewEnv()
	if env == nil {
		avp.lastError = "JNI environment not available"
		return fmt.Errorf("Android config error: %s", avp.lastError)
	}

	// Call Java method to remove VPN configuration
	return nil
}

// ActivateConfiguration activates a saved VPN configuration
func (avp *AndroidVPNProvider) ActivateConfiguration(name string) error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	env := jni.NewEnv()
	if env == nil {
		avp.lastError = "JNI environment not available"
		return fmt.Errorf("Android activation error: %s", avp.lastError)
	}

	// Call Java method to activate VPN configuration
	avp.activeConfig = name
	avp.connected = true
	avp.lastError = ""
	return nil
}

// SetTunFd receives the TUN file descriptor from Java VPN service
func (avp *AndroidVPNProvider) SetTunFd(fd int) error {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	// This would be called from the VPN service when TUN is established
	// The fd would be passed to the tunnel code for reading/writing packets
	return nil
}

func init() {
	SetProvider(NewAndroidVPNProvider())
}
