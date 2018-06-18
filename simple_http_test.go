package stunning

import (
	"golang.org/x/net/proxy"
	"hbx.ir/stunning/lib"
	"net/http"
	"testing"
	"log"
	"os"
	"io/ioutil"
	"fmt"
	"net"
)

func TestGet(t *testing.T) {
	log.SetOutput(os.Stderr)
	lib.StartTlsServer()
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, lib.GetTlsDialer())
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
