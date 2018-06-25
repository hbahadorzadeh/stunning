package client

import (
	tcommon "hbx.ir/stunning/lib/net/tunnel/common"
	"golang.org/x/net/proxy"
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
