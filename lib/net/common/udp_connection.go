package common

import (
	"net"
	"time"
)

type UdpConnection struct {
	net.Conn
	addr net.Addr
}

func (u UdpConnection) Read(b []byte) (n int, err error){}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u UdpConnection) Write(b []byte) (n int, err error){
	return nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u UdpConnection) Close() error{
	return nil
}

// LocalAddr returns the local network address.
func (u UdpConnection) LocalAddr() Addr{
	return nil
}

// RemoteAddr returns the remote network address.
func (u UdpConnection) RemoteAddr() Addr{
	return nil
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
func (u UdpConnection) SetDeadline(t time.Time) error{
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u UdpConnection) SetReadDeadline(t time.Time) error{
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u UdpConnection) SetWriteDeadline(t time.Time) error{
	return nil
}