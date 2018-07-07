package example

import (
	"fmt"
	"gitlab.com/h.bahadorzadeh/stunning/interface/socks"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func socks_example() {
	log.SetOutput(os.Stderr)
	ts, err := tlstun.StartTlsServer("../server.crt", "../server.key", ":4443")
	if err != nil {
		log.Fatal(err)
	}
	defer ts.Close()
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
