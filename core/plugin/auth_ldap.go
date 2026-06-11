package plugin

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

// ldap authenticates the client by username/password verified with an LDAP bind.
// The client sends its credentials; the server binds to the directory as the
// user's DN with the supplied password and accepts the connection if the bind
// succeeds. Identity is the username.
//
// Run ldap behind an encrypting chain (e.g. aead) or use ldaps, since the
// password crosses the connection.
//
// Client params: user, password.
// Server params: url (e.g. ldap://host:389 or ldaps://host:636),
//
//	userdn (DN template with a single %s for the username,
//	        e.g. uid=%s,ou=people,dc=example,dc=org).
func init() { RegisterAuth("ldap", newLDAPAuth) }

// ldapBind performs the directory bind. It is a package variable so tests can
// substitute an in-process fake without a live LDAP server.
var ldapBind = realLDAPBind

func realLDAPBind(serverURL, dn, password string) error {
	l, err := ldap.DialURL(serverURL)
	if err != nil {
		return fmt.Errorf("ldap: dial: %w", err)
	}
	defer l.Close()
	if err := l.Bind(dn, password); err != nil {
		return fmt.Errorf("ldap: bind: %w", err)
	}
	return nil
}

type ldapAuth struct {
	user     string // client
	password string // client
	url      string // server
	userDN   string // server: DN template
}

func newLDAPAuth(p Params) (Authenticator, error) {
	userDN := p.String("userdn", "")
	// A userdn without exactly one %s would ignore the username (static DN), so
	// anyone with that DN's password could authenticate as it. Reject it.
	if userDN != "" {
		if !strings.Contains(userDN, "%s") || strings.Count(userDN, "%") != 1 {
			return nil, fmt.Errorf("ldap: userdn must contain exactly one %%s placeholder")
		}
	}
	return &ldapAuth{
		user:     p.String("user", ""),
		password: p.String("password", ""),
		url:      p.String("url", ""),
		userDN:   userDN,
	}, nil
}

func (a *ldapAuth) ClientHandshake(conn net.Conn) (net.Conn, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	if a.user == "" {
		return nil, fmt.Errorf("ldap: client has no user param")
	}
	if err := writeMsg(conn, []byte(a.user)); err != nil {
		return nil, err
	}
	if err := writeMsg(conn, []byte(a.password)); err != nil {
		return nil, err
	}
	status, err := readMsg(conn)
	if err != nil {
		return nil, err
	}
	if string(status) != "OK" {
		return nil, fmt.Errorf("ldap: rejected by server")
	}
	return conn, nil
}

func (a *ldapAuth) ServerHandshake(conn net.Conn) (net.Conn, string, error) {
	_ = conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})
	if a.url == "" || a.userDN == "" {
		return nil, "", fmt.Errorf("ldap: server needs url and userdn")
	}
	userB, err := readMsg(conn)
	if err != nil {
		return nil, "", err
	}
	passB, err := readMsg(conn)
	if err != nil {
		return nil, "", err
	}
	user := string(userB)
	if !safeLDAPUser(user) {
		_ = writeMsg(conn, []byte("NO"))
		return nil, "", fmt.Errorf("ldap: invalid username")
	}
	dn := fmt.Sprintf(a.userDN, user)
	if err := ldapBind(a.url, dn, string(passB)); err != nil {
		_ = writeMsg(conn, []byte("NO"))
		return nil, "", fmt.Errorf("ldap: authentication failed: %w", err)
	}
	if err := writeMsg(conn, []byte("OK")); err != nil {
		return nil, "", err
	}
	return conn, user, nil
}

// safeLDAPUser rejects usernames containing DN-special characters, preventing DN
// injection when the username is interpolated into the bind DN template.
func safeLDAPUser(u string) bool {
	if u == "" || len(u) > 256 {
		return false
	}
	return !strings.ContainsAny(u, ",=+<>#;\"\\\x00")
}
