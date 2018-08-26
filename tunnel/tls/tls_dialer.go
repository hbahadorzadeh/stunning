package tls

import (
	"crypto/tls"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
)

type TlsDialer struct {
	tcommon.TunnelDialer
	dialer proxy.Dialer
}

func GetTlsDialer() TlsDialer {
	return TlsDialer{}
}

func (d TlsDialer) Dial(network, addr string) (c net.Conn, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial(network, addr, conf)
	return conn, err
}

func (d TlsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tls
}
