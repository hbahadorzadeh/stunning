package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// introspectionServer returns an RFC 7662 endpoint that marks goodToken active.
func introspectionServer(t *testing.T, goodToken, sub, scope string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		active := r.PostForm.Get("token") == goodToken
		resp := map[string]any{"active": active}
		if active {
			resp["sub"] = sub
			resp["scope"] = scope
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestOAuthAuthActive(t *testing.T) {
	srv := introspectionServer(t, "good-token", "dave", "tunnel read")
	defer srv.Close()
	clientSpec := "oauth?token=good-token"
	serverSpec := fmt.Sprintf("oauth?introspect=%s&client_id=cid&client_secret=sec", srv.URL)
	id, cErr, sErr := runAuth(t, clientSpec, serverSpec)
	if cErr != nil || sErr != nil {
		t.Fatalf("oauth active should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "dave" {
		t.Fatalf("identity=%q want dave", id)
	}
}

func TestOAuthAuthInactive(t *testing.T) {
	srv := introspectionServer(t, "good-token", "dave", "")
	defer srv.Close()
	clientSpec := "oauth?token=bad-token"
	serverSpec := fmt.Sprintf("oauth?introspect=%s", srv.URL)
	_, cErr, sErr := runAuth(t, clientSpec, serverSpec)
	if sErr == nil || cErr == nil {
		t.Fatal("inactive token must be rejected")
	}
}

func TestOAuthAuthScopeMissing(t *testing.T) {
	srv := introspectionServer(t, "good-token", "dave", "read")
	defer srv.Close()
	clientSpec := "oauth?token=good-token"
	serverSpec := fmt.Sprintf("oauth?introspect=%s&scope=admin", srv.URL)
	_, cErr, sErr := runAuth(t, clientSpec, serverSpec)
	if sErr == nil || cErr == nil {
		t.Fatal("token without required scope must be rejected")
	}
}

func TestLDAPAuth(t *testing.T) {
	orig := ldapBind
	defer func() { ldapBind = orig }()
	ldapBind = func(serverURL, dn, password string) error {
		if dn == "uid=alice,ou=people,dc=example,dc=org" && password == "s3cret" {
			return nil
		}
		return fmt.Errorf("fake-ldap: invalid credentials")
	}

	server := "ldap?url=ldap://dir:389&userdn=uid=%s,ou=people,dc=example,dc=org"
	id, cErr, sErr := runAuth(t, "ldap?user=alice&password=s3cret", server)
	if cErr != nil || sErr != nil {
		t.Fatalf("ldap should succeed: client=%v server=%v", cErr, sErr)
	}
	if id != "alice" {
		t.Fatalf("identity=%q want alice", id)
	}

	// wrong password rejected
	_, cErr, sErr = runAuth(t, "ldap?user=alice&password=wrong", server)
	if sErr == nil || cErr == nil {
		t.Fatal("wrong ldap password must be rejected")
	}
}

func TestLDAPInjectionRejected(t *testing.T) {
	orig := ldapBind
	defer func() { ldapBind = orig }()
	bound := false
	ldapBind = func(_, _, _ string) error { bound = true; return nil }

	server := "ldap?url=ldap://dir:389&userdn=uid=%s,ou=people,dc=example,dc=org"
	// username with DN-special characters must be rejected before any bind.
	_, _, sErr := runAuth(t, "ldap?user=a,dc=evil&password=x", server)
	if sErr == nil {
		t.Fatal("username with DN metacharacters must be rejected")
	}
	if bound {
		t.Fatal("must not attempt a bind with an unsafe username")
	}
}

func TestSafeLDAPUser(t *testing.T) {
	good := []string{"alice", "bob.smith", "user-1", "u_2@corp"}
	bad := []string{"", "a,b", "a=b", "a\\b", "a\x00b", "cn=admin"}
	for _, u := range good {
		if !safeLDAPUser(u) {
			t.Errorf("safeLDAPUser(%q) = false, want true", u)
		}
	}
	for _, u := range bad {
		if safeLDAPUser(u) {
			t.Errorf("safeLDAPUser(%q) = true, want false", u)
		}
	}
}

func TestLDAPUserDNValidation(t *testing.T) {
	// A userdn without exactly one %s would yield a static DN (auth bypass).
	for _, bad := range []string{"ldap?url=ldap://x&userdn=cn=admin,dc=x", "ldap?url=ldap://x&userdn=uid=%s%s,dc=x"} {
		if _, err := ParseAuth(bad); err == nil {
			t.Errorf("ParseAuth(%q) should error", bad)
		}
	}
	if _, err := ParseAuth("ldap?url=ldap://x&userdn=uid=%s,ou=users,dc=x"); err != nil {
		t.Fatalf("valid userdn rejected: %v", err)
	}
}

func TestLDAPEmptyPasswordRejected(t *testing.T) {
	orig := ldapBind
	defer func() { ldapBind = orig }()
	bound := false
	ldapBind = func(_, _, _ string) error { bound = true; return nil }
	server := "ldap?url=ldap://x&userdn=uid=%s,ou=users,dc=x"
	_, _, sErr := runAuth(t, "ldap?user=alice&password=", server)
	if sErr == nil {
		t.Fatal("empty password must be rejected (avoids unauthenticated bind)")
	}
	if bound {
		t.Fatal("must not attempt a bind with an empty password")
	}
}
