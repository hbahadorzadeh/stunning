package common

import (
	"net"
)

type TunnelDialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}