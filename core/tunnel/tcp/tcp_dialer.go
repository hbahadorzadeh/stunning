// Package tcp provides TCP tunnel implementation.
package tcp

import (
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

type TcpDialer struct {
	tcommon.TunnelDialer
}

func GetTcpDialer() TcpDialer {
	d := TcpDialer{}
	return d
}

func (TcpDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (TcpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
