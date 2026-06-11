package tcp

import (
	"io"
	"log"
	"net"

	icommon "github.com/hbahadorzadeh/stunning/core/interface/common"
)

// copyBufSize is the relay copy-buffer size. Larger than the legacy 1 KiB so a
// stream is carried in fewer, bigger tunnel frames -- far fewer plugin-chain
// invocations, syscalls, and allocations per megabyte.
const copyBufSize = 32 * 1024

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
	buf := make([]byte, copyBufSize)
	if _, err := io.CopyBuffer(tconn, conn, buf); err != nil {
		log.Printf("relay conn->tunnel closed: %v", err)
	}
}

func tcp_writer(conn net.Conn, tconn net.Conn) {
	defer conn.Close()
	defer tconn.Close()
	buf := make([]byte, copyBufSize)
	if _, err := io.CopyBuffer(conn, tconn, buf); err != nil {
		log.Printf("relay tunnel->conn closed: %v", err)
	}
}
