package app

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/app/extractor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("TZ", "UTC")

	ctx := context.Background()

	path := t.TempDir()

	// TODO: test helper

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

	tokenFilePath := filepath.Join(path, "token.json")
	require.NoError(t, os.WriteFile(tokenFilePath, []byte(`{
  "access_token": "XXXX.XXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "token_type": "Bearer",
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "expiry": "2000-01-01T00:00:00.0000000+09:00"
}`), 0700))

	t.Run("dry run", func(t *testing.T) {
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

		transport.RegisterResponder(http.MethodGet, "https://gmail.googleapis.com/gmail/v1/users/me/messages",
			httpmock.NewStringResponder(http.StatusOK, `{
  "messages": [
    {
      "id": "1111111111111111",
      "threadId": "ttttttttttttttt1"
    },
    {
      "id": "2222222222222222",
      "threadId": "ttttttttttttttt2"
    }
  ],
  "resultSizeEstimate": 2
}`),
		)

		transport.RegisterResponder(http.MethodGet, "https://gmail.googleapis.com/gmail/v1/users/me/messages/1111111111111111",
			httpmock.NewStringResponder(http.StatusOK, `{
  "payload": {
    "body": {
      "size": 0,
      "data": ""
    },
    "parts": [
      {
        "partId": "0",
        "mimeType": "text/plain",
        "filename": "",
        "headers": [
          {
            "name": "Content-Type",
            "value": "text/plain; charset=utf-8"
          },
          {
            "name": "Content-Transfer-Encoding",
            "value": "base64"
          }
        ],
        "body": {
          "size": 0,
          "data": "5ZCI6KiILirvv6UgMTAw"
        }
      }
    ]
  },
  "internalDate": "1577923200000"
}`),
		)

		transport.RegisterResponder(http.MethodGet, "https://gmail.googleapis.com/gmail/v1/users/me/messages/2222222222222222",
			httpmock.NewStringResponder(http.StatusOK, `{
  "payload": {
    "body": {
      "size": 0,
      "data": ""
    },
    "parts": [
      {
        "partId": "0",
        "mimeType": "text/plain",
        "filename": "",
        "headers": [
          {
            "name": "Content-Type",
            "value": "text/plain; charset=utf-8"
          },
          {
            "name": "Content-Transfer-Encoding",
            "value": "base64"
          }
        ],
        "body": {
          "size": 0,
          "data": "aWdub3JlZA=="
        }
      }
    ]
  },
  "internalDate": "1577923200000"
}`),
		)

		var buf bytes.Buffer

		measurements := []extractor.Measurement{
			{
				Name: "measurementName",
				Aggregations: []extractor.Aggregation{
					{
						Query: "query",
						Rules: []extractor.AggregationRule{
							{
								Type:    extractor.RuleTypeRegexp,
								Target:  extractor.TargetTypeBody,
								Pattern: `合計.*￥ (?P<amount>\d+)`,
								Mappings: map[string]extractor.RuleMapping{
									"amount": {
										Type: extractor.RuleMappingTypeField,
									},
								},
							},
						},
						Tags: map[string]string{
							"tagKey1": "tagValue",
						},
					},
				},
			},
		}

		influxdbServerURL := ""
		influxdbAuthToken := ""
		influxdbOrg := ""
		influxdbBucket := ""

		require.NoError(t, Run(
			ctx,
			credentialsJsonPath,
			tokenFilePath,
			measurements,
			influxdbServerURL,
			influxdbAuthToken,
			influxdbOrg,
			influxdbBucket,
			LogTransport("/tmp/gmailaggtest", transport),
			&buf,
		))

		assert.Equal(t, `{
  "Measurement": "measurementName",
  "Tags": [
    {
      "Key": "tagKey1",
      "Value": "tagValue"
    }
  ],
  "Fields": [
    {
      "Key": "amount",
      "Value": 100
    }
  ],
  "At": "2020-01-02T00:00:00Z"
}
`, buf.String())
	})

	t.Run("run", func(t *testing.T) {
		t.Skip("TODO")
	})
}
