package udp

import (
	"golang.org/x/net/proxy"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"net"
)

type UdpDialer struct {
	tcommon.TunnelDialer
	dialer        proxy.Dialer
}

func GetUdpDialer() UdpDialer {
	d := UdpDialer{}
	return d
}

func (d UdpDialer) Dial(network, addr string) (c net.Conn, err error) {
	rudpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	ludpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5050")
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP(network, ludpAddr, rudpAddr)
	if err != nil {
		return nil, err
	}
	return udp_connection{
		conn: conn,
	}, err
}
