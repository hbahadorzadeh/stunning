package dns

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
)

type DnsServer struct {
	tcommon.TunnelServerCommon
}

func StartDnsServer(address string) (*DnsServer, error) {
	serv := &DnsServer{}
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return serv, err
	}
	serv.Listener = ln
	return serv, nil
}

func (s *DnsServer) HandleConnection(conn net.Conn) {
	defer conn.Close()

	dnsConn := &DnsConn{conn: conn}
	s.Server.HandleConnection(dnsConn)
}
