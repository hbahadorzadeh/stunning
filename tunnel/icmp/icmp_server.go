package icmp

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"golang.org/x/net/icmp"
	"log"
	"sync"
)

type IcmpServer struct {
	tcommon.TunnelServerCommon
	conn                *icmp.PacketConn
	mux                 sync.Mutex
	wch                 chan icmpPacket
	connMap             map[string]*ServerIcmpConnection
	closeDone           chan struct{}
	writeGoroutineDone  chan struct{}
}

func StartIcmpServer(address string) (*IcmpServer, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", address)
	if err != nil {
		log.Printf("SECURITY ERROR: ICMP requires root or CAP_NET_RAW privileges. This is a critical security requirement: %v", err)
		return nil, err
	}

	return &IcmpServer{
		conn:                conn,
		mux:                 sync.Mutex{},
		wch:                 make(chan icmpPacket, 100),
		connMap:             make(map[string]*ServerIcmpConnection),
		closeDone:           make(chan struct{}),
		writeGoroutineDone:  make(chan struct{}),
	}, nil
}

func (s *IcmpServer) SetServer(server icommon.TunnelInterfaceServer) {
	s.Server = server
	go s.handleWriting()
	go s.WaitingForConnection()
}

func (s *IcmpServer) WaitingForConnection() {
	defer func() {
		s.mux.Lock()
		for _, conn := range s.connMap {
			close(conn.RCh)
		}
		s.mux.Unlock()
	}()

	buf := make([]byte, 65535)
	for {
		n, peer, err := s.conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			return
		}

		addr := peer.String()
		s.mux.Lock()
		conn, exists := s.connMap[addr]
		if !exists {
			conn = &ServerIcmpConnection{
				conn:   s.conn,
				addr:   peer,
				RCh:    make(chan []byte, 10),
				WCh:    s.wch,
				closed: false,
			}
			s.connMap[addr] = conn
		}
		s.mux.Unlock()

		if !exists {
			go s.HandleConnection(conn)
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		select {
		case conn.RCh <- data:
		default:
		}
	}
}

func (s *IcmpServer) handleWriting() {
	defer close(s.writeGoroutineDone)
	for {
		select {
		case packet, ok := <-s.wch:
			if !ok {
				return
			}
			s.conn.WriteTo(packet.buf, packet.addr)
		case <-s.closeDone:
			return
		}
	}
}

func (s *IcmpServer) Close() error {
	// Close the packet connection first to unblock ReadFrom
	err := s.conn.Close()

	// Signal writeGoroutine to exit and wait for it to finish
	close(s.closeDone)
	close(s.wch)

	// Wait for write goroutine to finish processing
	<-s.writeGoroutineDone

	// Close all connection channels
	s.mux.Lock()
	for _, conn := range s.connMap {
		close(conn.RCh)
	}
	s.mux.Unlock()

	return err
}
