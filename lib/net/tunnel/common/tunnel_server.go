package common

import (
	icommon "hbx.ir/stunning/lib/net/interface/common"
	//"hbx.ir/stunning/lib/net/interface/socks"
	//"hbx.ir/stunning/lib/net/interface/tcp"
	//"hbx.ir/stunning/lib/net/interface/tun"
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

//func (s *TunnelServer) SetSocksServer(ss *socks.SocksServer) {
//	s.Server = ss
//	go s.WaitingForConnection()
//}
//
//func (s *TunnelServer) SetTunServer(ss *tun.TunInterface) {
//	s.Server = ss
//	go s.WaitingForConnection()
//}
//
//func (s *TunnelServer) SetTcpServer(ss *tcp.TcpServer) {
//	s.Server = ss
//	go s.WaitingForConnection()
//}

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
