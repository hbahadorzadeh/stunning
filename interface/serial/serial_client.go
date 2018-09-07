package serial

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	"net"
)

type SerialClient struct {
	icommon.TunnelInterfaceClient
	tun_dialer tcommon.TunnelDialer
	saddress   string
	port       io.ReadWriteCloser
	listen     net.Listener
}

func GetTcpClient(saddress, PortName string, BaudRate, DataBits, StopBits, MinimumReadSize uint, tun_dialer tcommon.TunnelDialer) SerialClient {
	s := SerialClient{}
	options := serial.OpenOptions{
		PortName:        PortName,
		BaudRate:        BaudRate,
		DataBits:        DataBits,
		StopBits:        StopBits,
		MinimumReadSize: MinimumReadSize,
	}
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	s.port = port
	s.saddress = saddress
	return s
}

func (s SerialClient) WaitingForConnection() {
	//for {
	//
	//	sconn, serr := s.tun_dialer.Dial(s.tun_dialer.Protocol().String(), s.saddress)
	//	if serr != nil {
	//		log.Fatalln(serr)
	//		continue
	//	}
	//	// Write 4 bytes to the port.
	//	b := []byte{0x00, 0x01, 0x02, 0x03}
	//	n, err := port.Write(b)
	//	if err != nil {
	//		log.Fatalf("port.Write: %v", err)
	//	}
	//}
}

func (s SerialClient) HandleConnection(conn net.Conn, tconn net.Conn) error {
	go tcp_reader(conn, tconn)
	tcp_writer(conn, tconn)
	return nil
}

func (s SerialClient) Close() {
	s.port.Close()
}
