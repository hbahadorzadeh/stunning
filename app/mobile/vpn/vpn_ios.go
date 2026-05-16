// +build ios

package vpn

import (
	"fmt"
	"sync"
)

/*
#cgo CFLAGS: -fmodules -fblocks
#cgo LDFLAGS: -framework NetworkExtension -framework Foundation

#import <Foundation/Foundation.h>

// Declare Swift functions
void iOSStartVPN(const char *server, const char *protocol);
void iOSStopVPN(void);
BOOL iOSIsVPNActive(void);
const char *iOSGetLastError(void);
void iOSAddVPNConfiguration(const char *name, const char *server, const char *protocol);
void iOSRemoveVPNConfiguration(const char *name);
BOOL iOSActivateVPNConfiguration(const char *name);
*/
import "C"

import "unsafe"

// iOSVPNProvider implements VPN setup for iOS using NetworkExtension
type iOSVPNProvider struct {
	mu        sync.RWMutex
	connected bool
	lastError string
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

	cServer := C.CString(serverAddr)
	cProto := C.CString(protocol)
	defer C.free(unsafe.Pointer(cServer))
	defer C.free(unsafe.Pointer(cProto))

	// Start VPN through NetworkExtension
	C.iOSStartVPN(cServer, cProto)

	// Check for errors from the C call
	errStr := C.GoString(C.iOSGetLastError())
	if errStr != "" {
		ivp.lastError = errStr
		return fmt.Errorf("iOS VPN error: %s", errStr)
	}

	ivp.connected = true
	ivp.lastError = ""
	return nil
}

// Disconnect stops VPN connection on iOS
func (ivp *iOSVPNProvider) Disconnect() error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	C.iOSStopVPN()

	errStr := C.GoString(C.iOSGetLastError())
	if errStr != "" {
		ivp.lastError = errStr
		return fmt.Errorf("iOS VPN error: %s", errStr)
	}

	ivp.connected = false
	ivp.lastError = ""
	return nil
}

// IsConnected returns iOS VPN status
func (ivp *iOSVPNProvider) IsConnected() bool {
	ivp.mu.RLock()
	defer ivp.mu.RUnlock()
	return C.iOSIsVPNActive() == C.BOOL(true)
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

	cName := C.CString(name)
	cServer := C.CString(server)
	cProto := C.CString(protocol)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cServer))
	defer C.free(unsafe.Pointer(cProto))

	C.iOSAddVPNConfiguration(cName, cServer, cProto)

	errStr := C.GoString(C.iOSGetLastError())
	if errStr != "" {
		ivp.lastError = errStr
		return fmt.Errorf("iOS config error: %s", errStr)
	}

	return nil
}

// RemoveConfiguration removes a VPN configuration from iOS settings
func (ivp *iOSVPNProvider) RemoveConfiguration(name string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.iOSRemoveVPNConfiguration(cName)

	errStr := C.GoString(C.iOSGetLastError())
	if errStr != "" {
		ivp.lastError = errStr
		return fmt.Errorf("iOS config error: %s", errStr)
	}

	return nil
}

// ActivateConfiguration activates a saved VPN configuration
func (ivp *iOSVPNProvider) ActivateConfiguration(name string) error {
	ivp.mu.Lock()
	defer ivp.mu.Unlock()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if C.iOSActivateVPNConfiguration(cName) == C.BOOL(false) {
		errStr := C.GoString(C.iOSGetLastError())
		if errStr != "" {
			ivp.lastError = errStr
			return fmt.Errorf("iOS activation error: %s", errStr)
		}
		return fmt.Errorf("failed to activate VPN configuration")
	}

	ivp.activeConfig = name
	ivp.connected = true
	ivp.lastError = ""
	return nil
}

func init() {
	SetProvider(NewiOSVPNProvider())
}
