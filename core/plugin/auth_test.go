package plugin

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// runAuth drives a client/server handshake over a pipe and returns the results.
func runAuth(t *testing.T, clientSpec, serverSpec string) (identity string, cErr, sErr error) {
	t.Helper()
	ca, err := ParseAuth(clientSpec)
	if err != nil {
		t.Fatalf("client ParseAuth(%q): %v", clientSpec, err)
	}
	sa, err := ParseAuth(serverSpec)
	if err != nil {
		t.Fatalf("server ParseAuth(%q): %v", serverSpec, err)
	}
	cc, sc := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, identity, sErr = sa.ServerHandshake(sc)
	}()
	_, cErr = ca.ClientHandshake(cc)
	<-done
	cc.Close()
	sc.Close()
	return identity, cErr, sErr
}

func TestPSKAuth(t *testing.T) {
	id, cErr, sErr := runAuth(t, "psk?key=00112233", "psk?key=00112233")
	if cErr != nil || sErr != nil {
		t.Fatalf("psk should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "psk" {
		t.Fatalf("identity=%q", id)
	}
}

func TestPSKAuthWrongKey(t *testing.T) {
	_, cErr, sErr := runAuth(t, "psk?key=00112233", "psk?key=ffffffff")
	if sErr == nil {
		t.Fatal("server must reject wrong key")
	}
	if cErr == nil {
		t.Fatal("client must see rejection")
	}
}

// mintHS256 builds a signed HS256 JWT.
func mintHS256(t *testing.T, secret []byte, claims map[string]any) string {
	t.Helper()
	enc := func(v any) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	header := enc(map[string]string{"alg": "HS256", "typ": "JWT"})
	body := enc(claims)
	signing := header + "." + body
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(signing))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return signing + "." + sig
}

func TestJWTAuthHS256(t *testing.T) {
	secret := "73656372657421" // "secret!" hex
	tok := mintHS256(t, []byte{0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x21},
		map[string]any{"sub": "alice", "exp": time.Now().Add(time.Hour).Unix()})
	id, cErr, sErr := runAuth(t, "jwt?token="+tok, "jwt?alg=HS256&secret="+secret)
	if cErr != nil || sErr != nil {
		t.Fatalf("jwt should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "alice" {
		t.Fatalf("identity=%q want alice", id)
	}
}

func TestJWTAuthExpired(t *testing.T) {
	tok := mintHS256(t, []byte{1, 2, 3},
		map[string]any{"sub": "bob", "exp": time.Now().Add(-time.Hour).Unix()})
	_, cErr, sErr := runAuth(t, "jwt?token="+tok, "jwt?alg=HS256&secret=010203")
	if sErr == nil || cErr == nil {
		t.Fatal("expired token must be rejected")
	}
}

func TestJWTAuthBadSig(t *testing.T) {
	tok := mintHS256(t, []byte{1, 2, 3}, map[string]any{"sub": "x"})
	_, cErr, sErr := runAuth(t, "jwt?token="+tok, "jwt?alg=HS256&secret=ffeedd")
	if sErr == nil || cErr == nil {
		t.Fatal("wrong secret must be rejected")
	}
}

func TestJWTAuthRS256(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "pub.pem")
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0o600)

	// mint RS256
	enc := func(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	header := enc(map[string]string{"alg": "RS256", "typ": "JWT"})
	body := enc(map[string]any{"sub": "carol", "exp": time.Now().Add(time.Hour).Unix()})
	signing := header + "." + body
	sum := sha256.Sum256([]byte(signing))
	sigRaw, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	tok := signing + "." + base64.RawURLEncoding.EncodeToString(sigRaw)

	id, cErr, sErr := runAuth(t, "jwt?token="+tok, "jwt?alg=RS256&pubkey="+pubPath)
	if cErr != nil || sErr != nil {
		t.Fatalf("RS256 jwt should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "carol" {
		t.Fatalf("identity=%q want carol", id)
	}
}

func TestMTLSAuth(t *testing.T) {
	dir := t.TempDir()
	caCert, caKey := genCA(t)
	caPath := writePEM(t, dir, "ca.pem", "CERTIFICATE", caCert.Raw)

	srvCertPath, srvKeyPath := genLeaf(t, dir, "server", caCert, caKey, "server.local")
	cliCertPath, cliKeyPath := genLeaf(t, dir, "client", caCert, caKey, "client-id-1")

	clientSpec := fmt.Sprintf("mtls?cert=%s&key=%s&ca=%s&servername=server.local", cliCertPath, cliKeyPath, caPath)
	serverSpec := fmt.Sprintf("mtls?cert=%s&key=%s&clientca=%s", srvCertPath, srvKeyPath, caPath)

	id, cErr, sErr := runAuth(t, clientSpec, serverSpec)
	if cErr != nil || sErr != nil {
		t.Fatalf("mtls should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "client-id-1" {
		t.Fatalf("identity=%q want client-id-1", id)
	}
}

func TestMTLSAuthUntrustedClient(t *testing.T) {
	dir := t.TempDir()
	caCert, caKey := genCA(t)
	caPath := writePEM(t, dir, "ca.pem", "CERTIFICATE", caCert.Raw)
	srvCertPath, srvKeyPath := genLeaf(t, dir, "server", caCert, caKey, "server.local")

	// A client cert from a DIFFERENT, untrusted CA.
	otherCA, otherKey := genCA(t)
	rogueCertPath, rogueKeyPath := genLeaf(t, dir, "rogue", otherCA, otherKey, "rogue")

	clientSpec := fmt.Sprintf("mtls?cert=%s&key=%s&insecure=true", rogueCertPath, rogueKeyPath)
	serverSpec := fmt.Sprintf("mtls?cert=%s&key=%s&clientca=%s", srvCertPath, srvKeyPath, caPath)
	_, _, sErr := runAuth(t, clientSpec, serverSpec)
	if sErr == nil {
		t.Fatal("server must reject a client cert from an untrusted CA")
	}
}

// --- cert helpers ---

func genCA(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	return cert, key
}

func genLeaf(t *testing.T, dir, name string, ca *x509.Certificate, caKey *rsa.PrivateKey, cn string) (certPath, keyPath string) {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{cn},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	certPath = writePEM(t, dir, name+".crt", "CERTIFICATE", der)
	keyPath = writePEM(t, dir, name+".key", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
	return certPath, keyPath
}

func writePEM(t *testing.T, dir, file, typ string, der []byte) string {
	t.Helper()
	path := filepath.Join(dir, file)
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
