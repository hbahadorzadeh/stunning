package common

import (
	"bufio"
	"log"
	"net"
	"time"
)

type UdpPacket struct {
	Buffer []byte
	Addr   *net.UDPAddr
}

type ServerUdpConnection struct {
	net.Conn
	Connection *net.UDPConn
	Addr       *net.UDPAddr
	RCh        chan []byte
	WCh        chan UdpPacket
	Closed     bool
}

func (u ServerUdpConnection) IsClosed() bool {
	return u.Closed
}

func (u ServerUdpConnection) Read(b []byte) (n int, err error) {
	buff := <-u.RCh
	n = len(buff)
	copy(b, buff)
	log.Printf("reading bytes(%v) %s -> %s", buff, u.Addr.String(), u.Connection.LocalAddr().String())
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u ServerUdpConnection) Write(b []byte) (n int, err error) {
	log.Printf("writing bytes(%v) %s -> %s", b, u.Connection.LocalAddr().String(), u.Addr.String())
	u.WCh <- UdpPacket{
		Buffer: b,
		Addr:   u.Addr,
	}
	return len(b), nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u ServerUdpConnection) Close() error {
	u.Closed = true
	return nil
}

// LocalAddr returns the local network address.
func (u ServerUdpConnection) LocalAddr() net.Addr {
	return u.Connection.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u ServerUdpConnection) RemoteAddr() net.Addr {
	return u.Connection.RemoteAddr()
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
func (u ServerUdpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u ServerUdpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u ServerUdpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

type ClientUdpConnection struct {
	net.Conn
	Connection *net.UDPConn
	Buffer     []byte
	Reader     *bufio.Reader
}

func (u ClientUdpConnection) Read(b []byte) (n int, err error) {
	n, err = u.Reader.Read(b)
	log.Printf("reading bytes[%d](%v) %s -> %s", n, b, u.Connection.RemoteAddr().String(), u.Connection.LocalAddr().String())
	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u ClientUdpConnection) Write(b []byte) (n int, err error) {
	log.Printf("writing bytes(%v) %s -> %s", b, u.Connection.LocalAddr().String(), u.Connection.RemoteAddr().String())
	return u.Connection.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u ClientUdpConnection) Close() error {
	return u.Connection.Close()
}

// LocalAddr returns the local network address.
func (u ClientUdpConnection) LocalAddr() net.Addr {
	return u.Connection.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u ClientUdpConnection) RemoteAddr() net.Addr {
	return u.Connection.RemoteAddr()
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
func (u ClientUdpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u ClientUdpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u ClientUdpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
