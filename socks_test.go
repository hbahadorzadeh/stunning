package stunning

import (
	"fmt"
	"gitlab.com/h.bahadorzadeh/stunning/interface/socks"
	"gitlab.com/h.bahadorzadeh/stunning/tunnel/http"
	"gitlab.com/h.bahadorzadeh/stunning/tunnel/https"
	"gitlab.com/h.bahadorzadeh/stunning/tunnel/tcp"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	"gitlab.com/h.bahadorzadeh/stunning/tunnel/udp"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func TestSocksOverTls(t *testing.T) {
	log.SetOutput(os.Stderr)
	ts, err := tlstun.StartTlsServer("server.crt", "server.key", ":4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(socks.GetSocksServer())
	testBuff := append([]byte{}, []byte{0, 1, 2, 3, 4, 5, 6, 7}...)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		defer ln.Close()
		if err != nil {
			t.Fatal(err)
		}
		conn, cerr := ln.Accept()
		if cerr != nil {
			t.Fatal(cerr)
		}
		buff := make([]byte, 1024)
		n, rerr := conn.Read(buff)
		if rerr != nil {
			t.Fatal(rerr)
		}
		rbuff := buff[:n]
		assertEqualByteArray(t, rbuff, testBuff, "")
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
	}()

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, tlstun.GetTlsDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}

	log.Println("Connecting through proxy")
	conn, err := dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Reading from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertEqualByteArray(t, rbuff, testBuff, "")
}

func TestSocksOverTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
	ts, err := tcp.StartTcpServer(":4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(socks.GetSocksServer())
	testBuff := append([]byte{}, []byte{0, 1, 2, 3, 4, 5, 6, 7}...)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		defer ln.Close()
		if err != nil {
			t.Fatal(err)
		}
		conn, cerr := ln.Accept()
		if cerr != nil {
			t.Fatal(cerr)
		}
		buff := make([]byte, 1024)
		n, rerr := conn.Read(buff)
		if rerr != nil {
			t.Fatal(rerr)
		}
		rbuff := buff[:n]
		assertEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
	}()

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, tcp.GetTcpDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}
	log.Println("Connecting through proxy")
	conn, err := dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertEqualByteArray(t, rbuff, testBuff, "")
}

func TestSocksOverUdp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
	ts, err := udp.StartUdpServer(":4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(socks.GetSocksServer())
	testBuff := append([]byte{}, []byte{0, 1, 2, 3, 4, 5, 6, 7}...)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		defer ln.Close()
		if err != nil {
			t.Fatal(err)
		}
		conn, cerr := ln.Accept()
		log.Printf("TCP server: accepted connection from %s", conn.RemoteAddr().String())
		if cerr != nil {
			t.Fatal(cerr)
		}
		buff := make([]byte, 1024)
		n, rerr := conn.Read(buff)
		if rerr != nil {
			t.Fatal(rerr)
		}
		rbuff := buff[:n]
		assertEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
	}()

	dialSocksProxy, err := proxy.SOCKS5("udp", "127.0.0.1:4443", nil, udp.GetUdpDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}
	log.Println("Connecting through proxy")
	conn, err := dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertEqualByteArray(t, rbuff, testBuff, "")
}

func TestSocksOverHttp(t *testing.T) {
	log.SetOutput(os.Stderr)
	//time.Sleep(10*time.Second)
	ts, err := http.StartHttpServer("127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(socks.GetSocksServer())
	testBuff := append([]byte{}, []byte{0, 1, 2, 3, 4, 5, 6, 7}...)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		defer ln.Close()
		if err != nil {
			t.Fatal(err)
		}
		conn, cerr := ln.Accept()
		log.Printf("TCP server: accepted connection from %s", conn.RemoteAddr().String())
		if cerr != nil {
			t.Fatal(cerr)
		}
		buff := make([]byte, 1024)
		n, rerr := conn.Read(buff)
		if rerr != nil {
			t.Fatal(rerr)
		}
		rbuff := buff[:n]
		assertEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
	}()

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, http.GetHttpDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}
	log.Println("Connecting through proxy")
	conn, err := dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		time.Sleep(time.Second)
		conn, err = dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
		if err != nil {
			t.Fatal(err)
		}
	}
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertEqualByteArray(t, rbuff, testBuff, "")
}


func TestSocksOverHttps(t *testing.T) {
	log.SetOutput(os.Stderr)
	//time.Sleep(10*time.Second)
	ts, err := https.StartHttpsServer("server.crt", "server.key", "127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(socks.GetSocksServer())
	testBuff := append([]byte{}, []byte{0, 1, 2, 3, 4, 5, 6, 7}...)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		defer ln.Close()
		if err != nil {
			t.Fatal(err)
		}
		conn, cerr := ln.Accept()
		log.Printf("TCP server: accepted connection from %s", conn.RemoteAddr().String())
		if cerr != nil {
			t.Fatal(cerr)
		}
		buff := make([]byte, 1024)
		n, rerr := conn.Read(buff)
		if rerr != nil {
			t.Fatal(rerr)
		}
		rbuff := buff[:n]
		assertEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
	}()

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:4443", nil, https.GetHttpsDialer())
	if err != nil {
		log.Println("Error connecting to proxy:", err)
	}
	log.Println("Connecting through proxy")
	conn, err := dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		time.Sleep(time.Second)
		conn, err = dialSocksProxy.Dial("tcp", "127.0.0.1:8888")
		if err != nil {
			t.Fatal(err)
		}
	}
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertEqualByteArray(t, rbuff, testBuff, "")
}


func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertEqualInt(t *testing.T, a int, b int, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertEqualByteArray(t *testing.T, a []byte, b []byte, message string) {
	eq := true
	if len(a) == len(b) {
		for i := 0; i < len(a); i++ {
			eq = eq && a[i] == b[i]
			if !eq {
				break
			}
		}
		if eq {
			return
		}
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}
