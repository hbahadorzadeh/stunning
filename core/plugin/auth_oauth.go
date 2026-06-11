package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// oauth authenticates the client by an OAuth 2.0 access token validated against
// an RFC 7662 token-introspection endpoint. The client presents a bearer token;
// the server posts it to the introspection endpoint (with client credentials)
// and accepts the connection only if the token is active. Identity is the token's
// username/sub.
//
// Client param:  token  the access token to present.
// Server params: introspect (introspection URL), client_id, client_secret,
//
//	scope (optional required scope).
func init() { RegisterAuth("oauth", newOAuthAuth) }

type oauthAuth struct {
	token        string // client side
	introspect   string // server side
	clientID     string
	clientSecret string
	scope        string
	hc           *http.Client
}

func newOAuthAuth(p Params) (Authenticator, error) {
	a := &oauthAuth{
		token:        p.String("token", ""),
		introspect:   p.String("introspect", ""),
		clientID:     p.String("client_id", ""),
		clientSecret: p.String("client_secret", ""),
		scope:        p.String("scope", ""),
		hc:           &http.Client{Timeout: 10 * time.Second},
	}
	return a, nil
}

func (a *oauthAuth) ClientHandshake(conn net.Conn) (net.Conn, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	if a.token == "" {
		return nil, fmt.Errorf("oauth: client has no token param")
	}
	if err := writeMsg(conn, []byte(a.token)); err != nil {
		return nil, err
	}
	status, err := readMsg(conn)
	if err != nil {
		return nil, err
	}
	if string(status) != "OK" {
		return nil, fmt.Errorf("oauth: rejected by server")
	}
	return conn, nil
}

func (a *oauthAuth) ServerHandshake(conn net.Conn) (net.Conn, string, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	if a.introspect == "" {
		return nil, "", fmt.Errorf("oauth: server has no introspect URL")
	}
	tok, err := readMsg(conn)
	if err != nil {
		return nil, "", err
	}
	id, err := a.introspectToken(string(tok))
	if err != nil {
		_ = writeMsg(conn, []byte("NO"))
		return nil, "", err
	}
	if err := writeMsg(conn, []byte("OK")); err != nil {
		return nil, "", err
	}
	return conn, id, nil
}

// introspectToken posts the token to the introspection endpoint and returns the
// identity if it is active (and carries the required scope, if configured).
func (a *oauthAuth) introspectToken(token string) (string, error) {
	form := url.Values{"token": {token}}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		a.introspect, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if a.clientID != "" {
		req.SetBasicAuth(a.clientID, a.clientSecret)
	}
	resp, err := a.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth: introspection request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth: introspection status %d", resp.StatusCode)
	}
	var out struct {
		Active   bool   `json:"active"`
		Username string `json:"username"`
		Sub      string `json:"sub"`
		Scope    string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("oauth: decode introspection: %w", err)
	}
	if !out.Active {
		return "", fmt.Errorf("oauth: token inactive")
	}
	if a.scope != "" && !hasScope(out.Scope, a.scope) {
		return "", fmt.Errorf("oauth: missing required scope %q", a.scope)
	}
	id := out.Sub
	if id == "" {
		id = out.Username
	}
	return id, nil
}

func hasScope(scopes, want string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == want {
			return true
		}
	}
	return false
}
