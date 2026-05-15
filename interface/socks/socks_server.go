// Package socks provides SOCKS5 tunnel server interface.
package socks

import (
	"log"
	"net"

	"github.com/getlantern/go-socks5"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
)

type SocksServer struct {
	icommon.TunnelInterfaceServer
	conf   *socks5.Config
	server *socks5.Server
}

func GetSocksServer() *SocksServer {
	s := &SocksServer{}
	s.conf = &socks5.Config{}
	server, err := socks5.New(s.conf)
	if err != nil {
		panic(err)
	}
	s.server = server
	return s
}

func (*SocksServer) WaitingForConnection() {
	// No-op: WaitingForConnection is managed by the tunnel server
}

func (*SocksServer) Close() error {
	return nil
}

func (s *SocksServer) HandleConnection(conn net.Conn) error {
	log.Printf("Serving connection")
	return s.server.ServeConn(conn)
}
