package app

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("TZ", "UTC")

	ctx := context.Background()

	path := t.TempDir()

	tokenFilePath := filepath.Join(path, "token.json")
	require.NoError(t, os.WriteFile(tokenFilePath, gmailtest.TokenJSON(), 0700))

	configPath := filepath.Join(path, "config.yaml")
	configContent := `
measurements:
  - name: measurementName
    aggregations:
      - query: query
        rules:
          - type: regexp
            target: body
            pattern: 合計.*￥ (?P<amount>\d+)
            mappings:
              amount:
                type: field
                data_type: integer
        tags:
          tagKey1: tagValue

influxdb:
  serverUrl: http://gmailagg-test-influxdb
  org: test-org
  bucket: test-bucket
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0700))

	t.Run("dry run", func(t *testing.T) {
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

		config, err := ReadConfig(configPath)
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
			LogTransport("/tmp/gmailaggtest", transport),
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
      "value": 100
    }
  ],
  "at": "2020-01-02T00:00:00Z"
}
`, output.String())
	})

	t.Run("run", func(t *testing.T) {
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

		config, err := ReadConfig(configPath)
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
			LogTransport("/tmp/gmailaggtest", transport),
			nil,
		))
	})
}
