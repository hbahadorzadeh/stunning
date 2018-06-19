package lib

import (
	"crypto/tls"
	"log"
	"net"
)

type TlsServer struct {
	ln     net.Listener
	server Vpnserver
}

func StartTlsServer() *TlsServer {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Println(err)
		return nil
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":4443", config)

	if err != nil {
		log.Println(err)
		return nil
	}
	//defer ln.Close()
	serv := &TlsServer{
		ln: ln,
	}
	return serv
}

func (s *TlsServer) SetSocksServer(ss *socks_server) {
	s.server = ss
	go s.waitingForConnection()
}

func (s *TlsServer) SetTunServer(ss *TunInterface) {
	s.server = ss
	go s.waitingForConnection()
}

func (s *TlsServer) waitingForConnection() {
	for {
		conn, err := s.ln.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TlsServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	s.server.HandleConnection(conn)
}
