package stunning

import (
	"github.com/songgao/water"
	"hbx.ir/stunning/interface/tun"
	tlstun "hbx.ir/stunning/tunnel/tls"
	"log"
	"os"
	"testing"
)

func TestCliTunGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	stunclientconf := tun.TunConfig{
		DevType: water.TUN,
		Address: "10.0.5.2/24",
		Name:    "",
		MTU:     "1500",
	}
	tuncli := tun.GetTunIface(stunclientconf)

	log.Println("Tun Interface for client is up")

	tc := tlstun.GetTlsDialer()

	log.Println("Tls client started")
	conn, err := tc.Dial("tcp", "10.0.2.2:4443")
	if err != nil {
		panic(err)
	}
	log.Println("Tls client connected to server")
	tuncli.HandleConnection(conn)
	log.Println("Tls client set to client tun interface")
}
