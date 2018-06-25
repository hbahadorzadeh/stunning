package udp

import (
	"log"
	"net"
	icommon "hbx.ir/stunning/lib/net/interface/common"
)

type UdpServer struct {
	icommon.TunnelInterfaceServer
	address string
	connMap map[[10]byte]net.Conn
}

func GetUdpServer(url string) UdpServer {
	s := UdpServer{}
	s.address = url

	s.connMap = make(map[[10]byte]net.Conn)

	return s
}

func (s *UdpServer) getConnByAddr(addr [10]byte) net.Conn {

	//conn, err := net.Dial("udp", s.address)
	_, err := net.Dial("udp", s.address)

	if err != nil {
		panic(err)
	}
}

func (s UdpServer) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", s.address)
	//go s.udp_server_reader(conn, s.connMap)
	//s.udp_server_writer(conn)
	return nil
}

func (s *UdpServer) udp_server_reader(conn, sconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		//addr := buff[:10]
		buff = buff[10:n]
		wn, werr := sconn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", sconn.RemoteAddr().String(), wn)
	}
}

func (s *UdpServer) udp_server_writer(conn, sconn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := sconn.Read(buff)
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
