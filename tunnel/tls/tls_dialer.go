// Package tls provides TLS tunnel implementation.
package tls

import (
	"crypto/tls"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type TlsDialer struct {
	tcommon.TunnelDialer
}

func GetTlsDialer() TlsDialer {
	return TlsDialer{}
}

func (_ TlsDialer) Dial(network, addr string) (c net.Conn, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial(network, addr, conf)
	return conn, err
}

func GetTlsDialerWithConfig(cfg *tls.Config) TlsDialer {
	return TlsDialer{}
}

func (_ TlsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tls
}
