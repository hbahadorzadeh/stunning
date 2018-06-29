package client

import (
	"golang.org/x/net/proxy"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"net"
)

type TcpDialer struct {
	tcommon.TunnelDialer
	network, addr string
	dialer        proxy.Dialer
}

func GetTcpDialer() TcpDialer {
	d := TcpDialer{}
	d.network = "tcp"
	return d
}

func (d TcpDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial(d.network, d.addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}
