// Package ws provides WebSocket tunnel implementation.
package ws

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type WsConn struct {
	conn    *websocket.Conn
	readBuf []byte
	readPos int
}

func (w *WsConn) Read(b []byte) (n int, err error) {
	if w.readPos < len(w.readBuf) {
		n = copy(b, w.readBuf[w.readPos:])
		w.readPos += n
		if w.readPos >= len(w.readBuf) {
			w.readBuf = w.readBuf[:0]
			w.readPos = 0
		}
		return n, nil
	}

	_, data, err := w.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	w.readBuf = data
	w.readPos = 0
	n = copy(b, w.readBuf)
	w.readPos = n
	return n, nil
}

func (w *WsConn) Write(b []byte) (n int, err error) {
	err = w.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *WsConn) Close() error {
	return w.conn.Close()
}

func (w *WsConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WsConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (*WsConn) SetDeadline(t time.Time) error {
	return nil
}

func (*WsConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (*WsConn) SetWriteDeadline(t time.Time) error {
	return nil
}
