package socks

import (
	icommon "hbx.ir/stunning/interface/common"
	tcommon "hbx.ir/stunning/tunnel/common"
	"log"
	"net"
)

type SocksClient struct {
	icommon.TunnelInterfaceClient
	address    string
	tun_dialer tcommon.TunnelDialer
	saddress   string
	listen     net.Listener
}

func GetSocksClient(url, surl string, tls_dialer tcommon.TunnelDialer) *SocksClient {
	s := &SocksClient{}
	s.address = url
	s.saddress = surl
	s.tun_dialer = tls_dialer
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Panic(err)
	}
	s.listen = listen
	go s.waiting_for_connection()
	return s
}

func (t *SocksClient) waiting_for_connection() {
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		sconn, serr := t.tun_dialer.Dial("tcp", t.saddress)
		if serr != nil {
			log.Fatalln(serr)
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

func tcp_reader(conn net.Conn, tconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = buff[:n]
		wn, werr := tconn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", tconn.RemoteAddr().String(), wn)
	}
}

func tcp_writer(conn net.Conn, tconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := tconn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		buff = buff[:n]
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}
