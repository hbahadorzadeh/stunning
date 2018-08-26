package stunning

import (
	"fmt"
	"github.com/hbahadorzadeh/stunning/interface/tcp"
	httptun "github.com/hbahadorzadeh/stunning/tunnel/http"
	httpstun "github.com/hbahadorzadeh/stunning/tunnel/https"
	tcptun "github.com/hbahadorzadeh/stunning/tunnel/tcp"
	tlstun "github.com/hbahadorzadeh/stunning/tunnel/tls"
	udptun "github.com/hbahadorzadeh/stunning/tunnel/udp"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func TestTcpOverTls(t *testing.T) {
	log.SetOutput(os.Stderr)
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
		assertTEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertTEqualInt(t, bufflen, wn, "")
		time.Sleep(5 * time.Second)
	}()

	ts, err := tlstun.StartTlsServer("server.crt", "server.key", "127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:8888"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", tlstun.GetTlsDialer())

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertTEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertTEqualByteArray(t, rbuff, testBuff, "")
	defer recover()
}

func TestTcpOverTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
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
		assertTEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertTEqualInt(t, bufflen, wn, "")
	}()

	ts, err := tcptun.StartTcpServer("127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:8888"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", tcptun.GetTcpDialer())

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertTEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertTEqualByteArray(t, rbuff, testBuff, "")
	defer recover()
}

func TestTcpOverUdp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
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
		assertTEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertTEqualInt(t, bufflen, wn, "")
	}()

	ts, err := udptun.StartUdpServer("127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:8888"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", udptun.GetUdpDialer())

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertTEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertTEqualByteArray(t, rbuff, testBuff, "")
	defer recover()
}
func TestTcpOverHttp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
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
		assertTEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertTEqualInt(t, bufflen, wn, "")
	}()
	time.Sleep(time.Second)
	ts, err := httptun.StartHttpServer("127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:8888"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", httptun.GetHttpDialer())

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		time.Sleep(time.Second)
		conn, err = net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			t.Fatal(err)
		}
	}
	defer conn.Close()
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertTEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertTEqualByteArray(t, rbuff, testBuff, "")
	defer recover()
}
func TestTcpOverHttps(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10 * time.Second)
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
		assertTEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertTEqualInt(t, bufflen, wn, "")
	}()
	time.Sleep(time.Second)
	ts, err := httpstun.StartHttpsServer("server.crt", "server.key", "127.0.0.1:4443")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	ts.SetServer(tcp.GetTcpServer("127.0.0.1:8888"))
	tcp.GetTcpClient("127.0.0.1:8080", "127.0.0.1:4443", httpstun.GetHttpsDialer())

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		time.Sleep(time.Second)
		conn, err = net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			t.Fatal(err)
		}
	}
	defer conn.Close()
	log.Println("Writing through proxy")
	wn, werr := conn.Write(testBuff)
	if werr != nil {
		t.Fatal(werr)
	}
	bufflen := len(testBuff)
	assertTEqualInt(t, bufflen, wn, "")
	buff := make([]byte, 1024)
	log.Println("Read from proxy")
	n, rerr := conn.Read(buff)
	if rerr != nil {
		t.Fatal(rerr)
	}
	rbuff := buff[:n]
	assertTEqualByteArray(t, rbuff, testBuff, "")
	defer recover()
}

func assertTEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertTEqualInt(t *testing.T, a int, b int, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertTEqualByteArray(t *testing.T, a []byte, b []byte, message string) {
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
