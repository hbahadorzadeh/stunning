// Package icmp provides ICMP tunnel dialer.
package icmp

import (
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

type IcmpDialer struct {
	tcommon.TunnelDialer
}

func GetIcmpDialer() IcmpDialer {
	return IcmpDialer{}
}

func (IcmpDialer) Dial(network, addr string) (c net.Conn, err error) {
	remoteAddr, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveIPAddr("ip4", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	return &ClientIcmpConnection{
		conn: localAddr,
		addr: remoteAddr,
	}, nil
}

func (IcmpDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Icmp
}
