package icmp

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"testing"
)

func TestStartIcmpServer(t *testing.T) {
	server, err := StartIcmpServer("0.0.0.0")
	if err != nil {
		t.Skip("ICMP requires root or CAP_NET_RAW")
	}
	if server == nil {
		t.Fatal("Server is nil")
	}
	server.Close()
}

func TestGetIcmpDialer(t *testing.T) {
	dialer := GetIcmpDialer()
	if dialer.Protocol() != tcommon.Icmp {
		t.Errorf("Expected Icmp protocol, got %v", dialer.Protocol())
	}
}
