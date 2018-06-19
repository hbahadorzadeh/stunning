package net

import (
	"log"
	"net"
	"fmt"
)

type udp_client struct {
	address    string
	conn *net.UDPConn
}

func GetUdpClient(url string, conn net.Conn) *udp_client {
	s := &udp_client{}
	s.address = url
	udpAddr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	sconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	s.conn = sconn
	return s
}

func (t *udp_client) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go udp_server_reader(conn, t.conn)
	udp_server_writer(conn, t.conn)
	return nil
}
