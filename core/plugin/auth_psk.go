package plugin

import (
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"net"
	"time"
)

// psk authenticates the client with a pre-shared key via HMAC-SHA256
// challenge-response: the server sends a fresh random challenge, the client
// returns HMAC(key, challenge), and the server verifies it in constant time. The
// per-connection challenge makes responses non-replayable.
//
// Param: key (hex, required). The same key on both peers.
func init() { RegisterAuth("psk", newPSKAuth) }

type pskAuth struct{ key []byte }

func newPSKAuth(p Params) (Authenticator, error) {
	key, err := p.Bytes("key", nil)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("psk: missing required param key")
	}
	return &pskAuth{key: key}, nil
}

func (a *pskAuth) mac(challenge []byte) []byte {
	h := hmac.New(sha256.New, a.key)
	h.Write(challenge)
	return h.Sum(nil)
}

func (a *pskAuth) ClientHandshake(conn net.Conn) (net.Conn, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	challenge, err := readMsg(conn)
	if err != nil {
		return nil, err
	}
	if len(challenge) < 16 {
		return nil, fmt.Errorf("psk: short challenge")
	}
	if err := writeMsg(conn, a.mac(challenge)); err != nil {
		return nil, err
	}
	status, err := readMsg(conn)
	if err != nil {
		return nil, err
	}
	if string(status) != "OK" {
		return nil, fmt.Errorf("psk: rejected by server")
	}
	return conn, nil
}

func (a *pskAuth) ServerHandshake(conn net.Conn) (net.Conn, string, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	challenge := make([]byte, 32)
	if _, err := io.ReadFull(crand.Reader, challenge); err != nil {
		return nil, "", err
	}
	if err := writeMsg(conn, challenge); err != nil {
		return nil, "", err
	}
	resp, err := readMsg(conn)
	if err != nil {
		return nil, "", err
	}
	if subtle.ConstantTimeCompare(resp, a.mac(challenge)) != 1 {
		_ = writeMsg(conn, []byte("NO"))
		return nil, "", fmt.Errorf("psk: authentication failed")
	}
	if err := writeMsg(conn, []byte("OK")); err != nil {
		return nil, "", err
	}
	return conn, "psk", nil
}
