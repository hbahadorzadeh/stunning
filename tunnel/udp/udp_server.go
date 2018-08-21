package udp

import (
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
	"sync"
	"time"
)

type UdpServer struct {
	tcommon.TunnelServerCommon
	conn    *net.UDPConn
	mux     *sync.Mutex
	wch     chan tcommon.UdpPacket
	connMap map[string]*tcommon.ServerUdpConnection
}

func StartUdpServer(address string) (UdpServer, error) {
	serv := UdpServer{}
	serv.connMap = make(map[string]*tcommon.ServerUdpConnection)
	serv.mux = &sync.Mutex{}
	serv.wch = make(chan tcommon.UdpPacket)
	go func() {
		for range time.Tick(30 * time.Second) {
			serv.mux.Lock()
			toBeDeleted := make([]string, 0)
			for i, conn := range serv.connMap {
				if conn.IsClosed() {
					toBeDeleted = append(toBeDeleted, i)
				}
			}

			for _, i := range toBeDeleted {
				delete(serv.connMap, i)
			}
			serv.mux.Unlock()
		}
	}()
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return serv, err
	}
	sconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return serv, err
	}
	serv.conn = sconn
	return serv, nil
}

func (s UdpServer) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
	go s.handleWriting()
	go s.WaitingForConnection()
}

func (s UdpServer) WaitingForConnection() {
	for {
		buff := make([]byte, 1024)
		n, addr, err := s.conn.ReadFromUDP(buff)
		log.Printf("Accepting connection from %s(%d bytes [%v] read)(error: %v)", addr.String(), n, buff[:n], err)
		if err == nil {
			log.Printf("getting connection from map")
			s.mux.Lock()
			conn, exists := s.connMap[addr.String()]
			s.mux.Unlock()
			log.Printf("connection existance : %t", exists)
			if !exists {
				ch := make(chan []byte)
				conn = &tcommon.ServerUdpConnection{
					Connection: s.conn,
					Addr:       addr,
					RCh:        ch,
					WCh:        s.wch,
					Closed:     false,
				}
				s.mux.Lock()
				s.connMap[addr.String()] = conn
				s.mux.Unlock()
				log.Printf("Accepted connection from %s", addr.String())
				go s.HandleConnection(conn)
			}
			conn.RCh <- buff[:n]
		}
	}
}

func (s UdpServer) handleWriting() {
	for upack := range s.wch {
		n, err := s.conn.WriteToUDP(upack.Buffer, upack.Addr)
		if err != nil {
			log.Printf("Error in writing bytes(%v) to %s : %v", upack.Buffer, upack.Addr.String(), err)
		}
		if n != len(upack.Buffer) {
			log.Printf("Error in writing bytes(%v) to %s : %d bytes writen insted of %d", upack.Buffer, upack.Addr.String(), n, len(upack.Buffer))
		}
	}
}

func (s UdpServer) Close() error {
	return s.conn.Close()
}
