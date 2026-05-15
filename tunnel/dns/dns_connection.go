package dns

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

type DnsConn struct {
	conn net.Conn
}

const maxDnsMessageSize = 65535 // Max DNS message size per RFC 1035

func (d *DnsConn) Read(b []byte) (n int, err error) {
	lenBuf := make([]byte, 2)
	_, err = io.ReadFull(d.conn, lenBuf)
	if err != nil {
		return 0, err
	}

	msgLen := binary.BigEndian.Uint16(lenBuf)

	// Validate message length to prevent DoS or buffer allocation attacks
	if msgLen > maxDnsMessageSize {
		return 0, io.EOF
	}

	if len(b) < int(msgLen) {
		msgLen = uint16(len(b))
	}

	return io.ReadFull(d.conn, b[:msgLen])
}

func (d *DnsConn) Write(b []byte) (n int, err error) {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(b)))

	_, err = d.conn.Write(lenBuf)
	if err != nil {
		return 0, err
	}

	return d.conn.Write(b)
}

func (d *DnsConn) Close() error {
	return d.conn.Close()
}

func (d *DnsConn) LocalAddr() net.Addr {
	return d.conn.LocalAddr()
}

func (d *DnsConn) RemoteAddr() net.Addr {
	return d.conn.RemoteAddr()
}

func (d *DnsConn) SetDeadline(t time.Time) error {
	return d.conn.SetDeadline(t)
}

func (d *DnsConn) SetReadDeadline(t time.Time) error {
	return d.conn.SetReadDeadline(t)
}

func (d *DnsConn) SetWriteDeadline(t time.Time) error {
	return d.conn.SetWriteDeadline(t)
}
