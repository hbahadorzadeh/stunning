package client

import (
	tcommon "hbx.ir/stunning/tunnel/common"
	"log"
	"net"
)

type TcpServer struct {
	tcommon.TunnelServer
}

func StartTcpServer(address string) *TcpServer {
	log.SetFlags(log.Lshortfile)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return nil
	}
	serv := &TcpServer{}
	serv.Listener = ln
	return serv
}

func (s *TcpServer) WaitingForConnection() {
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
