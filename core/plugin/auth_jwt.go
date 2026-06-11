package plugin

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"time"
)

// jwt authenticates the client by a signed JSON Web Token. The client presents a
// token (param token); the server verifies its signature (HS256 with a shared
// secret, or RS256 with a public key) and its exp/nbf, and takes the subject
// claim as the identity.
//
// Client param:  token  the JWT to present.
// Server params: alg (HS256|RS256, default HS256),
//
//	secret (hex, HS256), pubkey (PEM file path, RS256).
func init() { RegisterAuth("jwt", newJWTAuth) }

type jwtAuth struct {
	token  string // client side
	alg    string // server side
	secret []byte // HS256
	pub    *rsa.PublicKey
}

func newJWTAuth(p Params) (Authenticator, error) {
	a := &jwtAuth{token: p.String("token", ""), alg: p.String("alg", "HS256")}
	switch a.alg {
	case "HS256":
		sec, err := p.Bytes("secret", nil)
		if err != nil {
			return nil, err
		}
		a.secret = sec
	case "RS256":
		if path := p.String("pubkey", ""); path != "" {
			pub, err := loadRSAPublicKey(path)
			if err != nil {
				return nil, err
			}
			a.pub = pub
		}
	default:
		return nil, fmt.Errorf("jwt: unsupported alg %q", a.alg)
	}
	return a, nil
}

func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("jwt: read pubkey: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("jwt: pubkey not PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Fall back to PKCS#1 (BEGIN RSA PUBLIC KEY).
		if rp, err2 := x509.ParsePKCS1PublicKey(block.Bytes); err2 == nil {
			return rp, nil
		}
		return nil, fmt.Errorf("jwt: parse pubkey: %w", err)
	}
	rp, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("jwt: pubkey is not RSA")
	}
	return rp, nil
}

func (a *jwtAuth) ClientHandshake(conn net.Conn) (net.Conn, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	if a.token == "" {
		return nil, fmt.Errorf("jwt: client has no token param")
	}
	if err := writeMsg(conn, []byte(a.token)); err != nil {
		return nil, err
	}
	status, err := readMsg(conn)
	if err != nil {
		return nil, err
	}
	if string(status) != "OK" {
		return nil, fmt.Errorf("jwt: rejected by server")
	}
	return conn, nil
}

func (a *jwtAuth) ServerHandshake(conn net.Conn) (net.Conn, string, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	tok, err := readMsg(conn)
	if err != nil {
		return nil, "", err
	}
	sub, err := a.verify(string(tok))
	if err != nil {
		_ = writeMsg(conn, []byte("NO"))
		return nil, "", err
	}
	if err := writeMsg(conn, []byte("OK")); err != nil {
		return nil, "", err
	}
	return conn, sub, nil
}

// verify checks the token signature and time claims, returning the subject.
func (a *jwtAuth) verify(token string) (string, error) {
	var p1, p2, p3 string
	parts := splitN(token, '.', 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("jwt: malformed token")
	}
	p1, p2, p3 = parts[0], parts[1], parts[2]
	signingInput := p1 + "." + p2

	var hdr struct{ Alg string }
	if err := decodeSegment(p1, &hdr); err != nil {
		return "", fmt.Errorf("jwt: bad header: %w", err)
	}
	if hdr.Alg != a.alg {
		return "", fmt.Errorf("jwt: alg mismatch: token=%s want=%s", hdr.Alg, a.alg)
	}
	sig, err := base64.RawURLEncoding.DecodeString(p3)
	if err != nil {
		return "", fmt.Errorf("jwt: bad signature encoding")
	}

	switch a.alg {
	case "HS256":
		h := hmac.New(sha256.New, a.secret)
		h.Write([]byte(signingInput))
		if subtle.ConstantTimeCompare(sig, h.Sum(nil)) != 1 {
			return "", fmt.Errorf("jwt: signature invalid")
		}
	case "RS256":
		if a.pub == nil {
			return "", fmt.Errorf("jwt: no public key configured")
		}
		sum := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(a.pub, crypto.SHA256, sum[:], sig); err != nil {
			return "", fmt.Errorf("jwt: signature invalid")
		}
	}

	var claims struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
		Nbf int64  `json:"nbf"`
	}
	if err := decodeSegment(p2, &claims); err != nil {
		return "", fmt.Errorf("jwt: bad claims: %w", err)
	}
	now := time.Now().Unix()
	const leeway = 60 // seconds of clock-skew tolerance
	if claims.Exp != 0 && now >= claims.Exp+leeway {
		return "", fmt.Errorf("jwt: token expired")
	}
	if claims.Nbf != 0 && now < claims.Nbf-leeway {
		return "", fmt.Errorf("jwt: token not yet valid")
	}
	return claims.Sub, nil
}

func decodeSegment(seg string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

// splitN splits s on sep into at most n parts (avoids importing strings just for
// this and keeps empty segments).
func splitN(s string, sep byte, n int) []string {
	out := make([]string, 0, n)
	start := 0
	for i := 0; i < len(s) && len(out) < n-1; i++ {
		if s[i] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
