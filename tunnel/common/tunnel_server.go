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

func (s TunnelServer) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
}

func (s TunnelServer) WaitingForConnection() {
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
	log.Printf("Listening on %s stopped\n", s.Listener.Addr().String())
}

func (s TunnelServer) Close() error {
	log.Println("Closing connection")
	return s.Listener.Close()
}

func (s TunnelServer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	s.Server.HandleConnection(conn)
}
