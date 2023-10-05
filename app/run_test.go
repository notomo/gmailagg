package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/app"
	"github.com/notomo/gmailagg/pkg/gcsext/gcstest"
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
		"influxdb": map[string]any{
			"serverUrl": "http://gmailagg-test-influxdb",
			"org":       "test-org",
			"bucket":    "test-bucket",
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

	t.Run("can dry run", func(t *testing.T) {
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

		config, err := app.ReadConfig(ctx, configPath, baseTransport)
		require.NoError(t, err)

		var output bytes.Buffer
		require.NoError(t, app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			config.Influxdb.ServerURL,
			"auth-token",
			config.Influxdb.Org,
			config.Influxdb.Bucket,
			baseTransport,
			&output,
		))

		assert.Equal(t, `{
  "measurement": "measurementName",
  "tags": [
    {
      "key": "tagKey1",
      "value": "tagValue"
    }
  ],
  "fields": [
    {
      "key": "amount",
      "value": 2700
    }
  ],
  "at": "2020-01-02T00:00:00Z"
}
`, output.String())
	})

	t.Run("can run with local token file", func(t *testing.T) {
		transport := httpmock.NewMockTransport()
		defer httpmockext.AssertCalled(t, transport)
		transport.RegisterResponder(googleoauthtest.TokenResponse())
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			matchedMessage,
		)
		transport.RegisterMatcherResponder(
			http.MethodPost,
			"http://gmailagg-test-influxdb/api/v2/write?bucket=test-bucket&org=test-org&precision=ns",
			httpmock.BodyContainsString(`measurementName,tagKey1=tagValue amount=2700i `+gmailtest.ToUnixMilli(t, "2020-01-02T00:00:00Z")),
			httpmock.NewStringResponder(http.StatusOK, `{}`),
		)

		ctx := context.Background()
		baseTransport := app.LogTransport(logDir, transport)

		config, err := app.ReadConfig(ctx, configPath, baseTransport)
		require.NoError(t, err)

		require.NoError(t, app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			config.Influxdb.ServerURL,
			"auth-token",
			config.Influxdb.Org,
			config.Influxdb.Bucket,
			baseTransport,
			nil,
		))
	})

	t.Run("can run with gcs token object", func(t *testing.T) {
		googleoauthtest.SetGoogleApplicationCredentials(t)

		transport := httpmock.NewMockTransport()
		defer httpmockext.AssertCalled(t, transport)
		transport.RegisterResponder(googleoauthtest.TokenResponse())
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			matchedMessage,
		)
		transport.RegisterMatcherResponder(
			http.MethodPost,
			"http://gmailagg-test-influxdb/api/v2/write?bucket=test-bucket&org=test-org&precision=ns",
			httpmock.BodyContainsString(`measurementName,tagKey1=tagValue amount=2700i `+gmailtest.ToUnixMilli(t, "2020-01-02T00:00:00Z")),
			httpmock.NewStringResponder(http.StatusOK, `{}`),
		)
		transport.RegisterResponder(gcstest.GetResponse("test-bucket", "test.json", string(googleoauthtest.TokenJSON())))

		tokenFilePath := "gs://test-bucket/test.json"

		ctx := context.Background()
		baseTransport := app.LogTransport(logDir, transport)

		config, err := app.ReadConfig(ctx, configPath, baseTransport)
		require.NoError(t, err)

		require.NoError(t, app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			config.Influxdb.ServerURL,
			"auth-token",
			config.Influxdb.Org,
			config.Influxdb.Bucket,
			baseTransport,
			nil,
		))
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
		transport.RegisterResponder(gcstest.GetResponse("test-bucket", "test.json", string(googleoauthtest.TokenJSON())))

		tokenFilePath := "gs://test-bucket/test.json"

		ctx := context.Background()
		baseTransport := app.LogTransport(logDir, transport)

		config, err := app.ReadConfig(ctx, configPath, baseTransport)
		require.NoError(t, err)

		runErr := app.Run(
			ctx,
			string(gmailtest.CredentialsJSON()),
			tokenFilePath,
			config.Measurements,
			config.Influxdb.ServerURL,
			"auth-token",
			config.Influxdb.Org,
			config.Influxdb.Bucket,
			baseTransport,
			nil,
		)
		assert.Contains(t, runErr.Error(), "does not matched with")
	})
}
