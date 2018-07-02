package udp

import (
	"bufio"
	"log"
	"net"
	"time"
)

type udp_packet struct {
	buff []byte
	addr *net.UDPAddr
}

type serverUdpConnection struct {
	net.Conn
	conn   *net.UDPConn
	addr   *net.UDPAddr
	ch     chan []byte
	wch    chan udp_packet
	closed bool
}

func (u serverUdpConnection) IsClosed() bool {
	return u.closed
}

func (u serverUdpConnection) Read(b []byte) (n int, err error) {
	buff := <-u.ch
	n = len(buff)
	copy(b, buff)
	log.Printf("reading bytes(%v) %s -> %s", buff, u.addr.String(), u.conn.LocalAddr().String())
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u serverUdpConnection) Write(b []byte) (n int, err error) {
	log.Printf("writing bytes(%v) %s -> %s", b, u.conn.LocalAddr().String(), u.addr.String())
	u.wch <- udp_packet{
		buff: b,
		addr: u.addr,
	}
	return len(b), nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u serverUdpConnection) Close() error {
	u.closed = true
	return nil
}

// LocalAddr returns the local network address.
func (u serverUdpConnection) LocalAddr() net.Addr {
	return u.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u serverUdpConnection) RemoteAddr() net.Addr {
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
func (u serverUdpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u serverUdpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u serverUdpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

type clientUdpConnection struct {
	net.Conn
	conn   *net.UDPConn
	buff   []byte
	reader *bufio.Reader
}

func (u clientUdpConnection) Read(b []byte) (n int, err error) {
	n, err = u.reader.Read(b)
	log.Printf("reading bytes[%d](%v) %s -> %s", n, b, u.conn.RemoteAddr().String(), u.conn.LocalAddr().String())
	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u clientUdpConnection) Write(b []byte) (n int, err error) {
	log.Printf("writing bytes(%v) %s -> %s", b, u.conn.LocalAddr().String(), u.conn.RemoteAddr().String())
	return u.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u clientUdpConnection) Close() error {
	return u.conn.Close()
}

// LocalAddr returns the local network address.
func (u clientUdpConnection) LocalAddr() net.Addr {
	return u.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u clientUdpConnection) RemoteAddr() net.Addr {
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
func (u clientUdpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u clientUdpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u clientUdpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
