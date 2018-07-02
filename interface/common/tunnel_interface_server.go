package common

import "net"

type TunnelInterfaceServer interface {
	HandleConnection(conn net.Conn) error
	Close() error
}
