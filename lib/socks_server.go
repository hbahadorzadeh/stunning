package lib

import (
	"github.com/getlantern/go-socks5"
	"net"
	"log"
)

type socks_server struct {
	conf   *socks5.Config
	server *socks5.Server
}

func get_socks_server() *socks_server {
	s := &socks_server{}
	s.conf = &socks5.Config{}
	server, err := socks5.New(s.conf)
	if err != nil {
		panic(err)
	}
	s.server = server
	return s
}

func (s *socks_server) handle_connection(conn net.Conn) error {
	log.Println("Serving connection")
	return s.server.ServeConn(conn)
}
