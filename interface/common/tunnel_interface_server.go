package common

import "net"

type TunnelInterfaceServer interface {
	WaitingForConnection()
	HandleConnection(conn net.Conn) error
	Close() error
}
