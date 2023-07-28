package gmailtest

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TokenJSON() []byte {
	return []byte(`{
  "access_token": "XXXX.XXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "token_type": "Bearer",
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "expiry": "2000-01-01T00:00:00.0000000+09:00"
}`)
}

func Token(t *testing.T) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal(TokenJSON(), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func RegisterTokenResponse(transport *httpmock.MockTransport) {
	transport.RegisterResponder(
		http.MethodPost,
		"https://oauth2.googleapis.com/token",
		httpmock.NewStringResponder(http.StatusOK, `{
  "access_token": "XXXX.XXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "expires_in": 3599,
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "scope": "https://www.googleapis.com/auth/gmail.readonly",
  "token_type": "Bearer"
}`),
	)
}
