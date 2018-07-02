package stunning

import (
	"gitlab.com/h.bahadorzadeh/stunning/interface/tcp"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	tcptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tcp"
	udptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/udp"
	"log"
	"os"
	"testing"
	"net"
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
		assertEqualByteArray(t, rbuff, testBuff, "")
		log.Printf("%v = %v", rbuff, testBuff)
		wn, werr := conn.Write(testBuff)
		if werr != nil {
			t.Fatal(werr)
		}
		log.Printf("%d bytes [%v] written", wn, testBuff)
		bufflen := len(testBuff)
		assertEqualInt(t, bufflen, wn, "")
		time.Sleep(5*time.Second)
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

func TestTcpOverTcp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10*time.Second)
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

func TestTcpOverUdp(t *testing.T) {
	log.SetOutput(os.Stderr)
	time.Sleep(10*time.Second)
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
