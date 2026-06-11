// Command tools provides the traffic endpoints for the DPI test harness.
//
//	tools echo  -listen :9000              run a TCP echo destination
//	tools gen   -connect host:1080 ...     drive traffic, emit JSON metrics
//
// gen connects through the stunning client interface, streams random data to the
// echo destination, verifies the bytes return intact, and reports throughput and
// latency as JSON on stdout so the scenario runner can assert on it.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: tools <echo|gen> [flags]")
	}
	switch os.Args[1] {
	case "echo":
		runEcho(os.Args[2:])
	case "gen":
		runGen(os.Args[2:])
	case "probe":
		runProbe(os.Args[2:])
	default:
		log.Fatalf("unknown subcommand %q", os.Args[1])
	}
}

// runProbe just checks that a TCP listener is accepting, independent of any
// downstream DPI verdict. Used for harness readiness on the local interface.
func runProbe(argv []string) {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)
	connect := fs.String("connect", "127.0.0.1:1080", "address to dial")
	timeout := fs.Duration("timeout", 2*time.Second, "dial timeout")
	fs.Parse(argv)
	conn, err := net.DialTimeout("tcp", *connect, *timeout)
	if err != nil {
		os.Exit(1)
	}
	conn.Close()
}

func runEcho(argv []string) {
	fs := flag.NewFlagSet("echo", flag.ExitOnError)
	listen := fs.String("listen", ":9000", "listen address")
	fs.Parse(argv)

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("echo listen: %v", err)
	}
	log.Printf("echo destination listening on %s", *listen)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			io.Copy(conn, conn)
		}(c)
	}
}

// Metrics is the JSON result emitted by gen.
type Metrics struct {
	OK          bool    `json:"ok"`
	Error       string  `json:"error,omitempty"`
	Bytes       int     `json:"bytes"`
	Streams     int     `json:"streams"`
	Wallclock   float64 `json:"wallclock_s"`
	ThroughMBps float64 `json:"throughput_mbps"`
	ConnectMs   float64 `json:"connect_ms"`
	RTTMs       float64 `json:"first_byte_rtt_ms"`
}

func runGen(argv []string) {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	connect := fs.String("connect", "127.0.0.1:1080", "target address (stunning client interface)")
	size := fs.Int("size", 4<<20, "bytes per stream")
	streams := fs.Int("streams", 4, "concurrent streams")
	timeout := fs.Duration("timeout", 30*time.Second, "overall timeout")
	fs.Parse(argv)

	m, err := generate(*connect, *size, *streams, *timeout)
	if err != nil {
		m.OK = false
		m.Error = err.Error()
	} else {
		m.OK = true
	}
	out, _ := json.Marshal(m)
	fmt.Println(string(out))
	if !m.OK {
		os.Exit(1)
	}
}

// textPayload builds a low-entropy, realistic-looking byte stream (so a plaintext
// tunnel resembles ordinary web traffic, not random noise). It embeds the marker
// "STUNNING-PROBE" so a DPI marker rule can catch an unobfuscated tunnel.
func textPayload(size int) []byte {
	const lorem = "GET /index.html HTTP/1.1\r\nHost: example.com STUNNING-PROBE\r\n" +
		"User-Agent: Mozilla/5.0 the quick brown fox jumps over the lazy dog 0123456789\r\n\r\n"
	out := make([]byte, size)
	for i := range out {
		out[i] = lorem[i%len(lorem)]
	}
	return out
}

func generate(addr string, size, streams int, timeout time.Duration) (Metrics, error) {
	var m Metrics
	m.Bytes = size * streams
	m.Streams = streams

	start := time.Now()
	deadline := start.Add(timeout)

	type res struct {
		connect, rtt time.Duration
		err          error
	}
	results := make(chan res, streams)
	for i := 0; i < streams; i++ {
		go func() {
			results <- runStream(addr, size, deadline)
		}()
	}
	var firstConnect, firstRTT time.Duration
	for i := 0; i < streams; i++ {
		r := <-results
		if r.err != nil {
			return m, r.err
		}
		if i == 0 || r.connect < firstConnect {
			firstConnect = r.connect
		}
		if i == 0 || r.rtt < firstRTT {
			firstRTT = r.rtt
		}
	}
	m.Wallclock = time.Since(start).Seconds()
	m.ConnectMs = float64(firstConnect.Microseconds()) / 1000
	m.RTTMs = float64(firstRTT.Microseconds()) / 1000
	if m.Wallclock > 0 {
		m.ThroughMBps = float64(m.Bytes) / m.Wallclock / 1e6
	}
	return m, nil
}

func runStream(addr string, size int, deadline time.Time) (r struct {
	connect, rtt time.Duration
	err          error
}) {
	t0 := time.Now()
	conn, err := net.DialTimeout("tcp", addr, time.Until(deadline))
	if err != nil {
		r.err = fmt.Errorf("dial: %w", err)
		return
	}
	defer conn.Close()
	conn.SetDeadline(deadline)
	r.connect = time.Since(t0)

	payload := textPayload(size)
	want := sha256.Sum256(payload)

	writeErr := make(chan error, 1)
	go func() {
		_, err := conn.Write(payload)
		writeErr <- err
	}()

	got := make([]byte, size)
	tRead := time.Now()
	if _, err := io.ReadFull(conn, got); err != nil {
		r.err = fmt.Errorf("read back: %w", err)
		return
	}
	r.rtt = time.Since(tRead)
	if err := <-writeErr; err != nil {
		r.err = fmt.Errorf("write: %w", err)
		return
	}
	if sha256.Sum256(got) != want {
		r.err = fmt.Errorf("integrity mismatch: echoed bytes differ")
		return
	}
	return
}
