package bindings

import (
	"testing"
)

// Test basic connect/disconnect flow
func TestConnectFlow(t *testing.T) {
	result := Connect("127.0.0.1:8888", "tcp")
	if result != "ok" {
		t.Errorf("Connect: expected 'ok', got '%s'", result)
	}

	result = Disconnect()
	if result != "ok" {
		t.Errorf("Disconnect: expected 'ok', got '%s'", result)
	}
}

// Test disconnect when not connected
func TestDisconnectNotConnected(t *testing.T) {
	Disconnect() // Clean up from previous tests

	result := Disconnect()
	if result != "not_connected" {
		t.Errorf("Expected 'not_connected', got '%s'", result)
	}
}

// Test getting status when not connected
func TestGetStatusWhenDisconnected(t *testing.T) {
	Disconnect() // Clean up from previous tests

	status := GetStatus()
	if status == "" {
		t.Error("Expected non-empty status string")
	}
}

// Test IsConnected returns false when not connected
func TestIsConnectedWhenDisconnected(t *testing.T) {
	Disconnect() // Clean up from previous tests

	connected := IsConnected()
	if connected {
		t.Error("Expected IsConnected() to return false")
	}
}
