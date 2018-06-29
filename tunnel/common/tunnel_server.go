package common

import (
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	"log"
	"net"
)

type TunnelServer struct {
	Server   icommon.TunnelInterfaceServer
	Listener net.Listener
}

func (s *TunnelServer) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
	go s.WaitingForConnection()
}

func (s *TunnelServer) WaitingForConnection() {
	for {
		conn, err := s.Listener.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			continue
		}
		go s.HandleConnection(conn)
	}
}

func (s *TunnelServer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	s.Server.HandleConnection(conn)
}
