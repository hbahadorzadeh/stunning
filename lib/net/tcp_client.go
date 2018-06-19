package net

import (
	"log"
	"net"
)

type tcp_client struct {
	address    string
	tls_dialer *tls_dialer
	saddress   string
	listen     net.Listener
}

func GetTcpClient(url, surl string, tls_dialer *tls_dialer) *tcp_client {
	s := &tcp_client{}
	s.address = url
	s.saddress = surl
	s.tls_dialer = tls_dialer
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Panic(err)
	}
	s.listen = listen
	go s.waiting_for_connection()
	return s
}

func (t *tcp_client) waiting_for_connection() {
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		sconn, serr := t.tls_dialer.Dial("tcp", t.saddress)
		if serr != nil {
			log.Fatalln(serr)
			continue
		}
		go t.HandleConnection(conn, sconn)
	}
}

func (t *tcp_client) HandleConnection(conn net.Conn, tconn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go tcp_server_reader(conn, tconn)
	tcp_server_writer(conn, tconn)
	return nil
}
