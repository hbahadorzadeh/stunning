package net

import (
	"log"
	"net"
)

type tcp_server struct {
	Vpnserver
	address string
	conn    net.Conn
}

func GetTcpServer(url string) *tcp_server {
	s := &tcp_server{}
	s.address = url

	conn, err := net.Dial("tcp", s.address)

	if err != nil {
		panic(err)
	}

	s.conn = conn

	return s
}

func (t *tcp_server) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go tcp_server_reader(conn, t.conn)
	tcp_server_writer(conn, t.conn)
	return nil
}

func tcp_server_reader(conn net.Conn, tconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = buff[:n]
		wn, werr := tconn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", tconn.RemoteAddr().String(), wn)
	}
}

func tcp_server_writer(conn net.Conn, tconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := tconn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = buff[:n]
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}
