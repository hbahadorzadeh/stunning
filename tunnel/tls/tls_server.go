package tls

import (
	"crypto/tls"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
)

type TlsServer struct {
	tcommon.TunnelServer
}

func StartTlsServer(crt, key, address string) (TlsServer, error) {
	serv := TlsServer{}
	cer, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		log.Println(err)
		return serv, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)

	if err != nil {
		log.Println(err)
		return serv, err
	}
	serv.Listener = ln
	return serv, nil
}
