// Package oauth gives mctop a login flow for OAuth-protected MCP servers:
// discover the server's authorization server, register a client, run the
// authorization-code flow with PKCE in the browser, and cache the resulting
// token so later commands authenticate without logging in again.
package oauth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// Creds is everything needed to use and refresh a token for one server, cached
// on disk between invocations.
type Creds struct {
	ClientID string        `json:"client_id"`
	AuthURL  string        `json:"auth_url"`
	TokenURL string        `json:"token_url"`
	Scopes   []string      `json:"scopes,omitempty"`
	Resource string        `json:"resource,omitempty"`
	Token    *oauth2.Token `json:"token"`
}

// hostKey reduces a server URL to the host used as the cache filename, so the
// MCP endpoint and its origin share one set of credentials.
func hostKey(server string) (string, error) {
	u, err := url.Parse(server)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid server url %q", server)
	}
	return u.Host, nil
}

func storePath(host string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mctop", "oauth", host+".json"), nil
}

// Load returns cached credentials for a server, or (nil, nil) when none exist.
func Load(server string) (*Creds, error) {
	host, err := hostKey(server)
	if err != nil {
		return nil, err
	}
	path, err := storePath(host)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var c Creds
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("read cached credentials: %w", err)
	}
	return &c, nil
}

// Save writes credentials for a server with owner-only permissions, since the
// file holds a bearer token.
func Save(server string, c *Creds) error {
	host, err := hostKey(server)
	if err != nil {
		return err
	}
	path, err := storePath(host)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Delete removes cached credentials for a server, succeeding when none exist.
func Delete(server string) error {
	host, err := hostKey(server)
	if err != nil {
		return err
	}
	path, err := storePath(host)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
