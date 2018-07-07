package http

import (
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
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
	return getCilentHttpConnection(addr)
}

func (d HttpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
