package udp

import (
	"hbx.ir/stunning/common"
	icommon "hbx.ir/stunning/interface/common"
	"log"
	"net"
	"sync"
	"time"
)

const AddrLenght int = 8

type UdpServer struct {
	icommon.TunnelInterfaceServer
	address string
	connMap map[[AddrLenght]byte]common.UdpConnection
	mux     sync.Mutex
}

func GetUdpServer(url string) UdpServer {
	s := UdpServer{}
	s.address = url
	s.connMap = make(map[[AddrLenght]byte]common.UdpConnection)
	go func() {
		s.mux.Lock()
		for range time.Tick(30 * time.Second) {
			toBeDeleted := make([][AddrLenght]byte, 0)
			for i, conn := range s.connMap {
				if conn.IsClosed() {
					toBeDeleted = append(toBeDeleted, i)
				}
			}

			for _, i := range toBeDeleted {
				delete(s.connMap, i)
			}
		}
		s.mux.Unlock()
	}()
	return s
}

func (s *UdpServer) getConnByAddr(buff []byte, ch chan []byte) net.Conn {
	var addr [8]byte
	for i := 0; i < AddrLenght; i++ {
		addr[i] = buff[i]
	}
	var conn common.UdpConnection
	var err bool
	s.mux.Lock()
	if conn, err = s.connMap[addr]; !err || conn.IsClosed() {
		var cerr error
		conn, cerr = common.GetUdpConnection(s.address, addr, ch)
		if cerr != nil {
			panic(cerr)
		}
	}
	s.mux.Unlock()
	return conn
}

func (s UdpServer) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", s.address)
	ch := make(chan []byte)
	go s.udp_server_reader(conn, ch)
	s.udp_server_writer(conn, ch)
	return nil
}

func (s *UdpServer) udp_server_reader(conn net.Conn, ch chan []byte) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		addr := buff[:AddrLenght]
		dconn := s.getConnByAddr(addr, ch)
		buff = buff[AddrLenght:n]
		wn, werr := dconn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", dconn.RemoteAddr().String(), wn)
	}
}

func (s *UdpServer) udp_server_writer(conn net.Conn, ch chan []byte) {
	for buff := range ch {
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, len(buff))
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}
