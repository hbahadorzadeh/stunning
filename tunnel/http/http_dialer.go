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

func (d HttpDialer) Dial(network, addr string) (c net.Conn, err error) {
	return tcommon.GetCilentHttpConnection("http", addr)
}

func (d HttpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
