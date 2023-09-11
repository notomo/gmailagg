package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/pkg/gcsext/gcstest"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("TZ", "UTC")

	path := t.TempDir()

	tokenFilePath := filepath.Join(path, "token.json")
	require.NoError(t, os.WriteFile(tokenFilePath, gmailtest.TokenJSON(), 0700))

	configMap := map[string]any{
		"measurements": []map[string]any{
			{
				"name": "measurementName",
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
	configStr := string(configBytes)

	t.Run("can dry run", func(t *testing.T) {
		transport := httpmock.NewMockTransport()
		gmailtest.RegisterTokenResponse(transport)
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			gmailtest.Message{
				ID:       "1111111111111111",
				ThreadID: "ttttttttttttttt1",
				Body: `
金額 ￥ 1,000 割引 ￥ 100
金額 ￥ 2,000 割引 ￥ 200
`,
				Timestamp: "2020-01-02T00:00:00Z",
			},
			gmailtest.Message{
				ID:        "2222222222222222",
				ThreadID:  "ttttttttttttttt2",
				Body:      `others`,
				Timestamp: "2020-01-03T00:00:00Z",
			},
		)

		ctx := context.Background()
		baseTransport := LogTransport("/tmp/gmailaggtest", transport)

		config, err := ReadConfig(ctx, "", configStr, baseTransport)
		require.NoError(t, err)

		var output bytes.Buffer
		require.NoError(t, Run(
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
		gmailtest.RegisterTokenResponse(transport)
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			gmailtest.Message{
				ID:        "1111111111111111",
				ThreadID:  "ttttttttttttttt1",
				Body:      `合計 ￥ 100`,
				Timestamp: "2020-01-02T00:00:00Z",
			},
			gmailtest.Message{
				ID:        "2222222222222222",
				ThreadID:  "ttttttttttttttt2",
				Body:      `others`,
				Timestamp: "2020-01-03T00:00:00Z",
			},
		)
		transport.RegisterMatcherResponder(
			http.MethodPost,
			"http://gmailagg-test-influxdb/api/v2/write?bucket=test-bucket&org=test-org&precision=ns",
			httpmock.BodyContainsString(`measurementName,tagKey1=tagValue amount=100i `+gmailtest.ToUnixMilli(t, "2020-01-02T00:00:00Z")),
			httpmock.NewStringResponder(http.StatusOK, `{}`),
		)

		ctx := context.Background()
		baseTransport := LogTransport("/tmp/gmailaggtest", transport)

		config, err := ReadConfig(ctx, "", configStr, baseTransport)
		require.NoError(t, err)

		require.NoError(t, Run(
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
		tmpDir := t.TempDir()

		credentialsFilePath := filepath.Join(tmpDir, "application_default_credentials.json")
		require.NoError(t, os.WriteFile(credentialsFilePath, gcstest.CredentialsJSON(), 0700))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFilePath)

		transport := httpmock.NewMockTransport()
		gmailtest.RegisterTokenResponse(transport)
		gmailtest.RegisterMessageResponse(
			t,
			transport,
			gmailtest.Message{
				ID:        "1111111111111111",
				ThreadID:  "ttttttttttttttt1",
				Body:      `合計 ￥ 100`,
				Timestamp: "2020-01-02T00:00:00Z",
			},
			gmailtest.Message{
				ID:        "2222222222222222",
				ThreadID:  "ttttttttttttttt2",
				Body:      `others`,
				Timestamp: "2020-01-03T00:00:00Z",
			},
		)
		transport.RegisterMatcherResponder(
			http.MethodPost,
			"http://gmailagg-test-influxdb/api/v2/write?bucket=test-bucket&org=test-org&precision=ns",
			httpmock.BodyContainsString(`measurementName,tagKey1=tagValue amount=100i `+gmailtest.ToUnixMilli(t, "2020-01-02T00:00:00Z")),
			httpmock.NewStringResponder(http.StatusOK, `{}`),
		)
		gcstest.RegisterGetResponse(transport, "test-bucket", "test.json", string(gmailtest.TokenJSON()))

		tokenFilePath := "gs://test-bucket/test.json"

		ctx := context.Background()
		baseTransport := LogTransport("/tmp/gmailaggtest", transport)

		config, err := ReadConfig(ctx, "", configStr, baseTransport)
		require.NoError(t, err)

		require.NoError(t, Run(
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
}
