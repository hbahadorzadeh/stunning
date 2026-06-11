package common

import (
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/core/interface/common"
)

type TunnelServer interface {
	SetServer(server icommon.TunnelInterfaceServer)
	WaitingForConnection()
	Close() error
	Closed() bool
	HandleConnection(conn net.Conn)
}

type TunnelServerCommon struct {
	TunnelServer
	closed     bool
	Server     icommon.TunnelInterfaceServer
	Listener   net.Listener
	PluginSpec string
	AuthSpec   string
}

// SetPluginSpec configures the per-connection plugin chain applied to accepted
// connections. An empty spec disables plugins.
func (s *TunnelServerCommon) SetPluginSpec(spec string) {
	s.PluginSpec = spec
}

// SetAuthSpec configures the server-side authentication handshake applied to
// accepted connections. An empty spec disables authentication.
func (s *TunnelServerCommon) SetAuthSpec(spec string) {
	s.AuthSpec = spec
}

func (s *TunnelServerCommon) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
}

func (s *TunnelServerCommon) WaitingForConnection() {
	s.closed = false
	log.Printf("listening for connection on %s\n", s.Listener.Addr().String())
	for {
		conn, err := s.Listener.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			break
		}
		go s.HandleConnection(conn)
	}
	s.closed = true
	log.Printf("Listening on %s stopped\n", s.Listener.Addr().String())
}

func (s *TunnelServerCommon) Close() error {
	log.Println("Closing connection")
	err := s.Listener.Close()
	s.closed = true
	return err
}

func (s *TunnelServerCommon) Closed() bool {
	return s.closed
}

func (s *TunnelServerCommon) HandleConnection(conn net.Conn) {
	defer conn.Close()
	wrapped, err := wrapServerConn(conn, s.PluginSpec, s.AuthSpec)
	if err != nil {
		log.Printf("plugin chain setup failed: %v", err)
		return
	}
	s.Server.HandleConnection(wrapped)
}
