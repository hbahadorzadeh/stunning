package plugin

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

// wireHead returns the first n bytes a client-side FramedConn puts on the wire
// for the given chain spec and message.
func wireHead(t *testing.T, spec string, msg []byte, n int) []byte {
	t.Helper()
	cc, sc := net.Pipe()
	ch, err := ParseChain(spec)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewFramedConn(cc, ch, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	go func() { client.Write(msg); client.Close() }()
	head := make([]byte, n)
	if _, err := io.ReadFull(sc, head); err != nil {
		t.Fatalf("read wire head: %v", err)
	}
	sc.Close()
	return head
}

func TestTLSMimicWireLooksLikeTLS(t *testing.T) {
	head := wireHead(t, "aead?key=00ff,tls-mimic", []byte("hello"), 6)
	// TLS handshake record: type=0x16, version 0x03 0x0x, body starts ClientHello 0x01.
	if head[0] != 0x16 || head[1] != 0x03 || head[5] != 0x01 {
		t.Fatalf("wire does not look like a TLS ClientHello: % x", head)
	}
}

func TestHTTPMimicWireLooksLikeHTTP(t *testing.T) {
	head := wireHead(t, "aead?key=00ff,http-mimic", []byte("hello"), 5)
	if !bytes.HasPrefix(head, []byte("POST ")) {
		t.Fatalf("wire does not look like HTTP: %q", head)
	}
}

func TestMimicRoundTrip(t *testing.T) {
	specs := []string{
		"tls-mimic",
		"http-mimic",
		"flate,aead?key=0123456789abcdef,tls-mimic",
		"flate,aead?key=0123456789abcdef,bucket?size=128,http-mimic",
		"aead?key=cafe,jitter?max=1ms,tls-mimic",
	}
	msgs := [][]byte{
		[]byte("first"),
		bytes.Repeat([]byte("B"), 5000),
		[]byte("last message through the disguise"),
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
					t.Fatalf("mismatch: want %d got %d bytes", len(want), len(got))
				}
			}
			wg.Wait()
		})
	}
}

func TestBucketRoundTripAndSizing(t *testing.T) {
	cl, _ := ParseChain("bucket?size=256")
	sv, _ := ParseChain("bucket?size=256")
	defer cl.Close()
	defer sv.Close()
	for _, in := range payloads() {
		enc, err := cl.Encode(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(enc)%256 != 0 {
			t.Fatalf("bucket output %d not a multiple of 256", len(enc))
		}
		out, err := sv.Decode(enc)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, in) {
			t.Fatal("bucket round-trip mismatch")
		}
	}
}

func TestJitterRoundTrip(t *testing.T) {
	cl, _ := ParseChain("jitter?min=0&max=1ms")
	sv, _ := ParseChain("jitter?min=0&max=1ms")
	in := []byte("timing")
	enc, _ := cl.Encode(in)
	out, _ := sv.Decode(enc)
	if !bytes.Equal(out, in) {
		t.Fatal("jitter must not alter bytes")
	}
}

func TestMimicOverTCP(t *testing.T) {
	spec := "flate,aead?key=0123456789abcdef,bucket?size=512,tls-mimic"
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
		ch, _ := ParseChain(spec)
		srv, err := NewFramedConn(raw, ch, false, 0)
		if err != nil {
			return
		}
		defer srv.Close()
		io.Copy(srv, srv)
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
	want := bytes.Repeat([]byte("disguised-as-tls-0123456789 "), 60)
	go func() { cli.Write(want) }()
	got := make([]byte, len(want))
	if _, err := io.ReadFull(cli, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("tcp mimic e2e mismatch")
	}
}
