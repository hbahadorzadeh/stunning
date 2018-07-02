package socks

import (
	"github.com/getlantern/go-socks5"
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	"log"
	"net"
)

type SocksServer struct {
	icommon.TunnelInterfaceServer
	conf   *socks5.Config
	server *socks5.Server
}

func GetSocksServer() SocksServer {
	s := SocksServer{}
	s.conf = &socks5.Config{}
	server, err := socks5.New(s.conf)
	if err != nil {
		panic(err)
	}
	s.server = server
	return s
}

func (s SocksServer) HandleConnection(conn net.Conn) error {
	log.Printf("Serving connection")
	return s.server.ServeConn(conn)
}
