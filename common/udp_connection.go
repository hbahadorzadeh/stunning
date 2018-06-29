package common

import (
	"log"
	"net"
	"time"
)

type UdpConnection struct {
	net.Conn
	conn   net.Conn
	addr   [8]byte
	ch     chan []byte
	closed bool
}

func GetUdpConnection(raddress string, laddr [8]byte, ch chan []byte) (UdpConnection, error) {
	u := UdpConnection{}
	u.addr = laddr
	u.ch = ch
	conn, err := net.Dial("udp", raddress)
	if err != nil {
		return u, err
	}
	u.conn = conn
	u.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	u.closed = false
	go u.Reader()
	return u, nil
}

func (u UdpConnection) Reader() {
	for {
		buff := make([]byte, 1024)
		n, err := u.conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = append(append([]byte{}, u.addr[0], u.addr[1], u.addr[2], u.addr[3], u.addr[4], u.addr[5], u.addr[6], u.addr[7]), buff[:n]...)
		u.ch <- buff
		log.Printf("%v : %d bytes wrote to socket", u.addr, n)
	}
	u.conn.Close()
	u.closed = true
}

func (u UdpConnection) IsClosed() bool {
	return u.closed
}

func (u UdpConnection) Read(b []byte) (n int, err error) {
	return u.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u UdpConnection) Write(b []byte) (n int, err error) {
	return u.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u UdpConnection) Close() error {
	return u.conn.Close()
}

// LocalAddr returns the local network address.
func (u UdpConnection) LocalAddr() net.Addr {
	return u.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u UdpConnection) RemoteAddr() net.Addr {
	return u.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (u UdpConnection) SetDeadline(t time.Time) error {
	return u.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u UdpConnection) SetReadDeadline(t time.Time) error {
	return u.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u UdpConnection) SetWriteDeadline(t time.Time) error {
	return u.conn.SetWriteDeadline(t)
}
