package test

import (
	"fmt"
	"github.com/songgao/water"
	"hbx.ir/stunning/lib/net"
	"log"
	"os"
	"testing"
	"time"
)

func TestSrvGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := net.StartTlsServer("server.crt", "server.key", ":4443")
	fmt.Println("Tls Server started")
	stunconf := net.TunConfig{
		DevType: water.TUN,
		Address: "10.0.5.1",
		Name:    "",
		MTU:     "1500",
	}
	tunserv := net.GetTunIface(stunconf)
	fmt.Println("Tun Interface is up")
	ts.SetTunServer(tunserv)
	fmt.Println("Tun Interface is set to Tls Server")

	//stunclientconf :=lib.TunConfig{
	//	DevType: water.TUN,
	//	Address : "10.0.5.2",
	//	Name: "",
	//	MTU: "1500",
	//}
	//tuncli := lib.GetTunIface(stunclientconf)
	//
	//log.Println("Tun Interface for client is up")
	//
	//tc := lib.GetTlsDialer()
	//
	//log.Println("Tls client started")
	//conn, err := tc.Dial("tcp", "127.0.0.1:4443")
	//if err != nil {
	//	panic(err)
	//}
	//log.Println("Tls client connected to server")
	//tuncli.HandleConnection(conn)
	//log.Println("Tls client set to client tun interface")
	time.Sleep(100 * time.Second)
}
