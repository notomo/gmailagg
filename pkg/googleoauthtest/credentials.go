package googleoauthtest

import (
	"os"
	"path/filepath"
	"testing"
)

func CredentialsJSON() []byte {
	return []byte(`{
  "client_id": "xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
  "client_secret": "x-xxxxxxxxxxxxxxxxxxxxxx",
  "quota_project_id": "test",
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "type": "authorized_user"
}`)
}

func SetGoogleApplicationCredentials(t *testing.T) {
	t.Helper()

	dirPath := t.TempDir()
	credentialsFilePath := filepath.Join(dirPath, "application_default_credentials.json")
	if err := os.WriteFile(credentialsFilePath, CredentialsJSON(), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFilePath)
}
