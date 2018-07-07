package tcp

import (
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type TcpDialer struct {
	tcommon.TunnelDialer
	dialer proxy.Dialer
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

func (d TcpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
