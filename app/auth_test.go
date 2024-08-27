package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/app"
	"github.com/notomo/gmailagg/pkg/fstestext"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/notomo/gmailagg/pkg/googleoauthtest"
	"github.com/notomo/gmailagg/pkg/httpmockext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	t.Setenv("TZ", "UTC")

	logDir := "/tmp/gmailaggtest"
	defaultOpener := &gmailtest.Opener{AuthCode: "test"}
	defaultTimeout := 3 * time.Minute
	defaultPort := uint(0)

	t.Run("can save as local file", func(t *testing.T) {
		ctx := context.Background()

		tmpDir := t.TempDir()

		tokenFileName := "gmailagg/token.json"
		tokenFilePath := filepath.Join(tmpDir, tokenFileName)

		transport := httpmock.NewMockTransport()
		defer httpmockext.AssertCalled(t, transport)
		transport.RegisterResponder(googleoauthtest.TokenResponse())

		require.NoError(t, app.Authorize(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			defaultOpener,
			defaultTimeout,
			defaultPort,
			app.LogTransport(logDir, transport),
			false,
		))

		var got map[string]string
		tokenJSON := fstestext.GetFileContent(t, os.DirFS(tmpDir), tokenFileName)
		require.NoError(t, json.Unmarshal(tokenJSON, &got))

		want := googleoauthtest.Token(t)
		// ignore expiry (depends time.Now())
		delete(want, "expiry")
		delete(got, "expiry")

		assert.Equal(t, want, got)
	})
}
