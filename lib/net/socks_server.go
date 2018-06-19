package net

import (
	"github.com/getlantern/go-socks5"
	"log"
	"net"
)

type socks_server struct {
	Vpnserver
	conf   *socks5.Config
	server *socks5.Server
}

func GetSocksServer() *socks_server {
	s := &socks_server{}
	s.conf = &socks5.Config{}
	server, err := socks5.New(s.conf)
	if err != nil {
		panic(err)
	}
	s.server = server
	return s
}

func (s *socks_server) HandleConnection(conn net.Conn) error {
	log.Println("Serving connection")
	return s.server.ServeConn(conn)
}
