package client

import (
	tcommon "hbx.ir/stunning/lib/net/tunnel/common"
	"golang.org/x/net/proxy"
	"net"
	"crypto/tls"
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
