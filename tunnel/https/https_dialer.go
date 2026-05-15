package https

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type HttpsDialer struct {
	tcommon.TunnelDialer
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
