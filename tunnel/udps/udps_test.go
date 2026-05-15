package udps

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"testing"
)

func TestStartUdpsServer(t *testing.T) {
	server, err := StartUdpsServer("cert.pem", "key.pem", "127.0.0.1:9443")
	if err == nil || server == nil {
		return
	}
}

func TestGetUdpsDialer(t *testing.T) {
	dialer := GetUdpsDialer()
	if dialer.Protocol() != tcommon.Udps {
		t.Errorf("Expected Udps protocol, got %v", dialer.Protocol())
	}
}
