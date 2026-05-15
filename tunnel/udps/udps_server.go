package udps

import (
	"crypto/tls"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"github.com/pion/dtls/v2"
	"log"
	"net"
)

type UdpsServer struct {
	tcommon.TunnelServerCommon
}

func StartUdpsServer(crt, key, address string) (*UdpsServer, error) {
	serv := &UdpsServer{}

	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return serv, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return serv, err
	}

	config := &dtls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   dtls.NoClientCert,
	}

	ln, err := dtls.Listen("udp", udpAddr, config)
	if err != nil {
		log.Println(err)
		return serv, err
	}

	serv.Listener = ln
	return serv, nil
}
