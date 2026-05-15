// Package h2 provides HTTP/2 tunnel dialer.
package h2

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type H2Dialer struct {
	tcommon.TunnelDialer
}

func GetH2Dialer() H2Dialer {
	return H2Dialer{}
}

func (_ H2Dialer) Dial(network, addr string) (c net.Conn, err error) {
	return tcommon.GetCilentHttpConnection("https", addr)
}

func (_ H2Dialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.H2
}
