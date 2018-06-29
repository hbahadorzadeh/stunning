package client

import (
	"crypto/tls"
	"golang.org/x/net/proxy"
	tcommon "hbx.ir/stunning/tunnel/common"
	"net"
)

type TlsDialer struct {
	tcommon.TunnelDialer
	network, addr string
	dialer        proxy.Dialer
}

func GetTlsDialer() TlsDialer {
	return TlsDialer{}
}

func (d TlsDialer) Dial(network, addr string) (c net.Conn, err error) {

	if network == "" {
		network = d.network
	}

	if addr == "" {
		addr = d.addr
	}
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial(network, addr, conf)
	return conn, err
}
