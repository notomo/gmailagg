package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/pkg/fstestext"
	"github.com/notomo/gmailagg/pkg/gcsext/gcstest"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	t.Setenv("TZ", "UTC")

	t.Run("can save as local file", func(t *testing.T) {
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
			3*time.Minute,
			0,
			LogTransport("/tmp/gmailaggtest", transport),
			false,
		))

		var got map[string]string
		tokenJSON := fstestext.GetFileContent(t, os.DirFS(tmpDir), tokenFileName)
		require.NoError(t, json.Unmarshal(tokenJSON, &got))

		want := gmailtest.Token(t)
		// ignore expiry (depends time.Now())
		delete(want, "expiry")
		delete(got, "expiry")

		assert.Equal(t, want, got)
	})

	t.Run("can save as gcs object", func(t *testing.T) {
		tmpDir := t.TempDir()

		credentialsFilePath := filepath.Join(tmpDir, "application_default_credentials.json")
		require.NoError(t, os.WriteFile(credentialsFilePath, gcstest.CredentialsJSON(), 0700))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFilePath)

		ctx := context.Background()

		tokenFilePath := "gs://test-bucket/test.json"

		token := gmailtest.Token(t)
		requestBodyShouldContains := fmt.Sprintf(`"refresh_token": "%s"`, token["refresh_token"])

		transport := httpmock.NewMockTransport()
		gmailtest.RegisterTokenResponse(transport)
		gcstest.RegisterUploadResponse(t, transport, "test-bucket", "test.json", requestBodyShouldContains)

		require.NoError(t, Authorize(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			&gmailtest.Opener{AuthCode: "test"},
			3*time.Minute,
			0,
			LogTransport("/tmp/gmailaggtest", transport),
			false,
		))
	})

	t.Run("does not save gcs object on timeout", func(t *testing.T) {
		tmpDir := t.TempDir()

		credentialsFilePath := filepath.Join(tmpDir, "application_default_credentials.json")
		require.NoError(t, os.WriteFile(credentialsFilePath, gcstest.CredentialsJSON(), 0700))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFilePath)

		ctx := context.Background()

		tokenFilePath := "gs://test-bucket/test.json"

		transport := httpmock.NewMockTransport()
		gmailtest.RegisterTokenResponse(transport)

		assert.ErrorIs(t, Authorize(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			&gmailtest.Opener{AuthCode: "test"},
			0*time.Minute,
			0,
			LogTransport("/tmp/gmailaggtest", transport),
			false,
		), context.Canceled)
	})
}
