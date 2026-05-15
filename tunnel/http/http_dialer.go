// Package http provides HTTP tunnel dialer.
package http

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type HttpDialer struct {
	tcommon.TunnelDialer
}

func GetHttpDialer() HttpDialer {
	d := HttpDialer{}
	return d
}

func (HttpDialer) Dial(network, addr string) (c net.Conn, err error) {
	return tcommon.GetCilentHttpConnection("http", addr)
}

func (HttpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
