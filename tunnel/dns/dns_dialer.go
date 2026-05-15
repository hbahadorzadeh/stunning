// Package dns provides DNS tunnel interface.
package dns

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"net"
)

type DnsDialer struct {
	tcommon.TunnelDialer
}

func GetDnsDialer() DnsDialer {
	return DnsDialer{}
}

func (_ DnsDialer) Dial(network, addr string) (c net.Conn, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &DnsConn{conn: conn}, nil
}

func (_ DnsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Dns
}
