package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/pkg/fstestext"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	t.Setenv("TZ", "UTC")

	ctx := context.Background()

	tmpDir := t.TempDir()

	tokenFileName := "gmailagg/token.json"
	tokenFilePath := filepath.Join(tmpDir, tokenFileName)

	transport := httpmock.NewMockTransport()
	gmailtest.RegisterTokenResponse(transport)

	require.NoError(t, Authorize(
		ctx,
		string(gmailtest.CredentialsJSON()),
		tokenFilePath,
		&gmailtest.Opener{AuthCode: "test"},
		LogTransport("/tmp/gmailaggtest", transport),
	))

	var got map[string]string
	tokenJSON := fstestext.GetFileContent(t, os.DirFS(tmpDir), tokenFileName)
	require.NoError(t, json.Unmarshal(tokenJSON, &got))

	want := gmailtest.Token(t)
	// ignore expiry (depends time.Now())
	delete(want, "expiry")
	delete(got, "expiry")

	assert.Equal(t, want, got)
}
