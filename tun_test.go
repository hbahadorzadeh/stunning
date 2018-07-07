package stunning

import (
	"fmt"
	"github.com/songgao/water"
	"gitlab.com/h.bahadorzadeh/stunning/interface/tun"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	"log"
	"os"
	"testing"
	"time"
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

func TestSrvGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts, err := tlstun.StartTlsServer("../server.crt", "../server.key", ":4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	fmt.Println("Tls Server started")
	stunconf := tun.TunConfig{
		DevType: water.TUN,
		Address: "10.0.5.1/24",
		Name:    "",
		MTU:     "1500",
	}
	tunserv := tun.GetTunIface(stunconf)
	fmt.Println("Tun Interface is up")
	ts.SetServer(tunserv)
	fmt.Println("Tun Interface is set to Tls Server")
	for {
		time.Sleep(1 * time.Second)
	}
}
