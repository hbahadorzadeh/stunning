package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// mtls authenticates via a mutual-TLS handshake layered over the connection. The
// server requires and verifies a client certificate against a CA; the client
// presents its certificate and verifies the server. The identity is the client
// certificate's Common Name. Unlike other authenticators this wraps the conn in
// a TLS session, which is returned for subsequent data.
//
// Client params: cert, key (client cert+key PEM paths), ca (server CA PEM),
//
//	insecure (skip server verification, testing only).
//
// Server params: cert, key (server cert+key PEM paths), clientca (client CA PEM).
func init() { RegisterAuth("mtls", newMTLSAuth) }

type mtlsAuth struct {
	cert       *tls.Certificate
	rootCAs    *x509.CertPool // client: verify server
	clientCA   *x509.CertPool // server: verify client
	serverName string         // client: expected server cert name
	insecure   bool
}

func newMTLSAuth(p Params) (Authenticator, error) {
	a := &mtlsAuth{insecure: p.Bool("insecure", false), serverName: p.String("servername", "")}
	if a.insecure {
		log.Printf("WARNING: mtls insecure=true disables server certificate verification (testing only)")
	}
	if c, k := p.String("cert", ""), p.String("key", ""); c != "" && k != "" {
		pair, err := tls.LoadX509KeyPair(c, k)
		if err != nil {
			return nil, fmt.Errorf("mtls: load keypair: %w", err)
		}
		a.cert = &pair
	}
	if ca := p.String("ca", ""); ca != "" {
		pool, err := loadCertPool(ca)
		if err != nil {
			return nil, err
		}
		a.rootCAs = pool
	}
	if ca := p.String("clientca", ""); ca != "" {
		pool, err := loadCertPool(ca)
		if err != nil {
			return nil, err
		}
		a.clientCA = pool
	}
	return a, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("mtls: read ca: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("mtls: no certs in %s", path)
	}
	return pool, nil
}

func (a *mtlsAuth) ClientHandshake(conn net.Conn) (net.Conn, error) {
	cfg := &tls.Config{
		RootCAs:            a.rootCAs,
		ServerName:         a.serverName,
		InsecureSkipVerify: a.insecure,
		MinVersion:         tls.VersionTLS12,
	}
	if a.cert != nil {
		cfg.Certificates = []tls.Certificate{*a.cert}
	}
	tc := tls.Client(conn, cfg)
	_ = tc.SetDeadline(time.Now().Add(HandshakeTimeout))
	if err := tc.Handshake(); err != nil {
		return nil, fmt.Errorf("mtls: client handshake: %w", err)
	}
	_ = tc.SetDeadline(time.Time{})
	return tc, nil
}

func (a *mtlsAuth) ServerHandshake(conn net.Conn) (net.Conn, string, error) {
	if a.cert == nil {
		return nil, "", fmt.Errorf("mtls: server requires cert+key")
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{*a.cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    a.clientCA,
		MinVersion:   tls.VersionTLS12,
	}
	tc := tls.Server(conn, cfg)
	_ = tc.SetDeadline(time.Now().Add(HandshakeTimeout))
	if err := tc.Handshake(); err != nil {
		return nil, "", fmt.Errorf("mtls: server handshake: %w", err)
	}
	_ = tc.SetDeadline(time.Time{})
	cn := ""
	if certs := tc.ConnectionState().PeerCertificates; len(certs) > 0 {
		cn = certs[0].Subject.CommonName
	}
	return tc, cn, nil
}
