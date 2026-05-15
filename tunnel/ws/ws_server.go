package ws

import (
	"github.com/gorilla/websocket"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"log"
	"net/http"
)

type WsServer struct {
	tcommon.TunnelServerCommon
	upgrader    websocket.Upgrader
	webserver   *http.Server
	mux_handler *http.ServeMux
	crt, key    string
}

func StartWsServer(crt, key, address string) (*WsServer, error) {
	serv := &WsServer{
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		webserver:   &http.Server{Addr: address},
		mux_handler: http.NewServeMux(),
		crt:         crt,
		key:         key,
	}
	return serv, nil
}

func (s *WsServer) SetServer(server icommon.TunnelInterfaceServer) {
	s.Server = server
	go s.WaitingForConnection()
}

func (s *WsServer) WaitingForConnection() {
	s.mux_handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		wsConn := &WsConn{conn: conn}
		go s.HandleConnection(wsConn)
	})

	s.webserver.Handler = s.mux_handler
	if err := s.webserver.ListenAndServeTLS(s.crt, s.key); err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
}

func (s *WsServer) Close() error {
	return s.webserver.Close()
}
