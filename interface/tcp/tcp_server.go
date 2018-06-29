package tcp

import (
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	"log"
	"net"
)

type TcpServer struct {
	icommon.TunnelInterfaceServer
	address string
	conn    net.Conn
}

func GetTcpServer(url string) TcpServer {
	s := TcpServer{}
	s.address = url

	conn, err := net.Dial("tcp", s.address)

	if err != nil {
		panic(err)
	}

	s.conn = conn

	return s
}

func (t TcpServer) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	go tcp_reader(conn, t.conn)
	tcp_writer(conn, t.conn)
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
