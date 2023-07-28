package gmailtest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

type Message struct {
	ID        string
	ThreadID  string
	Body      string
	Timestamp string
}

func RegisterMessageResponse(
	t *testing.T,
	transport *httpmock.MockTransport,
	messages ...Message,
) {
	t.Helper()

	{
		summaries := []map[string]any{}
		for _, message := range messages {
			summaries = append(summaries, map[string]any{
				"id":       message.ID,
				"threadId": message.ThreadID,
			})
		}

		var body bytes.Buffer
		encorder := json.NewEncoder(&body)
		encorder.SetIndent("", "  ")
		if err := encorder.Encode(map[string]any{
			"messages":           summaries,
			"resultSizeEstimate": len(summaries),
		}); err != nil {
			t.Fatal(err)
		}

		transport.RegisterResponder(
			http.MethodGet,
			"https://gmail.googleapis.com/gmail/v1/users/me/messages",
			httpmock.NewStringResponder(http.StatusOK, body.String()),
		)
	}

	for _, message := range messages {
		time, err := time.Parse(time.RFC3339, message.Timestamp)
		if err != nil {
			t.Fatal(err)
		}
		unixMilli := strconv.FormatInt(time.UnixMilli(), 10)

		bodyData := base64.URLEncoding.EncodeToString([]byte(message.Body))

		var body bytes.Buffer
		encorder := json.NewEncoder(&body)
		encorder.SetIndent("", "  ")
		if err := encorder.Encode(map[string]any{
			"payload": map[string]any{
				"body": map[string]any{
					"size": 0,
					"data": "",
				},
				"parts": []map[string]any{
					{
						"partId":   "0",
						"mimeType": "text/plain",
						"filename": "",
						"headers": []map[string]any{
							{
								"name":  "Content-Type",
								"value": "text/plain; charset=utf-8",
							},
							{
								"name":  "Content-Transfer-Encoding",
								"value": "base64",
							},
						},
						"body": map[string]any{
							"size": 0,
							"data": bodyData,
						},
					},
				},
			},
			"internalDate": unixMilli,
		}); err != nil {
			t.Fatal(err)
		}

		transport.RegisterResponder(
			http.MethodGet,
			"https://gmail.googleapis.com/gmail/v1/users/me/messages/"+message.ID,
			httpmock.NewStringResponder(http.StatusOK, body.String()),
		)
	}
}
