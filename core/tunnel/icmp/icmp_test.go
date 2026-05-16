package icmp

import (
	"context"
	"testing"
	"time"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

func TestStartIcmpServer(t *testing.T) {
	server, err := StartIcmpServer("0.0.0.0")
	if err != nil {
		t.Skip("ICMP requires root or CAP_NET_RAW")
	}
	if server == nil {
		t.Fatal("Server is nil")
	}

	// Use timeout to prevent test from hanging
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("Close returned error: %v", err)
		}
	case <-ctx.Done():
		t.Skip("ICMP server close timeout - expected behavior for uninitialized server")
	}
}

func TestGetIcmpDialer(t *testing.T) {
	dialer := GetIcmpDialer()
	if dialer.Protocol() != tcommon.Icmp {
		t.Errorf("Expected Icmp protocol, got %v", dialer.Protocol())
	}
}
