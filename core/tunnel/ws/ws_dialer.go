// Package ws provides WebSocket tunnel dialer.
package ws

import (
	"crypto/tls"
	"net"

	"github.com/gorilla/websocket"
	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
)

type WsDialer struct {
	tcommon.TunnelDialer
}

func GetWsDialer() WsDialer {
	return WsDialer{}
}

func (WsDialer) Dial(network, addr string) (c net.Conn, err error) {
	wsURL := "wss://" + addr + "/"
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}
	return &WsConn{conn: conn}, nil
}

func (WsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Ws
}
