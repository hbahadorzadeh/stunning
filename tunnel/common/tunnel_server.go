package common

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	"log"
	"net"
)

type TunnelServer interface {
	SetServer(server icommon.TunnelInterfaceServer)
	WaitingForConnection()
	Close() error
	Closed() bool
	HandleConnection(conn net.Conn)
}

type TunnelServerCommon struct {
	TunnelServer
	closed   bool
	Server   icommon.TunnelInterfaceServer
	Listener net.Listener
}

func (s TunnelServerCommon) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
}

func (s TunnelServerCommon) WaitingForConnection() {
	s.closed = false
	log.Printf("listening for connection on %s\n", s.Listener.Addr().String())
	for {
		conn, err := s.Listener.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			break
		}
		go s.HandleConnection(conn)
	}
	s.closed = true
	log.Printf("Listening on %s stopped\n", s.Listener.Addr().String())
}

func (s TunnelServerCommon) Close() error {
	log.Println("Closing connection")
	err := s.Listener.Close()
	s.closed = true
	return err
}

func (s TunnelServerCommon) Closed() bool {
	return s.closed
}

func (s TunnelServerCommon) HandleConnection(conn net.Conn) {
	defer conn.Close()
	s.Server.HandleConnection(conn)
}
