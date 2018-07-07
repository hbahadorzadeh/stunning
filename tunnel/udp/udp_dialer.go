package udp

import (
	"bufio"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type UdpDialer struct {
	tcommon.TunnelDialer
	dialer proxy.Dialer
}

func GetUdpDialer() UdpDialer {
	d := UdpDialer{}
	return d
}

func (d UdpDialer) Dial(network, addr string) (c net.Conn, err error) {
	rudpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	ludpAddr, err := net.ResolveUDPAddr(network, ":0")
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP(network, ludpAddr, rudpAddr)
	if err != nil {
		return nil, err
	}
	cnn := clientUdpConnection{
		conn:   conn,
		buff:   make([]byte, 1024),
		reader: bufio.NewReader(conn),
	}
	return cnn, err
}

func (d UdpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Udp
}
