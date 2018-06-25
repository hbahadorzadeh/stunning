package udp

import (
	"fmt"
	"hbx.ir/stunning/lib/utils"
	"log"
	"net"
	"strconv"
	icommon "hbx.ir/stunning/lib/net/interface/common"
)

type udp_client struct {
	icommon.TunnelInterfaceClient
	address  string
	conn     *net.UDPConn
	replyMap []string
}

func GetUdpClient(url string, conn net.Conn) *udp_client {
	s := &udp_client{}
	s.address = url
	udpAddr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	sconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Wrong Address")
		return nil
	}
	s.conn = sconn
	s.replyMap = make([]string, 0)
	return s
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
		//seq := binary.BigEndian.Uint16(strconv.buff[:10])
		buff = buff[10:n]
		wn, werr := c.conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", c.conn.RemoteAddr().String(), wn)
	}
}

func (c *udp_client) udp_client_writer(conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, addr, err := c.conn.ReadFrom(buff)
		if err != nil {
			log.Fatal(err)
		}

		i := utils.ArrayIndex(c.replyMap, addr.String())

		if i == -1 {
			c.replyMap = append(c.replyMap, addr.String())
			i = len(c.replyMap) - 1
		}

		buff = append(fillIntBytes([]byte(strconv.Itoa(i))), buff[:n]...)
		wn, werr := conn.Write(buff)
		if werr != nil || wn != len(buff) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%s : %d bytes wrote to socket", conn.RemoteAddr().String(), wn)
	}
}

func fillIntBytes(in []byte) []byte {
	for len(in) <= 10 {
		z := make([]byte, 1)
		z[0] = byte(0)
		in = append(z, in...)
	}
	return in
}
