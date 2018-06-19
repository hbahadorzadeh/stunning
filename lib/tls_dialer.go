package lib

import (
	"crypto/tls"
	"golang.org/x/net/proxy"
	"net"
)

type tls_dialer struct {
	proxy.Dialer
}

func GetTlsDialer() *tls_dialer {
	return &tls_dialer{}
}

func (d *tls_dialer) Dial(network, addr string) (c net.Conn, err error) {

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial(network, addr, conf)
	return conn, err
}
