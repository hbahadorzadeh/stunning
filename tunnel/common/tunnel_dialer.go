package common

import (
	"net"
)

type TunnelDialer interface {
	Dial(network, addr string) (c net.Conn, err error)
	Protocol() TunnelProtocol
}

type TunnelProtocol string

const (
	Tcp TunnelProtocol = "tcp"
	Tls TunnelProtocol = "tcp"
	Udp TunnelProtocol = "udp"
)

func (t TunnelProtocol) String() string {
	if t == Udp {
		return "udp"
	} else {
		return "tcp"
	}
}
