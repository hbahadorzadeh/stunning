package plugin

import (
	"bytes"
	"testing"
)

// payloads exercises empty, tiny, text, and binary inputs.
func payloads() [][]byte {
	return [][]byte{
		{},
		[]byte("x"),
		[]byte("hello world"),
		bytes.Repeat([]byte("compressible-compressible-"), 64),
		func() []byte {
			b := make([]byte, 4096)
			for i := range b {
				b[i] = byte(i * 31)
			}
			return b
		}(),
	}
}

func TestParamsGetters(t *testing.T) {
	p := Params{"a": "5", "b": "true", "c": "deadbeef", "d": "250ms", "e": "x"}
	if p.Int("a", 0) != 5 {
		t.Fatal("Int")
	}
	if p.Int("missing", 9) != 9 || p.Int("e", 9) != 9 {
		t.Fatal("Int default")
	}
	if !p.Bool("b", false) || p.Bool("missing", true) != true {
		t.Fatal("Bool")
	}
	if got, _ := p.Bytes("c", nil); !bytes.Equal(got, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatal("Bytes")
	}
	if _, err := p.Bytes("e", nil); err == nil {
		t.Fatal("Bytes should reject non-hex")
	}
	if p.Duration("d", 0).Milliseconds() != 250 {
		t.Fatal("Duration")
	}
	if p.String("missing", "def") != "def" {
		t.Fatal("String default")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	Register("flate", newFlate) // already registered in init
}

// directional builds a client and server instance of a single-plugin chain.
func directional(t *testing.T, spec string) (*Chain, *Chain) {
	t.Helper()
	cl, err := ParseChain(spec)
	if err != nil {
		t.Fatalf("client ParseChain(%q): %v", spec, err)
	}
	sv, err := ParseChain(spec)
	if err != nil {
		t.Fatalf("server ParseChain(%q): %v", spec, err)
	}
	return cl, sv
}

func TestPluginRoundTrip(t *testing.T) {
	specs := []string{
		"flate",
		"flate?level=9",
		"aead?key=0011223344556677",
		"aead?key=0011223344556677&algo=aesgcm",
		"pad?min=0&max=0",
		"pad?min=8&max=64",
		"probe-guard?key=cafe&taglen=16",
	}
	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			cl, sv := directional(t, spec)
			defer cl.Close()
			defer sv.Close()
			for _, in := range payloads() {
				enc, err := cl.Encode(in)
				if err != nil {
					t.Fatalf("Encode: %v", err)
				}
				out, err := sv.Decode(enc)
				if err != nil {
					t.Fatalf("Decode: %v", err)
				}
				if !bytes.Equal(out, in) {
					t.Fatalf("round-trip mismatch: in=%d out=%d", len(in), len(out))
				}
			}
		})
	}
}

func TestAEADTamperRejected(t *testing.T) {
	cl, sv := directional(t, "aead?key=00112233")
	defer cl.Close()
	defer sv.Close()
	enc, _ := cl.Encode([]byte("secret payload"))
	enc[len(enc)-1] ^= 0xff // flip a ciphertext bit
	if _, err := sv.Decode(enc); err == nil {
		t.Fatal("tampered aead frame must fail to open")
	}
}

func TestProbeGuardRejectsUnauthenticated(t *testing.T) {
	cl, sv := directional(t, "probe-guard?key=abcd")
	defer cl.Close()
	defer sv.Close()
	if _, err := sv.Decode([]byte("an unauthenticated probe payload")); err == nil {
		t.Fatal("probe-guard must reject frames with no valid tag")
	}
	enc, _ := cl.Encode([]byte("real"))
	enc[0] ^= 0x01 // corrupt the tag
	if _, err := sv.Decode(enc); err == nil {
		t.Fatal("probe-guard must reject corrupted tag")
	}
}

func TestChainErrors(t *testing.T) {
	cases := []string{
		"nope",                        // unknown plugin
		"aead",                        // missing key
		"probe-guard",                 // missing key
		"pad?min=10&max=2",            // bad bounds
		"aead?key=zz",                 // bad hex
		"probe-guard?key=ab&taglen=4", // taglen too small
		"flate?level=99",              // bad level
		"aead?key=ab&algo=bogus",      // bad algo
	}
	for _, spec := range cases {
		if _, err := ParseChain(spec); err == nil {
			t.Errorf("ParseChain(%q) should error", spec)
		}
	}
}

func TestEmptyChainPassthrough(t *testing.T) {
	c, err := ParseChain("")
	if err != nil {
		t.Fatal(err)
	}
	if c.Len() != 0 {
		t.Fatal("empty spec should yield empty chain")
	}
	in := []byte("unchanged")
	enc, _ := c.Encode(in)
	if !bytes.Equal(enc, in) {
		t.Fatal("empty chain must be identity")
	}
}

func TestFullChainPermutations(t *testing.T) {
	// Every ordering of the four categories must round-trip.
	parts := []string{"flate", "aead?key=00ff", "pad?min=4&max=40", "probe-guard?key=11ee"}
	perms := permute(parts)
	in := bytes.Repeat([]byte("the quick brown fox 0123456789 "), 40)
	for _, perm := range perms {
		spec := join(perm)
		t.Run(spec, func(t *testing.T) {
			cl, sv := directional(t, spec)
			defer cl.Close()
			defer sv.Close()
			enc, err := cl.Encode(in)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			out, err := sv.Decode(enc)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if !bytes.Equal(out, in) {
				t.Fatal("permutation round-trip mismatch")
			}
		})
	}
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}

func permute(s []string) [][]string {
	if len(s) <= 1 {
		return [][]string{append([]string(nil), s...)}
	}
	var res [][]string
	for i := range s {
		rest := append(append([]string(nil), s[:i]...), s[i+1:]...)
		for _, p := range permute(rest) {
			res = append(res, append([]string{s[i]}, p...))
		}
	}
	return res
}
