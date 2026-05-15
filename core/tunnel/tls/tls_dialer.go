// Package tls provides TLS tunnel implementation.
package tls

import (
	"crypto/tls"
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

type TlsDialer struct {
	tcommon.TunnelDialer
}

func GetTlsDialer() TlsDialer {
	return TlsDialer{}
}

func (TlsDialer) Dial(network, addr string) (c net.Conn, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial(network, addr, conf)
	return conn, err
}

func GetTlsDialerWithConfig(cfg *tls.Config) TlsDialer {
	return TlsDialer{}
}

func (TlsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Tls
}
