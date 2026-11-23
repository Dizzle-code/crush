// Package claude implements OAuth authentication against Anthropic's
// authorization server, letting users sign in with a Claude Pro/Max
// subscription. The flow is PKCE based; the user pastes the code shown on
// Anthropic's callback page back into Crush.
package claude

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/oauth"
)

const (
	ClientID    = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	RedirectURI = "https://console.anthropic.com/oauth/code/callback"
	Scope       = "org:create_api_key user:profile user:inference"

	// SystemPromptPrefix must lead the system prompt for subscription
	// tokens to be accepted by the API.
	SystemPromptPrefix = "You are Claude Code, Anthropic's official CLI for Claude."

	// BetaHeader enables OAuth bearer tokens on the Messages API.
	BetaHeader = "oauth-2025-04-20"
)

// Variables so tests can point them at a stub server.
var (
	authorizeEndpoint = "https://claude.ai/oauth/authorize"
	tokenEndpoint     = "https://console.anthropic.com/v1/oauth/token"
	HTTPClient        = &http.Client{Timeout: 30 * time.Second}
)

// Flow holds the PKCE verifier and state of one authorization attempt.
type Flow struct {
	verifier string
	state    string
}

// NewFlow generates fresh PKCE and state values.
func NewFlow() (*Flow, error) {
	verifier, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	state, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	return &Flow{verifier: verifier, state: state}, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AuthorizeURL returns the URL to open in the browser.
func (f *Flow) AuthorizeURL() string {
	sum := sha256.Sum256([]byte(f.verifier))
	vals := url.Values{
		"code":                  {"true"},
		"client_id":             {ClientID},
		"response_type":         {"code"},
		"redirect_uri":          {RedirectURI},
		"scope":                 {Scope},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
		"state":                 {f.state},
	}
	return authorizeEndpoint + "?" + vals.Encode()
}

// Exchange trades the pasted "code#state" string for a token.
func (f *Flow) Exchange(ctx context.Context, input string) (*oauth.Token, error) {
	code, state, _ := strings.Cut(strings.TrimSpace(input), "#")
	if code == "" {
		return nil, errors.New("empty authorization code")
	}
	if state != "" && state != f.state {
		return nil, errors.New("authorization state mismatch")
	}
	token, err := requestToken(ctx, map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"state":         f.state,
		"client_id":     ClientID,
		"redirect_uri":  RedirectURI,
		"code_verifier": f.verifier,
	})
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	return token, nil
}

// RefreshToken exchanges a refresh token for a fresh token pair.
func RefreshToken(ctx context.Context, refreshToken string) (*oauth.Token, error) {
	token, err := requestToken(ctx, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     ClientID,
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, nil
}

func requestToken(ctx context.Context, payload map[string]string) (*oauth.Token, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &oauth.TokenExchangeError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var token oauth.Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return nil, errors.New("token response contained no access token")
	}
	token.SetExpiresAt()
	return &token, nil
}
