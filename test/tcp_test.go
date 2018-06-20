package test

import (
	"hbx.ir/stunning/lib/net"
	"hbx.ir/stunning/lib/utils"
	"log"
	"os"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := net.StartTlsServer(utils.ReadConf().GetString("application.crt"), utils.ReadConf().GetString("application.key"), utils.ReadConf().GetString("tcp_srv_test.listen"))
	ts.SetTcpServer(net.GetTcpServer(utils.ReadConf().GetString("tcp_srv_test.connect")))

	log.Printf("%v", utils.ReadConf().AllKeys())
	net.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", net.GetTlsDialer())
	time.Sleep(100 * time.Second)
}
