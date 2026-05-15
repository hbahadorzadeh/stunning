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
	Tcp  TunnelProtocol = "tcp"
	Tls  TunnelProtocol = "tls"
	Udp  TunnelProtocol = "udp"
	H2   TunnelProtocol = "h2"
	Ws   TunnelProtocol = "ws"
	Udps TunnelProtocol = "udps"
	Dns  TunnelProtocol = "dns"
	Icmp TunnelProtocol = "icmp"
)

func (t TunnelProtocol) String() string {
	if t == Udp {
		return "udp"
	}
	if t == Udps {
		return "udp"
	}
	if t == Icmp {
		return "ip4:icmp"
	}
	return "tcp"
}
