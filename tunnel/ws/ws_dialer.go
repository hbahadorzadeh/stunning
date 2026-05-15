package ws

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net"
)

type WsDialer struct {
	tcommon.TunnelDialer
}

func GetWsDialer() WsDialer {
	return WsDialer{}
}

func (d WsDialer) Dial(network, addr string) (c net.Conn, err error) {
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

func (d WsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Ws
}
