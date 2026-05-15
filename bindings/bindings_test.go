package bindings

import (
	"testing"
)

func TestConnect(t *testing.T) {
	result := Connect("127.0.0.1:8888", "tcp")
	if result != "ok" {
		t.Errorf("Expected 'ok', got '%s'", result)
	}

	if !IsConnected() {
		t.Error("Expected tunnel to be connected")
	}
}

func TestDisconnect(t *testing.T) {
	Connect("127.0.0.1:8888", "tcp")

	result := Disconnect()
	if result != "ok" {
		t.Errorf("Expected 'ok', got '%s'", result)
	}
}

func TestGetStatus(t *testing.T) {
	status := GetStatus()
	if status == "" {
		t.Error("Expected non-empty status string")
	}
}

func TestConnectAlreadyConnected(t *testing.T) {
	Connect("127.0.0.1:8888", "tcp")
	result := Connect("127.0.0.1:9999", "tls")

	if result != "tunnel_already_running" {
		t.Errorf("Expected 'tunnel_already_running', got '%s'", result)
	}
}

func TestDisconnectWhenNotConnected(t *testing.T) {
	Disconnect()

	result := Disconnect()
	if result != "not_connected" {
		t.Errorf("Expected 'not_connected', got '%s'", result)
	}
}
