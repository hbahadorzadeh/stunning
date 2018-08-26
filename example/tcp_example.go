package example

import (
	"github.com/hbahadorzadeh/stunning/interface/tcp"
	tlstun "github.com/hbahadorzadeh/stunning/tunnel/tls"
	"log"
	"os"
	"time"
)

func tcp_example() {
	log.SetOutput(os.Stderr)
	ts, err := tlstun.StartTlsServer("server.crt", "server.key", "127.0.0.1:4443")
	if err != nil {
		log.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:4443"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", tlstun.GetTlsDialer())
	time.Sleep(100 * time.Second)
}
