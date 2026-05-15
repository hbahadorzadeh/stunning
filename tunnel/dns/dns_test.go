package dns

import (
	"testing"

	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
)

func TestStartDnsServer(t *testing.T) {
	server, err := StartDnsServer("127.0.0.1:9053")
	if err != nil {
		t.Fatalf("Failed to start DNS server: %v", err)
	}
	if server == nil {
		t.Fatal("Server is nil")
	}
	server.Close()
}

func TestGetDnsDialer(t *testing.T) {
	dialer := GetDnsDialer()
	if dialer.Protocol() != tcommon.Dns {
		t.Errorf("Expected Dns protocol, got %v", dialer.Protocol())
	}
}
