package common

import (
	"net"
)

type TunnelInterfaceClient interface {
	HandleConnection(conn net.Conn, tconn net.Conn) error
	WaitingForConnection()
	Close()
	Closed() bool
}
