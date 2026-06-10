package plugin

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

// framedPair wires two FramedConns over a net.Pipe with opposite roles.
func framedPair(t *testing.T, spec string) (client, server *FramedConn) {
	t.Helper()
	cc, sc := net.Pipe()
	clChain, _ := ParseChain(spec)
	svChain, _ := ParseChain(spec)
	client, err := NewFramedConn(cc, clChain, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	server, err = NewFramedConn(sc, svChain, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	return client, server
}

func TestFramedConnRoundTrip(t *testing.T) {
	specs := []string{
		"",
		"pad?min=8&max=64",
		"flate,aead?key=00ff,pad?min=4&max=40,probe-guard?key=11ee",
	}
	msgs := [][]byte{
		[]byte("first message"),
		bytes.Repeat([]byte("A"), 1500),
		[]byte("third"),
	}
	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			client, server := framedPair(t, spec)
			defer client.Close()
			defer server.Close()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, m := range msgs {
					if _, err := client.Write(m); err != nil {
						t.Errorf("write: %v", err)
						return
					}
				}
			}()

			for _, want := range msgs {
				got := make([]byte, len(want))
				if _, err := io.ReadFull(server, got); err != nil {
					t.Fatalf("read: %v", err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("mismatch: got %d bytes", len(got))
				}
			}
			wg.Wait()
		})
	}
}

func TestFramedConnBidirectional(t *testing.T) {
	spec := "aead?key=cafe,pad?min=0&max=32"
	client, server := framedPair(t, spec)
	defer client.Close()
	defer server.Close()

	// server echoes one message back to client
	go func() {
		buf := make([]byte, 64)
		n, err := server.Read(buf)
		if err != nil {
			return
		}
		_, _ = server.Write(buf[:n])
	}()

	ping := []byte("ping-pong")
	if _, err := client.Write(ping); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(ping))
	if _, err := io.ReadFull(client, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, ping) {
		t.Fatalf("bidirectional mismatch: %q", got)
	}
}

// TestFramedConnOverTCP is a real end-to-end check over a loopback TCP socket
// with the full recommended chain: client sends, server echoes, client verifies.
func TestFramedConnOverTCP(t *testing.T) {
	spec := "flate,aead?key=0123456789abcdef,pad?min=16&max=128,probe-guard?key=feedface"
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		raw, err := ln.Accept()
		if err != nil {
			return
		}
		ch, _ := ParseChain(spec)
		srv, err := NewFramedConn(raw, ch, false, 0)
		if err != nil {
			return
		}
		defer srv.Close()
		io.Copy(srv, srv) // echo decoded stream back, re-encoded
	}()

	raw, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ch, _ := ParseChain(spec)
	cli, err := NewFramedConn(raw, ch, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	want := bytes.Repeat([]byte("payload-through-the-firewall "), 50)
	go func() { cli.Write(want) }()

	got := make([]byte, len(want))
	if _, err := io.ReadFull(cli, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("tcp e2e payload mismatch")
	}
}
