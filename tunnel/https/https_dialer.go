package https

import (
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type HttpsDialer struct {
	tcommon.TunnelDialer
	dialer proxy.Dialer
}

func GetHttpsDialer() HttpsDialer {
	d := HttpsDialer{}
	return d
}

func (d HttpsDialer) Dial(network, addr string) (c net.Conn, err error) {
	return tcommon.GetCilentHttpConnection("https", addr)
}

func (d HttpsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tcp
}
