package stunning

import (
	"fmt"
	"golang.org/x/net/proxy"
	"hbx.ir/stunning/interface/socks"
	tlstun "hbx.ir/stunning/tunnel/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestHttpGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := tlstun.StartTlsServer("../server.crt", "../server.key", ":4443")
	ts.SetServer(socks.GetSocksServer())
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, tlstun.GetTlsDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
	}
	res, rerr := myClient.Get("https://google.com")
	if rerr != nil {
		panic(rerr)
	}

	if res.StatusCode == 200 {
		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Panicln("error reading body:", err)
		}
		fmt.Println(string(b))
	}
}
