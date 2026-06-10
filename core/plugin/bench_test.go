package plugin

import (
	"bytes"
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
