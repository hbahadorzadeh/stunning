package common

import (
	"net"

	"github.com/hbahadorzadeh/stunning/core/plugin"
)

// WrapDialer decorates a TunnelDialer so every dialed connection is wrapped in a
// per-connection plugin chain (client role). An empty spec returns the dialer
// unchanged, so plugins are zero-cost when unconfigured.
func WrapDialer(inner TunnelDialer, spec string) TunnelDialer {
	if spec == "" {
		return inner
	}
	return &pluginDialer{inner: inner, spec: spec}
}

type pluginDialer struct {
	inner TunnelDialer
	spec  string
}

func (d *pluginDialer) Dial(network, addr string) (net.Conn, error) {
	conn, err := d.inner.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	ch, err := plugin.ParseChain(d.spec)
	if err != nil {
		conn.Close()
		return nil, err
	}
	fc, err := plugin.NewFramedConn(conn, ch, true, 0)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return fc, nil
}

func (d *pluginDialer) Protocol() TunnelProtocol { return d.inner.Protocol() }

// wrapServerConn applies the server-side plugin chain to an accepted connection.
// It returns the original conn unchanged when spec is empty.
func wrapServerConn(conn net.Conn, spec string) (net.Conn, error) {
	if spec == "" {
		return conn, nil
	}
	ch, err := plugin.ParseChain(spec)
	if err != nil {
		return nil, err
	}
	return plugin.NewFramedConn(conn, ch, false, 0)
}
