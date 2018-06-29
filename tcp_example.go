package stunning

import (
	"hbx.ir/stunning/interface/tcp"
	tlstun "hbx.ir/stunning/tunnel/tls"
	"log"
	"os"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := tlstun.StartTlsServer("server.crt", "server.key", "127.0.0.1:4443")
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:4443"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", tlstun.GetTlsDialer())
	time.Sleep(100 * time.Second)
}
