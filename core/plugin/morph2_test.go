package plugin

import (
	"bytes"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestProfileRoundTripAndQuantum(t *testing.T) {
	for _, spec := range []string{"profile?name=web&min=0&max=0", "profile?name=voip&min=0&max=0", "profile?name=custom&quantum=128&min=0&max=0"} {
		cl, _ := ParseChain(spec)
		sv, _ := ParseChain(spec)
		q := cl.plugins[0].(*profilePlugin).quantum
		for _, in := range payloads() {
			enc, err := cl.Encode(in)
			if err != nil {
				t.Fatalf("%s encode: %v", spec, err)
			}
			if len(enc)%q != 0 {
				t.Fatalf("%s: output %d not a multiple of quantum %d", spec, len(enc), q)
			}
			out, err := sv.Decode(enc)
			if err != nil {
				t.Fatalf("%s decode: %v", spec, err)
			}
			if !bytes.Equal(out, in) {
				t.Fatalf("%s round-trip mismatch", spec)
			}
		}
	}
}

func TestProfileErrors(t *testing.T) {
	for _, spec := range []string{"profile?name=bogus", "profile?name=custom&quantum=0", "profile?name=web&min=5ms&max=1ms"} {
		if _, err := ParseChain(spec); err == nil {
			t.Errorf("ParseChain(%q) should error", spec)
		}
	}
}

func TestChaffFrameTyping(t *testing.T) {
	c, _ := ParseChain("chaff")
	enc, _ := c.Encode([]byte("real-data"))
	out, err := c.Decode(enc)
	if err != nil || !bytes.Equal(out, []byte("real-data")) {
		t.Fatalf("real frame round-trip: out=%q err=%v", out, err)
	}
	// A decoy payload decodes to empty (dropped).
	decoy, err := c.EncodeChaff()
	if err != nil {
		t.Fatal(err)
	}
	dropped, err := c.Decode(decoy)
	if err != nil {
		t.Fatalf("decoy decode: %v", err)
	}
	if len(dropped) != 0 {
		t.Fatalf("decoy must decode to empty, got %d bytes", len(dropped))
	}
}

// countingConn counts Write calls so we can confirm decoy frames hit the wire.
type countingConn struct {
	net.Conn
	writes atomic.Int64
}

func (c *countingConn) Write(p []byte) (int, error) {
	c.writes.Add(1)
	return c.Conn.Write(p)
}

func TestChaffInjectsDecoyFrames(t *testing.T) {
	// Use a TCP pair (buffered) so writes don't block; drain the far end.
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		io.Copy(io.Discard, c)
	}()
	raw, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	cc := &countingConn{Conn: raw}

	ch, _ := ParseChain("chaff?interval=15ms&jitter=0&min=8&max=32,aead?key=00ff,tls-mimic")
	fc, err := NewFramedConn(cc, ch, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fc.Write([]byte("one real message")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond) // ~10 decoy intervals
	fc.Close()

	if n := cc.writes.Load(); n < 3 {
		t.Fatalf("expected decoy frames on the wire, only %d writes", n)
	}
}

func TestChaffRoundTripOverTCP(t *testing.T) {
	const spec = "chaff?interval=10ms&min=8&max=64,aead?key=0123456789abcdef,tls-mimic"
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
		c, _ := ParseChain(spec)
		srv, err := NewFramedConn(raw, c, false, 0)
		if err != nil {
			return
		}
		defer srv.Close()
		io.Copy(srv, srv) // echo real data; decoys from the client are dropped
	}()

	raw, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c, _ := ParseChain(spec)
	cli, err := NewFramedConn(raw, c, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	// Send several messages with gaps so decoy frames interleave with real ones.
	msgs := [][]byte{
		bytes.Repeat([]byte("alpha "), 100),
		[]byte("beta"),
		bytes.Repeat([]byte("gamma-"), 300),
	}
	go func() {
		for _, m := range msgs {
			cli.Write(m)
			time.Sleep(25 * time.Millisecond)
		}
	}()
	for _, want := range msgs {
		got := make([]byte, len(want))
		if _, err := io.ReadFull(cli, got); err != nil {
			t.Fatalf("read echo: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Fatal("real payload corrupted with chaff active")
		}
	}
}
