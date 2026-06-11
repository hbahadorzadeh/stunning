package plugin

import (
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// spa implements single-packet authorization port knocking. The client sends one
// UDP packet to a knock port:
//
//	nonce(16) || timestamp(8, unix sec, big-endian) || HMAC-SHA256(key, nonce||ts)
//
// The server verifies the tag, checks the timestamp is within a window, rejects
// replayed nonces, and authorizes the source IP for a TTL. The tunnel server then
// only accepts connections from authorized IPs. This reveals nothing on the
// tunnel port to an unauthenticated scanner/prober.
//
// Params:
//
//	key    hex secret (required).
//	port   UDP knock port (required).
//	ttl    how long an IP stays authorized (default 10s).
//	window timestamp tolerance (default 30s).
//	delay  client settle after sending the knock, before the caller connects, so
//	       the server authorizes the IP first (default 100ms).
func init() { RegisterKnock("spa", newSPA) }

const spaPacketLen = 16 + 8 + sha256.Size

type spaKnocker struct {
	key    []byte
	port   int
	ttl    time.Duration
	window time.Duration
	delay  time.Duration

	mu        sync.Mutex
	allow     map[string]time.Time // ip -> expiry
	seen      map[string]time.Time // nonce(hex) -> expiry (replay defense)
	lastSweep time.Time
	pc        net.PacketConn
	once      sync.Once
}

// sweepInterval bounds how often the O(N) map eviction runs, so a flood of knock
// packets cannot turn verification into a per-packet full-map scan under the lock.
const sweepInterval = 2 * time.Second

func newSPA(p Params) (Knocker, error) {
	key, err := p.Bytes("key", nil)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("spa: missing required param key")
	}
	port := p.Int("port", 0)
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("spa: invalid port %d", port)
	}
	return &spaKnocker{
		key:    key,
		port:   port,
		ttl:    p.Duration("ttl", 10*time.Second),
		window: p.Duration("window", 30*time.Second),
		delay:  p.Duration("delay", 100*time.Millisecond),
		allow:  map[string]time.Time{},
		seen:   map[string]time.Time{},
	}, nil
}

func (s *spaKnocker) tag(nonce, ts []byte) []byte {
	h := hmac.New(sha256.New, s.key)
	h.Write(nonce)
	h.Write(ts)
	return h.Sum(nil)
}

func (s *spaKnocker) Knock(host string) error {
	pkt := make([]byte, spaPacketLen)
	if _, err := io.ReadFull(crand.Reader, pkt[:16]); err != nil {
		return err
	}
	binary.BigEndian.PutUint64(pkt[16:24], uint64(time.Now().Unix()))
	copy(pkt[24:], s.tag(pkt[:16], pkt[16:24]))

	conn, err := net.Dial("udp", net.JoinHostPort(host, fmt.Sprintf("%d", s.port)))
	if err != nil {
		return fmt.Errorf("spa: dial knock port: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Write(pkt); err != nil {
		return fmt.Errorf("spa: send knock: %w", err)
	}
	// Give the server a moment to authorize this IP before the caller connects.
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return nil
}

func (s *spaKnocker) Start() error {
	var startErr error
	s.once.Do(func() {
		pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			startErr = fmt.Errorf("spa: listen knock port: %w", err)
			return
		}
		s.pc = pc
		go s.serve()
	})
	return startErr
}

func (s *spaKnocker) serve() {
	buf := make([]byte, spaPacketLen)
	for {
		n, addr, err := s.pc.ReadFrom(buf)
		if err != nil {
			return // listener closed
		}
		if n != spaPacketLen {
			continue
		}
		if ip := s.verify(buf[:n]); ip {
			host, _, _ := net.SplitHostPort(addr.String())
			s.authorize(host)
		}
	}
}

// verify validates the tag, timestamp window, and replay status.
func (s *spaKnocker) verify(pkt []byte) bool {
	nonce, ts, mac := pkt[:16], pkt[16:24], pkt[24:]
	if subtle.ConstantTimeCompare(mac, s.tag(nonce, ts)) != 1 {
		return false
	}
	when := int64(binary.BigEndian.Uint64(ts))
	now := time.Now().Unix()
	if d := now - when; d > int64(s.window/time.Second) || d < -int64(s.window/time.Second) {
		return false
	}
	key := hex.EncodeToString(nonce)
	s.mu.Lock()
	defer s.mu.Unlock()
	tnow := time.Now()
	if tnow.Sub(s.lastSweep) > sweepInterval {
		s.sweep(tnow)
		s.lastSweep = tnow
	}
	if _, replay := s.seen[key]; replay {
		return false
	}
	s.seen[key] = tnow.Add(s.window)
	return true
}

func (s *spaKnocker) authorize(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allow[ip] = time.Now().Add(s.ttl)
}

func (s *spaKnocker) Authorized(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.allow[ip]
	return ok && time.Now().Before(exp)
}

// sweep drops expired allow/seen entries (caller holds mu).
func (s *spaKnocker) sweep(now time.Time) {
	for k, exp := range s.allow {
		if now.After(exp) {
			delete(s.allow, k)
		}
	}
	for k, exp := range s.seen {
		if now.After(exp) {
			delete(s.seen, k)
		}
	}
}

func (s *spaKnocker) Close() error {
	if s.pc != nil {
		return s.pc.Close()
	}
	return nil
}
