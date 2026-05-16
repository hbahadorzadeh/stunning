package udps

import (
	"crypto/tls"
	"log"
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
	"github.com/pion/dtls/v3"
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

	// Use new options-based API
	ln, err := dtls.ListenWithOptions("udp", udpAddr,
		dtls.WithCertificates(cert),
	)
	if err != nil {
		log.Println(err)
		return serv, err
	}

	serv.Listener = ln
	return serv, nil
}
