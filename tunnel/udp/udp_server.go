package udp

import (
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"net"
	"time"
)

type UdpServer struct {
	tcommon.TunnelServer
	conn udp_connection
}

func StartUdpServer(address string) (UdpServer, error) {
	serv := UdpServer{}
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return serv, err
	}
	sconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return serv, err
	}
	serv.conn = udp_connection{
		conn: sconn,
		addr : udpAddr,
	}
	return serv, nil
}


func (s UdpServer) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
	go s.WaitingForConnection()
}

func (s UdpServer) WaitingForConnection() {
	go s.HandleConnection(s.conn)
}

func(s UdpServer)Close()error{
	return s.conn.Close()
}

type udp_connection struct {
	net.Conn
	conn *net.UDPConn
	addr *net.UDPAddr
}

func (u udp_connection) Read(b []byte) (n int, err error) {
	u.conn.ReadFromUDP(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (u udp_connection) Write(b []byte) (n int, err error) {
	return u.conn.WriteToUDP(b, u.addr)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (u udp_connection) Close() error {
	return u.Close()
}

// LocalAddr returns the local network address.
func (u udp_connection) LocalAddr() net.Addr {
	return u.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (u udp_connection) RemoteAddr() net.Addr {
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
func (u udp_connection) SetDeadline(t time.Time) error {
	return u.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (u udp_connection) SetReadDeadline(t time.Time) error {
	return u.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (u udp_connection) SetWriteDeadline(t time.Time) error {
	return u.conn.SetWriteDeadline(t)
}
