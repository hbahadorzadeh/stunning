package test

import (
	"fmt"
	"github.com/songgao/water"
	"hbx.ir/stunning/lib/net/interface/tun"
	tlstun "hbx.ir/stunning/lib/net/tunnel/tls"
	"log"
	"os"
	"testing"
	"time"
)

func TestSrvGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := tlstun.StartTlsServer("../server.crt", "../server.key", ":4443")
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
