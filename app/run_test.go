package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/app"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/notomo/gmailagg/pkg/googleoauthtest"
	"github.com/notomo/gmailagg/pkg/httpmockext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("TZ", "UTC")

	logDir := "/tmp/gmailaggtest"

	path := t.TempDir()

	tokenFilePath := filepath.Join(path, "token.json")
	require.NoError(t, os.WriteFile(tokenFilePath, googleoauthtest.TokenJSON(), 0700))

	configMap := map[string]any{
		"measurements": []map[string]any{
			{
				"name":  "measurementName",
				"query": "measurement_query",
				"aggregations": []map[string]any{
					{
						"query": "query",
						"rules": []map[string]any{
							{
								"type":          "regexp",
								"target":        "body",
								"pattern":       `金額.*￥ (?P<price>[\d,]+) 割引.*￥ (?P<discount>[\d,]+)`,
								"matchMaxCount": -1,
								"mappings": map[string]any{
									"amount": map[string]any{
										"type":       "field",
										"dataType":   "integer",
										"expression": "price-discount",
									},
									"price": map[string]any{
										"type":     "hidden",
										"dataType": "integer",
										"replacers": []map[string]any{
											{
												"old": ",",
												"new": "",
											},
										},
									},
									"discount": map[string]any{
										"type":     "hidden",
										"dataType": "integer",
									},
								},
							},
						},
						"tags": map[string]any{
							"tagKey1": "tagValue",
						},
					},
				},
			},
		},
	}
	configBytes, err := json.Marshal(configMap)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(configPath, configBytes, 0700))

	matchedMessage := gmailtest.Message{
		ID:       "1111111111111111",
		ThreadID: "ttttttttttttttt1",
		Body: `
金額 ￥ 1,000 割引 ￥ 100
金額 ￥ 2,000 割引 ￥ 200
`,
		Timestamp: "2020-01-02T00:00:00Z",
	}

	t.Run("can run", func(t *testing.T) {
		transport := httpmock.NewMockTransport()
		defer httpmockext.AssertCalled(t, transport)
		transport.RegisterResponder(googleoauthtest.TokenResponse())
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			matchedMessage,
		)

		ctx := context.Background()
		baseTransport := app.LogTransport(logDir, transport)

		config, err := app.ReadConfig(configPath)
		require.NoError(t, err)

		var output bytes.Buffer
		require.NoError(t, app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			baseTransport,
			&output,
		))

		assert.Equal(t, `{
  "measurement": "measurementName",
  "tags": {
    "tagKey1": "tagValue"
  },
  "fields": {
    "amount": 2700
  },
  "at": "2020-01-02T00:00:00Z"
}
`, output.String())
	})

	t.Run("raises error if rule does not match with mail body", func(t *testing.T) {
		googleoauthtest.SetGoogleApplicationCredentials(t)

		transport := httpmock.NewMockTransport()
		defer httpmockext.AssertCalled(t, transport)
		transport.RegisterResponder(googleoauthtest.TokenResponse())
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			gmailtest.Message{
				ID:       "1111111111111111",
				ThreadID: "ttttttttttttttt1",
				Body: `
others
`,
				Timestamp: "2020-01-02T00:00:00Z",
			},
		)

		ctx := context.Background()
		baseTransport := app.LogTransport(logDir, transport)

		config, err := app.ReadConfig(configPath)
		require.NoError(t, err)

		output := &bytes.Buffer{}
		runErr := app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			baseTransport,
			output,
		)
		assert.Contains(t, runErr.Error(), "does not matched with")
	})
}
