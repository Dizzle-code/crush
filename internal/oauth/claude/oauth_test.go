package claude

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func stubToken(t *testing.T, got *map[string]string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(got))
		_, _ = w.Write([]byte(`{"access_token":"at","refresh_token":"rt","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)
	old := tokenEndpoint
	tokenEndpoint = srv.URL
	t.Cleanup(func() { tokenEndpoint = old })
}

func TestAuthorizeURLAndExchange(t *testing.T) {
	var got map[string]string
	stubToken(t, &got)

	f, err := NewFlow()
	require.NoError(t, err)
	u, err := url.Parse(f.AuthorizeURL())
	require.NoError(t, err)
	require.Equal(t, ClientID, u.Query().Get("client_id"))
	require.Equal(t, "S256", u.Query().Get("code_challenge_method"))

	token, err := f.Exchange(t.Context(), " abc#"+f.state+"\n")
	require.NoError(t, err)
	require.Equal(t, "at", token.AccessToken)
	require.Equal(t, "abc", got["code"])
	require.Equal(t, f.verifier, got["code_verifier"])

	_, err = f.Exchange(t.Context(), "abc#wrong")
	require.Error(t, err)
}

func TestRefreshKeepsOldRefreshToken(t *testing.T) {
	var got map[string]string
	stubToken(t, &got)
	token, err := RefreshToken(t.Context(), "old")
	require.NoError(t, err)
	require.Equal(t, "refresh_token", got["grant_type"])
	require.Equal(t, "rt", token.RefreshToken)
}
