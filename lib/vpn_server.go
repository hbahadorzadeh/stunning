package lib

import "net"

type Vpnserver interface {
	HandleConnection(conn net.Conn) error
}
