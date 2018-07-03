package http

import (
	"net/http"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"fmt"
	"net"
	"log"
)


type TcpServer struct {
	tcommon.TunnelServer
}

func StartHttpServer(address string) (TcpServer, error) {

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {

	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":80", nil)
	serv := TcpServer{}
	return serv, nil
}
