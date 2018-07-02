package tcp

import (
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
)

type TcpServer struct {
	tcommon.TunnelServer
}

func StartTcpServer(address string) (TcpServer, error) {
	serv := TcpServer{}
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return serv, err
	}
	serv.Listener = ln
	return serv, nil
}
