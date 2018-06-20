package net

import (
	"log"
	"net"
)

type udp_server struct {
	Vpnserver
	address string
	conn    net.Conn
}

func GetUdpServer(url string) *udp_server {
	s := &udp_server{}
	s.address = url

	conn, err := net.Dial("udp", s.address)

	if err != nil {
		panic(err)
	}

	s.conn = conn

	return s
}

func (t *udp_server) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go t.udp_server_reader(conn)
	t.udp_server_writer(conn)
	return nil
}

func (t *udp_server) udp_server_reader(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = buff[:n]
		wn, werr := t.conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", t.conn.RemoteAddr().String(), wn)
	}
}

func (t *udp_server) udp_server_writer(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := t.conn.Read(buff)
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
