package client

import (
	"crypto/tls"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
)

type TlsServer struct {
	tcommon.TunnelServer
}

func StartTlsServer(crt, key, address string) *TlsServer {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		log.Println(err)
		return nil
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)

	if err != nil {
		log.Println(err)
		return nil
	}
	//defer ln.Close()
	serv := &TlsServer{}
	serv.Listener = ln
	return serv
}

func (s *TlsServer) WaitingForConnection() {
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
