package stunning_test

import (
	"fmt"
	"golang.org/x/net/proxy"
	"hbx.ir/stunning/lib/net"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestHttpGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts := net.StartTlsServer("server.crt", "server.key", ":4443")
	ts.SetSocksServer(net.GetSocksServer())
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, net.GetTlsDialer())
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
