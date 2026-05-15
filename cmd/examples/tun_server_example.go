package example

import (
	"fmt"
	"github.com/hbahadorzadeh/stunning/interface/tun"
	tlstun "github.com/hbahadorzadeh/stunning/tunnel/tls"
	"github.com/songgao/water"
	"log"
	"os"
	"time"
)

func tun_server_example() {
	log.SetOutput(os.Stderr)
	ts, err := tlstun.StartTlsServer("server.crt", "server.key", ":4443")
	if err != nil {
		log.Fatal(err)
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
