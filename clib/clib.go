package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* name;
    int   running;
    long long rx_bytes;
    long long tx_bytes;
    char* error;
} TunnelStatus;

static inline void cgo_free(void *ptr) {
    free(ptr);
}
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"

	"github.com/hbahadorzadeh/stunning/core"
)

// Global tunnel storage with thread-safe access
var (
	tunnels   map[string]core.Tunnel
	tunnelsMu sync.RWMutex
)

func init() {
	tunnels = make(map[string]core.Tunnel)
}

// JSON response structures
type JSONResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type TunnelStatusJSON struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
	Error   string `json:"error,omitempty"`
}

// StartTunnelJSON starts a tunnel from JSON config string
// NOTE: The returned C string must be freed by the caller using FreeString()
//
//export StartTunnelJSON
func StartTunnelJSON(name *C.char, configJSON *C.char) (result *C.char) {
	defer func() {
		if r := recover(); r != nil {
			resp := JSONResponse{
				Success: false,
				Error:   fmt.Sprintf("tunnel factory panic: %v", r),
			}
			data, _ := json.Marshal(resp)
			result = C.CString(string(data))
		}
	}()

	if name == nil || configJSON == nil {
		resp := JSONResponse{
			Success: false,
			Error:   "name or configJSON is nil",
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}

	nameStr := C.GoString(name)
	configStr := C.GoString(configJSON)

	// Check if tunnel already exists
	tunnelsMu.RLock()
	if _, exists := tunnels[nameStr]; exists {
		tunnelsMu.RUnlock()
		resp := JSONResponse{
			Success: false,
			Error:   fmt.Sprintf("tunnel %s already exists", nameStr),
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}
	tunnelsMu.RUnlock()

	// Parse JSON config
	var config core.TunnelConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		resp := JSONResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON config: %v", err),
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}

	tunnel := core.TunnelFactory(nameStr, config)

	// Store tunnel
	tunnelsMu.Lock()
	tunnels[nameStr] = tunnel
	tunnelsMu.Unlock()

	// Start tunnel in a goroutine
	go tunnel.ListenAndServer()

	resp := JSONResponse{
		Success: true,
		Data: map[string]string{
			"name": nameStr,
		},
	}
	data, _ := json.Marshal(resp)
	return C.CString(string(data))
}

// StopTunnelJSON stops a tunnel by name
//
//export StopTunnelJSON
func StopTunnelJSON(name *C.char) *C.char {
	if name == nil {
		resp := JSONResponse{
			Success: false,
			Error:   "name is nil",
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}

	nameStr := C.GoString(name)

	tunnelsMu.Lock()
	tunnel, exists := tunnels[nameStr]
	if !exists {
		tunnelsMu.Unlock()
		resp := JSONResponse{
			Success: false,
			Error:   fmt.Sprintf("tunnel %s not found", nameStr),
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}
	delete(tunnels, nameStr)
	tunnelsMu.Unlock()

	// Note: We don't have explicit Close() on the Tunnel interface,
	// but removing from map releases the reference
	_ = tunnel

	resp := JSONResponse{
		Success: true,
		Data: map[string]string{
			"name": nameStr,
		},
	}
	data, _ := json.Marshal(resp)
	return C.CString(string(data))
}

// GetStatusJSON gets tunnel status as JSON
//
//export GetStatusJSON
func GetStatusJSON(name *C.char) *C.char {
	if name == nil {
		resp := JSONResponse{
			Success: false,
			Error:   "name is nil",
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}

	nameStr := C.GoString(name)

	tunnelsMu.RLock()
	tunnel, exists := tunnels[nameStr]
	tunnelsMu.RUnlock()

	if !exists {
		resp := JSONResponse{
			Success: false,
			Error:   fmt.Sprintf("tunnel %s not found", nameStr),
		}
		data, _ := json.Marshal(resp)
		return C.CString(string(data))
	}

	// Build status
	status := TunnelStatusJSON{
		Name:    nameStr,
		Running: tunnel.IsAlive(),
		RxBytes: 0, // These would be populated from tunnel metrics if available
		TxBytes: 0,
	}

	resp := JSONResponse{
		Success: true,
		Data:    status,
	}
	data, _ := json.Marshal(resp)
	return C.CString(string(data))
}

// StartTunnel starts a tunnel from JSON config (struct API)
//
//export StartTunnel
func StartTunnel(name *C.char, configJSON *C.char) (result C.int) {
	defer func() {
		if r := recover(); r != nil {
			result = -1
		}
	}()

	if name == nil || configJSON == nil {
		return -1
	}

	nameStr := C.GoString(name)
	configStr := C.GoString(configJSON)

	// Check if tunnel already exists
	tunnelsMu.RLock()
	if _, exists := tunnels[nameStr]; exists {
		tunnelsMu.RUnlock()
		return -1
	}
	tunnelsMu.RUnlock()

	// Parse JSON config
	var config core.TunnelConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return -1
	}

	tunnel := core.TunnelFactory(nameStr, config)

	// Store tunnel
	tunnelsMu.Lock()
	tunnels[nameStr] = tunnel
	tunnelsMu.Unlock()

	// Start tunnel in a goroutine
	go tunnel.ListenAndServer()

	return 0
}

// StopTunnel stops a tunnel by name (struct API)
//
//export StopTunnel
func StopTunnel(name *C.char) C.int {
	if name == nil {
		return -1
	}

	nameStr := C.GoString(name)

	tunnelsMu.Lock()
	tunnel, exists := tunnels[nameStr]
	if !exists {
		tunnelsMu.Unlock()
		return -1
	}
	delete(tunnels, nameStr)
	tunnelsMu.Unlock()

	_ = tunnel
	return 0
}

// GetStatus gets tunnel status as a C struct
//
//export GetStatus
func GetStatus(name *C.char) C.TunnelStatus {
	if name == nil {
		return C.TunnelStatus{
			name:     C.CString(""),
			running:  0,
			rx_bytes: 0,
			tx_bytes: 0,
			error:    C.CString("name is nil"),
		}
	}

	nameStr := C.GoString(name)

	tunnelsMu.RLock()
	tunnel, exists := tunnels[nameStr]
	tunnelsMu.RUnlock()

	if !exists {
		return C.TunnelStatus{
			name:     C.CString(nameStr),
			running:  0,
			rx_bytes: 0,
			tx_bytes: 0,
			error:    C.CString(fmt.Sprintf("tunnel %s not found", nameStr)),
		}
	}

	running := 0
	if tunnel.IsAlive() {
		running = 1
	}

	return C.TunnelStatus{
		name:     C.CString(nameStr),
		running:  C.int(running),
		rx_bytes: 0,
		tx_bytes: 0,
		error:    C.CString(""),
	}
}

// FreeString frees a C string allocated by this library
//
//export FreeString
func FreeString(s *C.char) {
	C.cgo_free(unsafe.Pointer(s))
}

// main is required for buildmode=c-shared
func main() {}
