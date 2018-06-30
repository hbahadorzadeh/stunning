package tls

import (
	"crypto/tls"
	"golang.org/x/net/proxy"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"net"
)

type TlsDialer struct {
	tcommon.TunnelDialer
	dialer        proxy.Dialer
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
