package test

import (
	"github.com/songgao/water"
	"hbx.ir/stunning/lib/net"
	"log"
	"os"
	"testing"
)

func TestCliTunGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	//ts := lib.StartTlsServer()
	//log.Println("Tls Server started")
	//stunconf :=lib.TunConfig{
	//	DevType: water.TUN,
	//	Address : "10.0.5.1",
	//	Name: "",
	//	MTU: "1500",
	//}
	//tunserv := lib.GetTunIface(stunconf)
	//log.Println("Tun Interface is up")
	//ts.SetTunServer(tunserv)
	//log.Println("Tun Interface is set to Tls Server")

	stunclientconf := net.TunConfig{
		DevType: water.TUN,
		Address: "10.0.5.2/24",
		Name:    "",
		MTU:     "1500",
	}
	tuncli := net.GetTunIface(stunclientconf)

	log.Println("Tun Interface for client is up")

	tc := net.GetTlsDialer()

	log.Println("Tls client started")
	conn, err := tc.Dial("tcp", "10.0.2.2:4443")
	if err != nil {
		panic(err)
	}
	log.Println("Tls client connected to server")
	tuncli.HandleConnection(conn)
	log.Println("Tls client set to client tun interface")
}
