package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	t.Setenv("TZ", "UTC")

	ctx := context.Background()

	path := t.TempDir()

	credentialsJsonPath := filepath.Join(path, "credentials.json")
	require.NoError(t, os.WriteFile(credentialsJsonPath, []byte(`{
  "installed": {
    "client_id": "888888888888-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "project_id": "test",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "redirect_uris": [
      "http://localhost"
    ]
  }
}`), 0700))

	tokenFilePath := filepath.Join(path, "gmailagg/token.json")

	transport := httpmock.NewMockTransport()
	transport.RegisterResponder(http.MethodPost, "https://oauth2.googleapis.com/token",
		httpmock.NewStringResponder(http.StatusOK, `{
  "access_token": "XXXX.XXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "expires_in": 3599,
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "scope": "https://www.googleapis.com/auth/gmail.readonly",
  "token_type": "Bearer"
}`),
	)

	require.NoError(t, Authorize(
		ctx,
		credentialsJsonPath,
		tokenFilePath,
		&gmailtest.Opener{AuthCode: "test"},
		LogTransport("/tmp/gmailaggtest", transport),
	))

	fileName := "gmailagg/token.json"
	tmpfs := os.DirFS(path)
	require.NoError(t, fstest.TestFS(tmpfs, fileName))

	f, err := tmpfs.Open(fileName)
	require.NoError(t, err)

	got, err := io.ReadAll(f)
	require.NoError(t, err)

	// ignore expiry (depends time.Now())
	assert.Contains(t, string(got), `{
  "access_token": "XXXX.XXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "token_type": "Bearer",
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",`, string(got))
}
