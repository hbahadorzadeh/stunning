package ws

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"testing"
)

func TestStartWsServer(t *testing.T) {
	server, err := StartWsServer("cert.pem", "key.pem", "127.0.0.1:9443")
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}
	if server == nil {
		t.Fatal("Server is nil")
	}
}

func TestGetWsDialer(t *testing.T) {
	dialer := GetWsDialer()
	if dialer.Protocol() != tcommon.Ws {
		t.Errorf("Expected Ws protocol, got %v", dialer.Protocol())
	}
}
