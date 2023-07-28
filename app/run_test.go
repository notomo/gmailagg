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
	"github.com/notomo/gmailagg/pkg/gmailext/gmailtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("TZ", "UTC")

	ctx := context.Background()

	path := t.TempDir()

	credentialsJsonPath := filepath.Join(path, "credentials.json")
	require.NoError(t, os.WriteFile(credentialsJsonPath, gmailtest.CredentialsJSON(), 0700))

	tokenFilePath := filepath.Join(path, "token.json")
	require.NoError(t, os.WriteFile(tokenFilePath, gmailtest.TokenJSON(), 0700))

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
										Type:     extractor.RuleMappingTypeField,
										DataType: extractor.RuleMappingFieldDataTypeInteger,
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

		var output bytes.Buffer
		require.NoError(t, Run(
			ctx,
			credentialsJsonPath,
			tokenFilePath,
			measurements,
			"",
			"",
			"",
			"",
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
										Type:     extractor.RuleMappingTypeField,
										DataType: extractor.RuleMappingFieldDataTypeInteger,
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

		influxdbServerURL := "http://gmailagg-test-influxdb"
		influxdbAuthToken := "auth-token"
		influxdbOrg := "test-org"
		influxdbBucket := "test-bucket"

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
			nil,
		))
	})
}
