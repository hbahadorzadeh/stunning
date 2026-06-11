package common

import (
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/core/interface/common"
	"github.com/hbahadorzadeh/stunning/core/plugin"
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
	KnockSpec  string
	knocker    plugin.Knocker
}

// SetKnockSpec configures port-knocking: the server only accepts connections from
// source IPs that have recently sent a valid knock. An empty spec disables it.
func (s *TunnelServerCommon) SetKnockSpec(spec string) {
	s.KnockSpec = spec
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
	if s.KnockSpec != "" {
		k, err := plugin.ParseKnock(s.KnockSpec)
		if err != nil {
			log.Panicf("knock setup failed: %v", err)
		}
		if err := k.Start(); err != nil {
			log.Panicf("knock listener failed: %v", err)
		}
		s.knocker = k
		defer k.Close()
		log.Printf("port-knock gate active (%s)\n", s.KnockSpec)
	}
	log.Printf("listening for connection on %s\n", s.Listener.Addr().String())
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		if s.knocker != nil {
			host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			if !s.knocker.Authorized(host) {
				log.Printf("dropping unauthorized (un-knocked) connection from %s", host)
				conn.Close()
				continue
			}
		}
		log.Println("new connection")
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
	wrapped, err := wrapServerConn(conn, s.PluginSpec, s.AuthSpec)
	if err != nil {
		log.Printf("plugin chain setup failed: %v", err)
		conn.Close()
		return
	}
	// Close the wrapper, not the raw conn, so the plugin chain and chaff
	// goroutine are also torn down.
	defer wrapped.Close()
	s.Server.HandleConnection(wrapped)
}
