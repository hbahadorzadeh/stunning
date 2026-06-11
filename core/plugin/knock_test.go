package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func freeUDPPort(t *testing.T) int {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	_, ps, _ := net.SplitHostPort(pc.LocalAddr().String())
	p, _ := strconv.Atoi(ps)
	return p
}

func mustSPA(t *testing.T, spec string) *spaKnocker {
	t.Helper()
	k, err := ParseKnock(spec)
	if err != nil {
		t.Fatalf("ParseKnock(%q): %v", spec, err)
	}
	return k.(*spaKnocker)
}

func waitAuthorized(k Knocker, ip string, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if k.Authorized(ip) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestSPAKnockAuthorizes(t *testing.T) {
	port := freeUDPPort(t)
	spec := fmt.Sprintf("spa?key=00112233&port=%d&ttl=2s", port)
	srv := mustSPA(t, spec)
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	if srv.Authorized("127.0.0.1") {
		t.Fatal("should not be authorized before knocking")
	}
	cli := mustSPA(t, spec)
	if err := cli.Knock("127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if !waitAuthorized(srv, "127.0.0.1", time.Second) {
		t.Fatal("IP should be authorized after a valid knock")
	}
}

func TestSPAWrongKeyNotAuthorized(t *testing.T) {
	port := freeUDPPort(t)
	srv := mustSPA(t, fmt.Sprintf("spa?key=aaaa&port=%d", port))
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := mustSPA(t, fmt.Sprintf("spa?key=bbbb&port=%d", port))
	if err := cli.Knock("127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if waitAuthorized(srv, "127.0.0.1", 300*time.Millisecond) {
		t.Fatal("wrong-key knock must not authorize")
	}
}

// craftPacket builds an SPA packet with the given key and timestamp.
func craftPacket(t *testing.T, k *spaKnocker, ts int64) []byte {
	t.Helper()
	pkt := make([]byte, spaPacketLen)
	if _, err := io.ReadFull(rand.Reader, pkt[:16]); err != nil {
		t.Fatal(err)
	}
	binary.BigEndian.PutUint64(pkt[16:24], uint64(ts))
	copy(pkt[24:], k.tag(pkt[:16], pkt[16:24]))
	return pkt
}

func TestSPAReplayRejected(t *testing.T) {
	k := mustSPA(t, "spa?key=00112233&port=1&window=30s")
	pkt := craftPacket(t, k, time.Now().Unix())
	if !k.verify(pkt) {
		t.Fatal("first valid packet should verify")
	}
	if k.verify(pkt) {
		t.Fatal("replayed packet (same nonce) must be rejected")
	}
}

func TestSPAExpiredTimestamp(t *testing.T) {
	k := mustSPA(t, "spa?key=00112233&port=1&window=5s")
	old := craftPacket(t, k, time.Now().Add(-time.Hour).Unix())
	if k.verify(old) {
		t.Fatal("packet outside the timestamp window must be rejected")
	}
}

func TestSPATamperRejected(t *testing.T) {
	k := mustSPA(t, "spa?key=00112233&port=1")
	pkt := craftPacket(t, k, time.Now().Unix())
	pkt[len(pkt)-1] ^= 0xff // corrupt the tag
	if k.verify(pkt) {
		t.Fatal("tampered packet must be rejected")
	}
}

func TestSPATTLExpiry(t *testing.T) {
	k := mustSPA(t, "spa?key=00112233&port=1&ttl=40ms")
	k.authorize("10.0.0.5")
	if !k.Authorized("10.0.0.5") {
		t.Fatal("should be authorized immediately")
	}
	time.Sleep(80 * time.Millisecond)
	if k.Authorized("10.0.0.5") {
		t.Fatal("authorization must expire after ttl")
	}
}

func TestSPAParamErrors(t *testing.T) {
	for _, spec := range []string{"spa?port=1234", "spa?key=ab", "spa?key=ab&port=0"} {
		if _, err := ParseKnock(spec); err == nil {
			t.Errorf("ParseKnock(%q) should error", spec)
		}
	}
}

func TestSPAStartCloseRestart(t *testing.T) {
	port := freeUDPPort(t)
	k := mustSPA(t, fmt.Sprintf("spa?key=00112233&port=%d", port))
	if err := k.Start(); err != nil {
		t.Fatal(err)
	}
	if err := k.Close(); err != nil {
		t.Fatal(err)
	}
	// Start must rebind after Close (pc reset to nil), not silently no-op.
	if err := k.Start(); err != nil {
		t.Fatalf("restart after close should succeed: %v", err)
	}
	k.Close()
}
