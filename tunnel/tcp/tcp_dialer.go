package tcp

import (
	"golang.org/x/net/proxy"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"net"
)

type TcpDialer struct {
	tcommon.TunnelDialer
	dialer        proxy.Dialer
}

func GetTcpDialer() TcpDialer {
	d := TcpDialer{}
	return d
}

func (d TcpDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}
