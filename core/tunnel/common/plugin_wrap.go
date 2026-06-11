package common

import (
	"net"

	"github.com/hbahadorzadeh/stunning/core/plugin"
)

// WrapDialer decorates a TunnelDialer so every dialed connection is wrapped in a
// per-connection plugin chain (client role) and then runs the client side of an
// optional authentication handshake. Empty specs are zero-cost: an empty plugin
// and auth spec returns the dialer unchanged.
//
// The auth handshake runs after the plugin framing so it is carried inside the
// obfuscated/disguised channel.
func WrapDialer(inner TunnelDialer, pluginSpec, authSpec, knockSpec string) TunnelDialer {
	if pluginSpec == "" && authSpec == "" && knockSpec == "" {
		return inner
	}
	return &pluginDialer{inner: inner, pluginSpec: pluginSpec, authSpec: authSpec, knockSpec: knockSpec}
}

type pluginDialer struct {
	inner      TunnelDialer
	pluginSpec string
	authSpec   string
	knockSpec  string
}

func (d *pluginDialer) Dial(network, addr string) (net.Conn, error) {
	// Port-knock before dialing so the server's tunnel port accepts us.
	if d.knockSpec != "" {
		k, err := plugin.ParseKnock(d.knockSpec)
		if err != nil {
			return nil, err
		}
		host, _, herr := net.SplitHostPort(addr)
		if herr != nil {
			host = addr
		}
		if err := k.Knock(host); err != nil {
			return nil, err
		}
	}
	conn, err := d.inner.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	if d.pluginSpec != "" {
		ch, err := plugin.ParseChain(d.pluginSpec)
		if err != nil {
			conn.Close()
			return nil, err
		}
		fc, err := plugin.NewFramedConn(conn, ch, true, 0)
		if err != nil {
			conn.Close()
			return nil, err
		}
		conn = fc
	}
	if d.authSpec != "" {
		auth, err := plugin.ParseAuth(d.authSpec)
		if err != nil {
			conn.Close()
			return nil, err
		}
		authed, err := auth.ClientHandshake(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
		conn = authed
	}
	return conn, nil
}

func (d *pluginDialer) Protocol() TunnelProtocol { return d.inner.Protocol() }

// wrapServerConn applies the server-side plugin chain then the server side of the
// optional auth handshake to an accepted connection. It returns the conn to use
// for subsequent data, or an error to drop the connection.
func wrapServerConn(conn net.Conn, pluginSpec, authSpec string) (net.Conn, error) {
	if pluginSpec != "" {
		ch, err := plugin.ParseChain(pluginSpec)
		if err != nil {
			return nil, err
		}
		fc, err := plugin.NewFramedConn(conn, ch, false, 0)
		if err != nil {
			return nil, err
		}
		conn = fc
	}
	if authSpec != "" {
		auth, err := plugin.ParseAuth(authSpec)
		if err != nil {
			return nil, err
		}
		authed, _, err := auth.ServerHandshake(conn)
		if err != nil {
			return nil, err
		}
		conn = authed
	}
	return conn, nil
}
