package dns

import (
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
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

func (s *DnsServer) SetServer(server icommon.TunnelInterfaceServer) {
	s.Server = server
	go s.WaitingForConnection()
}

func (s *DnsServer) WaitingForConnection() {
	log.Printf("listening for connection on %s\n", s.Listener.Addr().String())
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go s.HandleConnection(conn)
	}
	log.Printf("Listening on %s stopped\n", s.Listener.Addr().String())
}

func (s *DnsServer) HandleConnection(conn net.Conn) {
	defer conn.Close()

	dnsConn := &DnsConn{conn: conn}
	s.Server.HandleConnection(dnsConn)
}
