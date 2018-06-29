package stunning

import (
	"hbx.ir/stunning/interface/tcp"
	tlstun "hbx.ir/stunning/tunnel/tls"
	"hbx.ir/stunning/lib/utils"
	"log"
	"os"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := tlstun.StartTlsServer(utils.ReadConf().GetString("application.crt"), utils.ReadConf().GetString("application.key"), utils.ReadConf().GetString("tcp_srv_test.listen"))
	ts.SetServer(tcp.GetTcpServer(utils.ReadConf().GetString("tcp_srv_test.connect")))

	log.Printf("%v", utils.ReadConf().AllKeys())
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", tlstun.GetTlsDialer())
	time.Sleep(100 * time.Second)
}
