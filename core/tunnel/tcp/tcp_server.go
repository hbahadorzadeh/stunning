package tcp

import (
	"log"
	"net"

	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

type TcpServer struct {
	tcommon.TunnelServerCommon
}

func StartTcpServer(address string) (*TcpServer, error) {
	serv := &TcpServer{}
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return serv, err
	}
	serv.Listener = ln
	return serv, nil
}
