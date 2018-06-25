package client

import (
	tcommon "hbx.ir/stunning/lib/net/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type UdpDialer struct {
	tcommon.TunnelDialer
	network, addr string
	dialer        proxy.Dialer
}

func GetUdpDialer() UdpDialer {
	d := UdpDialer{}
	d.network = "udp"
	return d
}


func (d UdpDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial(d.network, d.addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}
