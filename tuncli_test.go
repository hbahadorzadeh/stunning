package stunning


import (
	"hbx.ir/stunning/lib"
	"testing"
	"log"
	"os"
	"github.com/songgao/water"
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

	stunclientconf :=lib.TunConfig{
		DevType: water.TUN,
		Address : "10.0.5.2",
		Name: "",
		MTU: "1500",
	}
	tuncli := lib.GetTunIface(stunclientconf)

	log.Println("Tun Interface for client is up")

	tc := lib.GetTlsDialer()

	log.Println("Tls client started")
	conn, err := tc.Dial("tcp", "10.0.2.2:4443")
	if err != nil {
		panic(err)
	}
	log.Println("Tls client connected to server")
	tuncli.HandleConnection(conn)
	log.Println("Tls client set to client tun interface")
}
