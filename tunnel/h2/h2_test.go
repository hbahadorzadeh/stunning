package h2

import (
	"testing"

	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
)

func TestStartH2Server(t *testing.T) {
	server, err := StartH2Server("cert.pem", "key.pem", "127.0.0.1:9443")
	if err == nil || server == nil {
		return
	}
}

func TestGetH2Dialer(t *testing.T) {
	dialer := GetH2Dialer()
	if dialer.Protocol() != tcommon.H2 {
		t.Errorf("Expected H2 protocol, got %v", dialer.Protocol())
	}
}
