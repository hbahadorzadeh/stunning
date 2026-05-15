// Package icmp provides ICMP tunnel implementation.
package icmp

import (
	"golang.org/x/net/icmp"
	"net"
	"time"
)

type icmpPacket struct {
	buf  []byte
	addr net.Addr
}

type ServerIcmpConnection struct {
	conn   *icmp.PacketConn
	addr   net.Addr
	RCh    chan []byte
	WCh    chan icmpPacket
	closed bool
}

func (u *ServerIcmpConnection) Read(b []byte) (n int, err error) {
	data := <-u.RCh
	n = len(data)
	if n > len(b) {
		n = len(b)
	}
	copy(b, data[:n])
	return n, nil
}

func (u *ServerIcmpConnection) Write(b []byte) (n int, err error) {
	u.WCh <- icmpPacket{buf: b, addr: u.addr}
	return len(b), nil
}

func (u *ServerIcmpConnection) Close() error {
	u.closed = true
	return nil
}

func (u *ServerIcmpConnection) LocalAddr() net.Addr {
	return u.conn.LocalAddr()
}

func (u *ServerIcmpConnection) RemoteAddr() net.Addr {
	return u.addr
}

func (_ *ServerIcmpConnection) SetDeadline(t time.Time) error {
	return nil
}

func (_ *ServerIcmpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (_ *ServerIcmpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

type ClientIcmpConnection struct {
	conn net.Addr
	addr net.Addr
}

func (_ *ClientIcmpConnection) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (_ *ClientIcmpConnection) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (_ *ClientIcmpConnection) Close() error {
	return nil
}

func (u *ClientIcmpConnection) LocalAddr() net.Addr {
	return u.conn
}

func (u *ClientIcmpConnection) RemoteAddr() net.Addr {
	return u.addr
}

func (_ *ClientIcmpConnection) SetDeadline(t time.Time) error {
	return nil
}

func (_ *ClientIcmpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (_ *ClientIcmpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
