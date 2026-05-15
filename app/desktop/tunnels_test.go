package main

import (
	"sync"
	"testing"
	"time"

	"github.com/hbahadorzadeh/stunning/core"
)

func TestTunnelInstanceCreation(t *testing.T) {
	config := core.TunnelConfig{
		ServiceMode:   "server",
		ServerType:    "tcp",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:8888",
	}

	inst := &TunnelInstance{
		config: config,
		stats: TunnelStats{
			startTime: time.Time{},
		},
	}

	if inst.config.ServiceMode != "server" {
		t.Errorf("Expected service mode 'server', got '%s'", inst.config.ServiceMode)
	}

	if inst.tunnel != nil {
		t.Error("Expected tunnel to be nil initially")
	}
}

func TestTunnelStatsMutex(t *testing.T) {
	stats := TunnelStats{
		rxBytes: 100,
		txBytes: 200,
	}

	// Test concurrent access to stats
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.mu.Lock()
			stats.rxBytes += 10
			stats.txBytes += 20
			stats.mu.Unlock()
		}()
	}

	wg.Wait()

	if stats.rxBytes != 200 {
		t.Errorf("Expected rxBytes 200, got %d", stats.rxBytes)
	}

	if stats.txBytes != 400 {
		t.Errorf("Expected txBytes 400, got %d", stats.txBytes)
	}
}

func TestFormatUptime(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{3660 * time.Second, "1h 1m 0s"},
	}

	for _, tc := range testCases {
		result := formatUptime(tc.duration)
		if result != tc.expected {
			t.Errorf("For duration %v, expected '%s', got '%s'", tc.duration, tc.expected, result)
		}
	}
}

func TestConcurrentTunnelAccess(t *testing.T) {
	tunnelsMutex = sync.RWMutex{}
	tunnels = make(map[string]*TunnelInstance)

	config := core.TunnelConfig{
		ServiceMode: "server",
	}

	tunnels["test"] = &TunnelInstance{
		config: config,
	}

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tunnelsMutex.RLock()
			_ = tunnels["test"]
			tunnelsMutex.RUnlock()
		}()
	}

	wg.Wait()

	// Verify tunnel still exists
	tunnelsMutex.RLock()
	if _, exists := tunnels["test"]; !exists {
		t.Error("Expected tunnel to exist after concurrent access")
	}
	tunnelsMutex.RUnlock()
}

func TestSaveTunnel(t *testing.T) {
	tunnelsMutex = sync.RWMutex{}
	tunnels = make(map[string]*TunnelInstance)

	config := core.TunnelConfig{
		ServiceMode: "server",
		ServerType:  "tcp",
	}

	if err := saveTunnel("test-tunnel", config); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	tunnelsMutex.RLock()
	inst, exists := tunnels["test-tunnel"]
	tunnelsMutex.RUnlock()

	if !exists {
		t.Error("Expected tunnel to be saved")
	}

	if inst.config.ServiceMode != "server" {
		t.Errorf("Expected service mode 'server', got '%s'", inst.config.ServiceMode)
	}
}

func TestDeleteTunnel(t *testing.T) {
	tunnelsMutex = sync.RWMutex{}
	tunnels = make(map[string]*TunnelInstance)

	tunnels["test-tunnel"] = &TunnelInstance{
		config: core.TunnelConfig{},
	}

	if err := deleteTunnel("test-tunnel"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	tunnelsMutex.RLock()
	_, exists := tunnels["test-tunnel"]
	tunnelsMutex.RUnlock()

	if exists {
		t.Error("Expected tunnel to be deleted")
	}
}
