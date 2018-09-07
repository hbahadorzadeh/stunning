package udp

import (
	"encoding/binary"
	"fmt"
	"github.com/hbahadorzadeh/stunning/common"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	"log"
	"net"
	"sync"
	"time"
)

type udp_client struct {
	icommon.TunnelInterfaceClient
	address  string
	conn     *net.UDPConn
	replyMap []*common.UdpAddress
	mux      sync.Mutex
	closed   bool
}

func GetUdpClient(url string) *udp_client {
	s := &udp_client{}
	s.address = url
	udpAddr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	s.conn = conn
	s.replyMap = make([]*common.UdpAddress, 0)
	go func() {
		s.mux.Lock()
		for range time.Tick(30 * time.Second) {
			for i, conn := range s.replyMap {
				if conn.IsTimedOut() {
					s.replyMap[i] = nil
				}
			}
		}
		s.mux.Unlock()
	}()
	s.closed = false
	return s
}

func (c *udp_client) Close() {
	c.closed = true
}

func (c udp_client) Closed() bool {
	return c.closed
}

func (c *udp_client) WaitingForConnection() {

}

func (c *udp_client) HandleConnection(conn net.Conn) error {
	log.Printf("Socket to %s handling connection \n", c.address)
	go c.udp_client_reader(conn)
	c.udp_client_writer(conn)
	return nil
}

func (c *udp_client) udp_client_reader(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		addr := c.replyMap[int(binary.BigEndian.Uint64(buff[:AddrLenght]))]
		buff = buff[AddrLenght:n]
		wn, werr := c.conn.WriteToUDP(buff, addr.GetAddress())
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", addr.GetAddress().String(), wn)
	}
}

func (c *udp_client) udp_client_writer(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, addr, err := c.conn.ReadFromUDP(buff)
		if err != nil {
			log.Fatal(err)
		}
		uaddr := common.GetUdpAddress(addr)
		c.mux.Lock()
		i := arrayIndex(c.replyMap, uaddr)
		if i == -1 {
			c.replyMap = append(c.replyMap, uaddr)
			i = len(c.replyMap) - 1
		}
		c.mux.Unlock()
		addrb := append([]byte{}, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
		binary.BigEndian.PutUint64(addrb, uint64(i))
		buff = append(addrb, buff[:n]...)
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}

func arrayIndex(arr []*common.UdpAddress, search *common.UdpAddress) int {
	for i, v := range arr {
		if v.Equals(search) {
			return i
		}
	}
	return -1
}
