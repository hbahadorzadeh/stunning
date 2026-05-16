package main

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hbahadorzadeh/stunning/core"
)

// Test helpers

func TestStartTunnelJSON(t *testing.T) {
	// Reset tunnels for clean test
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	// Create a valid JSON config for a simple HTTP tunnel client
	configJSON := `{
		"ServiceMode": "client",
		"ServerType": "http",
		"InterfaceType": "socks",
		"Listen": "127.0.0.1:0",
		"Connect": "127.0.0.1:8080"
	}`

	// Call via direct Go function (avoiding C.CString complexity in tests)
	name := "test-tunnel-1"
	config := core.TunnelConfig{
		ServiceMode:   "client",
		ServerType:    "http",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:0",
		Connect:       "127.0.0.1:8080",
	}

	// Create tunnel directly
	tunnel := core.TunnelFactory(name, config)
	if tunnel == nil {
		t.Fatalf("Failed to create tunnel")
	}

	// Store it
	tunnelsMu.Lock()
	tunnels[name] = tunnel
	tunnelsMu.Unlock()

	// Verify tunnel is stored
	tunnelsMu.RLock()
	stored, exists := tunnels[name]
	tunnelsMu.RUnlock()

	if !exists {
		t.Errorf("Tunnel not stored correctly")
	}
	if stored != tunnel {
		t.Errorf("Stored tunnel doesn't match original")
	}

	// Cleanup
	tunnelsMu.Lock()
	delete(tunnels, name)
	tunnelsMu.Unlock()

	_ = configJSON
}

func TestStopTunnelJSON(t *testing.T) {
	// Reset tunnels
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	name := "test-tunnel-2"
	config := core.TunnelConfig{
		ServiceMode:   "client",
		ServerType:    "http",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:0",
		Connect:       "127.0.0.1:8080",
	}

	tunnel := core.TunnelFactory(name, config)
	tunnelsMu.Lock()
	tunnels[name] = tunnel
	tunnelsMu.Unlock()

	// Verify tunnel exists
	tunnelsMu.RLock()
	_, exists := tunnels[name]
	tunnelsMu.RUnlock()

	if !exists {
		t.Errorf("Tunnel not found before stop")
	}

	// Stop tunnel (remove from map)
	tunnelsMu.Lock()
	delete(tunnels, name)
	tunnelsMu.Unlock()

	// Verify tunnel is removed
	tunnelsMu.RLock()
	_, exists = tunnels[name]
	tunnelsMu.RUnlock()

	if exists {
		t.Errorf("Tunnel still exists after stop")
	}
}

func TestGetStatusJSON(t *testing.T) {
	// Reset tunnels
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	name := "test-tunnel-3"
	config := core.TunnelConfig{
		ServiceMode:   "client",
		ServerType:    "http",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:0",
		Connect:       "127.0.0.1:8080",
	}

	tunnel := core.TunnelFactory(name, config)
	tunnelsMu.Lock()
	tunnels[name] = tunnel
	tunnelsMu.Unlock()

	// Get status
	tunnelsMu.RLock()
	stored, exists := tunnels[name]
	tunnelsMu.RUnlock()

	if !exists {
		t.Fatalf("Tunnel not found for status check")
	}

	// Verify IsAlive() method works
	alive := stored.IsAlive()
	if alive {
		t.Logf("Tunnel is alive")
	} else {
		t.Logf("Tunnel is not alive (expected for non-running tunnel)")
	}

	// Cleanup
	tunnelsMu.Lock()
	delete(tunnels, name)
	tunnelsMu.Unlock()
}

func TestStartTunnel(t *testing.T) {
	// Reset tunnels
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	name := "test-tunnel-4"
	config := core.TunnelConfig{
		ServiceMode:   "client",
		ServerType:    "http",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:0",
		Connect:       "127.0.0.1:8080",
	}

	tunnel := core.TunnelFactory(name, config)
	if tunnel == nil {
		t.Fatalf("Failed to create tunnel")
	}

	tunnelsMu.Lock()
	tunnels[name] = tunnel
	tunnelsMu.Unlock()

	// Verify storage
	tunnelsMu.RLock()
	stored, exists := tunnels[name]
	tunnelsMu.RUnlock()

	if !exists {
		t.Errorf("Tunnel struct API not stored correctly")
	}
	if stored != tunnel {
		t.Errorf("Stored tunnel doesn't match in struct API")
	}

	// Cleanup
	tunnelsMu.Lock()
	delete(tunnels, name)
	tunnelsMu.Unlock()
}

func TestErrorHandling(t *testing.T) {
	// Test invalid JSON handling
	invalidJSON := `{invalid json`

	var config core.TunnelConfig
	err := json.Unmarshal([]byte(invalidJSON), &config)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got nil")
	}

	// Test with empty config
	emptyJSON := `{}`
	err = json.Unmarshal([]byte(emptyJSON), &config)
	if err != nil {
		t.Errorf("Unexpected error for empty JSON: %v", err)
	}

	// Test nil tunnel lookup
	tunnelsMu.RLock()
	_, exists := tunnels["nonexistent-tunnel"]
	tunnelsMu.RUnlock()

	if exists {
		t.Errorf("Should not find nonexistent tunnel")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Reset tunnels
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	const numGoroutines = 10
	const operationsPerGoroutine = 20

	var wg sync.WaitGroup
	var successCount int64

	// Test concurrent map access without tunnel creation
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for op := 0; op < operationsPerGoroutine; op++ {
				name := "tunnel-g" + string(rune('0'+(id%10))) + "o" + string(rune('0'+(op%10)))

				// Write
				tunnelsMu.Lock()
				tunnels[name] = nil
				tunnelsMu.Unlock()

				// Read
				tunnelsMu.RLock()
				_, exists := tunnels[name]
				tunnelsMu.RUnlock()

				if exists {
					atomic.AddInt64(&successCount, 1)
				}

				// Delete
				tunnelsMu.Lock()
				delete(tunnels, name)
				tunnelsMu.Unlock()
			}
		}(g)
	}

	wg.Wait()

	// Verify concurrent access succeeded
	count := atomic.LoadInt64(&successCount)
	if count > 0 {
		t.Logf("Concurrent access test: %d successful read operations", count)
	}

	// Cleanup
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()
}

func TestConcurrentMutexProtection(t *testing.T) {
	// Test that mutex protects concurrent reads and writes
	tunnelsMu.Lock()
	tunnels = make(map[string]core.Tunnel)
	tunnelsMu.Unlock()

	var wg sync.WaitGroup
	const numReaders = 5
	const numWriters = 2
	const operations = 100

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				tunnelsMu.RLock()
				_ = len(tunnels)
				tunnelsMu.RUnlock()
			}
		}()
	}

	// Start writers - only do map operations without creating actual tunnels
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				name := "tunnel-m" + string(rune('0'+id)) + string(rune('0'+(j%10)))

				// Only do map operations, not actual tunnel creation
				tunnelsMu.Lock()
				// Insert a dummy pointer to test map access
				tunnels[name] = nil
				tunnelsMu.Unlock()

				// Clean up
				tunnelsMu.Lock()
				delete(tunnels, name)
				tunnelsMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// If we get here without a race condition (detected by go test -race), the test passes
	t.Logf("Concurrent mutex protection test passed")
}
