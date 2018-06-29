package client

import (
	"fmt"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
)

type UdpServer struct {
	tcommon.TunnelServer
	conn *net.UDPConn
}

func StartTlsServer(address string) *UdpServer {
	log.SetFlags(log.Lshortfile)

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	sconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}

	serv := &UdpServer{}
	serv.conn = sconn
	return serv
}

func (s *UdpServer) WaitingForConnection() {
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
