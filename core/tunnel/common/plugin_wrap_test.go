package common

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hbahadorzadeh/stunning/core/plugin"
)

// rawDialer is a minimal TunnelDialer over plain TCP for exercising WrapDialer
// without importing a concrete tunnel package (which would cycle).
type rawDialer struct{}

func (rawDialer) Dial(network, addr string) (net.Conn, error) { return net.Dial("tcp", addr) }
func (rawDialer) Protocol() TunnelProtocol                    { return Tcp }

func TestWrapDialerServerRoundTrip(t *testing.T) {
	const spec = "flate,aead?key=0123456789abcdef,pad?min=8&max=64,probe-guard?key=feed"

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		raw, err := ln.Accept()
		if err != nil {
			return
		}
		conn, err := wrapServerConn(raw, spec, "") // server side of the chain
		if err != nil {
			return
		}
		io.Copy(conn, conn) // echo
	}()

	dialer := WrapDialer(rawDialer{}, spec, "", "") // client side of the chain
	conn, err := dialer.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	want := bytes.Repeat([]byte("wire-traffic-9876543210 "), 80)
	go func() { conn.Write(want) }()

	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("payload mismatch through wrapped dialer/server")
	}
}

func TestWrapDialerEmptySpecIsPassthrough(t *testing.T) {
	if got := WrapDialer(rawDialer{}, "", "", ""); got == nil {
		t.Fatal("empty spec should return the inner dialer unchanged")
	}
	if _, ok := WrapDialer(rawDialer{}, "", "", "").(rawDialer); !ok {
		t.Fatal("empty spec must not wrap the dialer")
	}
}

// TestWrapDialerWithAuth exercises the full client path: plugin chain framing +
// PSK auth handshake + data echo, plus rejection on a key mismatch.
func TestWrapDialerWithAuth(t *testing.T) {
	const chain = "aead?key=0123456789abcdef,tls-mimic"
	const goodAuth = "psk?key=cafebabe"

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			raw, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				conn, err := wrapServerConn(c, chain, goodAuth)
				if err != nil {
					c.Close()
					return
				}
				io.Copy(conn, conn)
			}(raw)
		}
	}()

	// Happy path: matching key authenticates and echoes.
	conn, err := WrapDialer(rawDialer{}, chain, goodAuth, "").Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("authenticated dial failed: %v", err)
	}
	defer conn.Close()
	want := bytes.Repeat([]byte("authenticated-traffic "), 64)
	go func() { conn.Write(want) }()
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("payload mismatch through auth+chain")
	}

	// Rejection: wrong PSK key must fail the dial.
	if _, err := WrapDialer(rawDialer{}, chain, "psk?key=deadbeef", "").Dial("tcp", ln.Addr().String()); err == nil {
		t.Fatal("dial with wrong PSK key must be rejected")
	}
}

// TestWrapDialerKnockGate verifies the full port-knock path: a knocking client is
// authorized and reaches the (gated) tunnel; a non-knocking connection is dropped.
func TestWrapDialerKnockGate(t *testing.T) {
	// free UDP port for the knock listener
	upc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, ps, _ := net.SplitHostPort(upc.LocalAddr().String())
	upc.Close()
	spec := fmt.Sprintf("spa?key=00112233&port=%s&ttl=3s&delay=120ms", ps)

	knocker, err := plugin.ParseKnock(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := knocker.Start(); err != nil {
		t.Fatal(err)
	}
	defer knocker.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if !knocker.Authorized(host) { // same gate as TunnelServerCommon
				c.Close()
				continue
			}
			go func(conn net.Conn) { io.Copy(conn, conn) }(c)
		}
	}()

	// Knocking client reaches the service.
	conn, err := WrapDialer(rawDialer{}, "", "", spec).Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("knocking dial failed: %v", err)
	}
	defer conn.Close()
	msg := []byte("through-the-knock-gate")
	go func() { conn.Write(msg) }()
	got := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatal("payload mismatch through knock gate")
	}

	// Non-knocking connection (from an un-authorized state) is dropped. Use a
	// fresh source by waiting out the TTL so 127.0.0.1 is no longer authorized.
	time.Sleep(3200 * time.Millisecond)
	raw, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	raw.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := raw.Read(make([]byte, 1)); err == nil {
		t.Fatal("un-knocked connection should be dropped")
	}
}
