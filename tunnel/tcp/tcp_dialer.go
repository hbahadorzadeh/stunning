// Package tcp provides TCP tunnel implementation.
package tcp

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type TcpDialer struct {
	tcommon.TunnelDialer
}

func GetTcpDialer() TcpDialer {
	d := TcpDialer{}
	return d
}

func (_ TcpDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (_ TcpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
