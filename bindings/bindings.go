// Package bindings provides Go bindings for native mobile platforms
package bindings

import (
	"fmt"
	"sync"

	"github.com/hbahadorzadeh/stunning/core"
)

var (
	mu                  sync.RWMutex
	currentTunnel       core.Tunnel
	vpnConnected        bool
	tunnelError         string
	tunnelDoneFunc      func()
	tunnelCancelContext chan struct{}
)

// ConnectRequest holds VPN connection parameters
type ConnectRequest struct {
	ServerAddress string
	Protocol      string // tcp, tls, h2, wss, etc.
	InterfaceType string // socks, tun, tcp
	ListenAddr    string // for tun interface
}

// Status holds current connection status
type Status struct {
	Connected bool
	Error     string
	Uptime    string
	RxBytes   uint64
	TxBytes   uint64
}

// Connect initiates a VPN tunnel to the server
//
//export Connect
func Connect(serverAddr string, protocol string) string {
	mu.Lock()
	if currentTunnel != nil && currentTunnel.IsAlive() {
		mu.Unlock()
		return "tunnel_already_running"
	}

	// Cancel previous tunnel goroutine if any
	if tunnelCancelContext != nil {
		close(tunnelCancelContext)
	}
	tunnelCancelContext = make(chan struct{})
	cancelChan := tunnelCancelContext
	mu.Unlock()

	// Build tunnel config for client mode
	config := core.TunnelConfig{
		ServiceMode:   "client",
		ServerType:    protocol,
		InterfaceType: "socks",       // Default to SOCKS for client
		Listen:        "127.0.0.1:0", // Let OS assign available port
		Connect:       serverAddr,
	}

	tunnel := core.TunnelFactory("mobile-vpn", config)

	mu.Lock()
	currentTunnel = tunnel
	vpnConnected = true
	tunnelError = ""
	mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				mu.Lock()
				tunnelError = fmt.Sprintf("panic: %v", r)
				vpnConnected = false
				currentTunnel = nil
				mu.Unlock()
			}
		}()

		// Use a goroutine to run the tunnel with cancellation support
		tunnelDone := make(chan struct{})
		go func() {
			defer close(tunnelDone)
			tunnel.ListenAndServer()
		}()

		select {
		case <-tunnelDone:
			mu.Lock()
			vpnConnected = false
			currentTunnel = nil
			mu.Unlock()
		case <-cancelChan:
			mu.Lock()
			vpnConnected = false
			currentTunnel = nil
			mu.Unlock()
			return
		}

		if tunnelDoneFunc != nil {
			tunnelDoneFunc()
		}
	}()

	return "ok"
}

// Disconnect stops the active VPN tunnel
//
//export Disconnect
func Disconnect() string {
	mu.Lock()
	if currentTunnel == nil {
		// Still need to clean up cancel context if it exists
		if tunnelCancelContext != nil {
			select {
			case <-tunnelCancelContext:
				// Already closed
			default:
				close(tunnelCancelContext)
			}
			tunnelCancelContext = nil
		}
		mu.Unlock()
		return "not_connected"
	}

	currentTunnel = nil
	vpnConnected = false

	// Signal the tunnel goroutine to stop
	if tunnelCancelContext != nil {
		select {
		case <-tunnelCancelContext:
			// Already closed
		default:
			close(tunnelCancelContext)
		}
		tunnelCancelContext = nil
	}
	mu.Unlock()

	return "ok"
}

// GetStatus returns current VPN connection status
//
//export GetStatus
func GetStatus() string {
	mu.RLock()
	defer mu.RUnlock()

	status := "disconnected"
	if vpnConnected && currentTunnel != nil && currentTunnel.IsAlive() {
		status = "connected"
	}

	if tunnelError != "" {
		status = "error: " + tunnelError
	}

	return status
}

// IsConnected checks if VPN is currently connected
//
//export IsConnected
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()

	return vpnConnected && currentTunnel != nil && currentTunnel.IsAlive()
}
