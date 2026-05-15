// Package socks provides SOCKS5-based tunnel interfaces.
package socks

import (
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
)

type SocksClient struct {
	icommon.TunnelInterfaceClient
	address    string
	tun_dialer tcommon.TunnelDialer
	saddress   string
	listen     net.Listener
	closed     bool
}

func GetSocksClient(url, surl string, tun_dialer tcommon.TunnelDialer) *SocksClient {
	s := &SocksClient{}
	s.address = url
	s.saddress = surl
	s.tun_dialer = tun_dialer
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Panic(err)
	}
	s.listen = listen
	s.closed = false
	return s
}

func (t *SocksClient) WaitingForConnection() {
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		sconn, serr := t.tun_dialer.Dial(t.tun_dialer.Protocol().String(), t.saddress)
		if serr != nil {
			log.Printf("error dialing upstream: %v", serr)
			conn.Close()
			continue
		}
		go t.HandleConnection(conn, sconn)
	}
}

func (t *SocksClient) HandleConnection(conn net.Conn, tconn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go tcp_reader(conn, tconn)
	tcp_writer(conn, tconn)
	return nil
}

func (t *SocksClient) Close() {
	t.listen.Close()
	t.closed = true
}

func (t *SocksClient) Closed() bool {
	return t.closed
}

func tcp_reader(conn net.Conn, tconn net.Conn) {
	defer conn.Close()
	defer tconn.Close()
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Printf("Error reading from conn: %v", err)
			return
		}
		buff = buff[:n]
		wn, werr := tconn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Printf("Error writing to tconn: %v", werr)
			return
		}
		log.Printf("%s : %d bytes wrote to socket", tconn.RemoteAddr().String(), wn)
	}
}

func tcp_writer(conn net.Conn, tconn net.Conn) {
	defer conn.Close()
	defer tconn.Close()
	for {
		buff := make([]byte, 1024)
		n, err := tconn.Read(buff)
		if err != nil {
			log.Printf("Error reading from tconn: %v", err)
			return
		}
		buff = buff[:n]
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Printf("Error writing to conn: %v", werr)
			return
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}
