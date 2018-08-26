package tcp

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
)

type TcpClient struct {
	icommon.TunnelInterfaceClient
	address    string
	tun_dialer tcommon.TunnelDialer
	saddress   string
	listen     net.Listener
}

func GetTcpClient(url, surl string, tun_dialer tcommon.TunnelDialer) *TcpClient {
	s := &TcpClient{}
	s.address = url
	s.saddress = surl
	s.tun_dialer = tun_dialer
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Panic(err)
	}
	s.listen = listen
	return s
}

func (t *TcpClient) WaitingForConnection() {
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		sconn, serr := t.tun_dialer.Dial(t.tun_dialer.Protocol().String(), t.saddress)
		if serr != nil {
			log.Fatalln(serr)
			continue
		}
		go t.HandleConnection(conn, sconn)
	}
}

func (t *TcpClient) HandleConnection(conn net.Conn, tconn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go tcp_reader(conn, tconn)
	tcp_writer(conn, tconn)
	return nil
}
