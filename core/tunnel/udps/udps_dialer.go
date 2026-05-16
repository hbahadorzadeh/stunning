// Package udps provides DTLS tunnel implementation.
package udps

import (
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
	"github.com/pion/dtls/v3"
)

type UdpsDialer struct {
	tcommon.TunnelDialer
}

func GetUdpsDialer() UdpsDialer {
	return UdpsDialer{}
}

func (UdpsDialer) Dial(network, addr string) (c net.Conn, err error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	// Use default options-based API
	conn, err := dtls.DialWithOptions("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (UdpsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Udps
}
