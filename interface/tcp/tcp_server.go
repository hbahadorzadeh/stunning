package tcp

import (
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/interface/common"
)

type TcpServer struct {
	icommon.TunnelInterfaceServer
	address string
}

func GetTcpServer(url string) *TcpServer {
	s := &TcpServer{}
	s.address = url
	return s
}

func (*TcpServer) WaitingForConnection() {
	// No-op: WaitingForConnection is managed by the tunnel server
}

func (*TcpServer) Close() error {
	return nil
}

func (t *TcpServer) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", t.address)
	upconn, err := net.Dial("tcp", t.address)
	if err != nil {
		log.Printf("Failed to dial upstream %s: %v", t.address, err)
		conn.Close()
		return err
	}
	go tcp_reader(conn, upconn)
	tcp_writer(conn, upconn)
	return nil
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
		log.Printf("%d bytes wrote to socket", wn)
	}
}
