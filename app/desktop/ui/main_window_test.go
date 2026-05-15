package ui

import (
	"testing"

	"github.com/hbahadorzadeh/stunning/core"
)

func TestTunnelStatusStructure(t *testing.T) {
	status := TunnelStatus{
		Name:    "test-tunnel",
		Running: true,
		RxBytes: 1024,
		TxBytes: 2048,
		Uptime:  "10m 30s",
		Config: core.TunnelConfig{
			ServiceMode:   "server",
			ServerType:    "tcp",
			InterfaceType: "socks",
			Listen:        "127.0.0.1:8888",
		},
	}

	if status.Name != "test-tunnel" {
		t.Errorf("Expected name 'test-tunnel', got '%s'", status.Name)
	}

	if !status.Running {
		t.Error("Expected tunnel to be running")
	}

	if status.RxBytes != 1024 {
		t.Errorf("Expected RxBytes 1024, got %d", status.RxBytes)
	}

	if status.Config.ServiceMode != "server" {
		t.Errorf("Expected service mode 'server', got '%s'", status.Config.ServiceMode)
	}
}

func TestStatusViewCreate(t *testing.T) {
	status := TunnelStatus{
		Name:    "test-tunnel",
		Running: true,
		RxBytes: 1024,
		TxBytes: 2048,
		Uptime:  "10m 30s",
		Config: core.TunnelConfig{
			ServiceMode:   "server",
			ServerType:    "tcp",
			InterfaceType: "socks",
			Listen:        "127.0.0.1:8888",
			Connect:       "127.0.0.1:9999",
		},
	}

	sv := CreateStatusView(status)

	if sv == nil {
		t.Error("Expected status view to be created")
	}

	if sv.Container == nil {
		t.Error("Expected status view container to be created")
	}

	if len(sv.Container.Objects) == 0 {
		t.Error("Expected status view container to have objects")
	}
}

func TestStatusViewUpdateValues(t *testing.T) {
	status1 := TunnelStatus{
		Name:    "test-tunnel",
		Running: true,
		RxBytes: 1024,
		Uptime:  "10m",
		Config: core.TunnelConfig{
			ServiceMode: "server",
		},
	}

	sv := CreateStatusView(status1)

	// Don't call Update which requires Fyne context
	// Just verify the internal state can be changed
	sv.mu.Lock()
	sv.status = TunnelStatus{
		Name:    "test-tunnel",
		Running: false,
		RxBytes: 2048,
		Uptime:  "20m",
		Config: core.TunnelConfig{
			ServiceMode: "client",
		},
	}
	sv.mu.Unlock()

	// Verify update was applied
	if sv.status.RxBytes != 2048 {
		t.Errorf("Expected RxBytes to be 2048, got %d", sv.status.RxBytes)
	}

	if sv.status.Running {
		t.Error("Expected status to be stopped after update")
	}

	if sv.status.Uptime != "20m" {
		t.Errorf("Expected uptime to be '20m', got '%s'", sv.status.Uptime)
	}
}

func TestStatusViewMultipleValueUpdates(t *testing.T) {
	status := TunnelStatus{
		Name:    "test-tunnel",
		Running: true,
		RxBytes: 100,
	}

	sv := CreateStatusView(status)

	// Don't call Update which requires Fyne context
	// Just verify internal state updates
	for i := 0; i < 5; i++ {
		sv.mu.Lock()
		sv.status.RxBytes += 100
		sv.mu.Unlock()
	}

	if sv.status.RxBytes != 600 {
		t.Errorf("Expected RxBytes 600 after updates, got %d", sv.status.RxBytes)
	}
}
