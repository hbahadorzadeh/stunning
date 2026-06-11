package plugin

import (
	"bytes"
	"io"
	"net"
	"testing"
)

func benchChain(b *testing.B, spec string, size int) {
	cl, _ := ParseChain(spec)
	sv, _ := ParseChain(spec)
	defer cl.Close()
	defer sv.Close()
	in := bytes.Repeat([]byte("benchmark-payload-mostly-text-0123456789 "), size/41+1)[:size]
	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc, err := cl.Encode(in)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := sv.Decode(enc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlate(b *testing.B)      { benchChain(b, "flate", 4096) }
func BenchmarkAEADChaCha(b *testing.B) { benchChain(b, "aead?key=0123456789abcdef", 4096) }
func BenchmarkAEADAESGCM(b *testing.B) { benchChain(b, "aead?key=0123456789abcdef&algo=aesgcm", 4096) }
func BenchmarkPad(b *testing.B)        { benchChain(b, "pad?min=16&max=256", 4096) }
func BenchmarkProbeGuard(b *testing.B) { benchChain(b, "probe-guard?key=cafe", 4096) }
func BenchmarkFullChain(b *testing.B) {
	benchChain(b, "flate,aead?key=0123456789abcdef,pad?min=16&max=256,probe-guard?key=cafe", 4096)
}
func BenchmarkFullChain64K(b *testing.B) {
	benchChain(b, "flate,aead?key=0123456789abcdef,pad?min=16&max=256,probe-guard?key=cafe", 65536)
}

// benchFramed measures a full FramedConn write+read round trip over a socket pair,
// so it includes the framing/mimicry path (Frame/Deframe), not just the inner
// chain transforms.
func benchFramed(b *testing.B, spec string, size int) {
	cc, sc := net.Pipe()
	clCh, _ := ParseChain(spec)
	svCh, _ := ParseChain(spec)
	client, err := NewFramedConn(cc, clCh, true, 1<<20)
	if err != nil {
		b.Fatal(err)
	}
	server, err := NewFramedConn(sc, svCh, false, 1<<20)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()
	defer server.Close()
	in := bytes.Repeat([]byte("framed-benchmark-payload "), size/25+1)[:size]
	out := make([]byte, size)
	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go client.Write(in)
		if _, err := io.ReadFull(server, out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFramedTLSMimic(b *testing.B) {
	benchFramed(b, "aead?key=0123456789abcdef,tls-mimic", 4096)
}
func BenchmarkFramedHTTPMimic(b *testing.B) {
	benchFramed(b, "aead?key=0123456789abcdef,http-mimic", 4096)
}
func BenchmarkFramedFullMimic(b *testing.B) {
	benchFramed(b, "flate,aead?key=0123456789abcdef,bucket?size=512,tls-mimic", 4096)
}
func BenchmarkFramedBaseline(b *testing.B) {
	benchFramed(b, "aead?key=0123456789abcdef", 4096)
}
