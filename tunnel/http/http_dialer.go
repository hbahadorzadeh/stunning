package http

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type HttpDialer struct {
	tcommon.TunnelDialer
	dialer proxy.Dialer
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
