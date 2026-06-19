package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"golang.org/x/oauth2"
)

// Login runs the full OAuth authorization-code flow for an MCP server and
// returns credentials ready to cache: it discovers the authorization server,
// registers a client dynamically, opens the browser for the user to approve,
// and exchanges the returned code for a token using PKCE.
func Login(ctx context.Context, server string) (*Creds, error) {
	origin, err := originOf(server)
	if err != nil {
		return nil, err
	}

	prm, err := oauthex.GetProtectedResourceMetadata(ctx, origin+"/.well-known/oauth-protected-resource", origin, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("discover protected resource: %w", err)
	}
	if len(prm.AuthorizationServers) == 0 {
		return nil, fmt.Errorf("server advertises no authorization servers")
	}
	issuer := prm.AuthorizationServers[0]

	meta, err := oauthex.GetAuthServerMeta(ctx, issuer+"/.well-known/oauth-authorization-server", issuer, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("discover authorization server: %w", err)
	}
	if meta.RegistrationEndpoint == "" {
		return nil, fmt.Errorf("authorization server does not support dynamic client registration")
	}

	// Bind the callback listener first so the redirect URI is known before
	// registration, which must record that exact URI.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("open callback listener: %w", err)
	}
	defer listener.Close()
	redirect := fmt.Sprintf("http://%s/callback", listener.Addr().String())

	reg, err := oauthex.RegisterClient(ctx, meta.RegistrationEndpoint, &oauthex.ClientRegistrationMetadata{
		ClientName:              "mctop",
		RedirectURIs:            []string{redirect},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
		Scope:                   join(prm.ScopesSupported),
	}, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("register client: %w", err)
	}

	conf := &oauth2.Config{
		ClientID:    reg.ClientID,
		RedirectURL: redirect,
		Scopes:      prm.ScopesSupported,
		Endpoint:    oauth2.Endpoint{AuthURL: meta.AuthorizationEndpoint, TokenURL: meta.TokenEndpoint},
	}

	verifier := oauth2.GenerateVerifier()
	state, err := randomState()
	if err != nil {
		return nil, err
	}
	// The resource indicator (RFC 8707) binds the token to this MCP server, as
	// the MCP authorization spec requires.
	resourceParam := oauth2.SetAuthURLParam("resource", prm.Resource)
	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier), resourceParam)

	code, err := awaitCode(ctx, listener, state, authURL)
	if err != nil {
		return nil, err
	}

	token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier), resourceParam)
	if err != nil {
		return nil, fmt.Errorf("exchange code for token: %w", err)
	}

	return &Creds{
		ClientID: reg.ClientID,
		AuthURL:  meta.AuthorizationEndpoint,
		TokenURL: meta.TokenEndpoint,
		Scopes:   prm.ScopesSupported,
		Resource: prm.Resource,
		Token:    token,
	}, nil
}

// AccessToken returns a valid access token, refreshing it when expired, and
// reports whether the token changed so the caller can persist the new one.
func (c *Creds) AccessToken(ctx context.Context) (token string, changed bool, err error) {
	conf := &oauth2.Config{
		ClientID: c.ClientID,
		Scopes:   c.Scopes,
		Endpoint: oauth2.Endpoint{AuthURL: c.AuthURL, TokenURL: c.TokenURL},
	}
	fresh, err := conf.TokenSource(ctx, c.Token).Token()
	if err != nil {
		return "", false, err
	}
	changed = fresh.AccessToken != c.Token.AccessToken
	c.Token = fresh
	return fresh.AccessToken, changed, nil
}

func originOf(server string) (string, error) {
	u, err := url.Parse(server)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid server url %q", server)
	}
	return u.Scheme + "://" + u.Host, nil
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func join(scopes []string) string {
	out := ""
	for i, s := range scopes {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}
