package gcstest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func UploadResponse(
	t *testing.T,
	bucket string,
	object string,
	requestBodyShouldContains string,
) (string, string, httpmock.Matcher, httpmock.Responder) {
	t.Helper()

	u := fmt.Sprintf(
		"https://storage.googleapis.com/upload/storage/v1/b/%s/o?alt=json&name=%s&prettyPrint=false&projection=full&uploadType=multipart",
		bucket,
		object,
	)

	var responseBody bytes.Buffer
	encorder := json.NewEncoder(&responseBody)
	encorder.SetIndent("", "  ")
	if err := encorder.Encode(map[string]any{
		"kind":                    "storage#object",
		"id":                      fmt.Sprintf("%s/%s/8888888888888888", bucket, object),
		"selfLink":                fmt.Sprintf("https://www.googleapis.com/storage/v1/b/%s/o/%s", bucket, object),
		"mediaLink":               fmt.Sprintf("https://storage.googleapis.com/download/storage/v1/b/%s/o/%s?generation=8888888888888888&alt=media", bucket, object),
		"name":                    object,
		"bucket":                  bucket,
		"generation":              "8888888888888888",
		"metageneration":          "1",
		"contentType":             "text/plain; charset=utf-8",
		"storageClass":            "STANDARD",
		"size":                    "8888",
		"md5Hash":                 "8888888888888888888888==",
		"crc32c":                  "888888==",
		"etag":                    "8888888/8888888=",
		"timeCreated":             "2020-01-01T00:00:00.000Z",
		"updated":                 "2020-01-01T00:00:00.000Z",
		"timeStorageClassUpdated": "2020-01-01T00:00:00.000Z",
	}); err != nil {
		t.Fatal(err)
	}

	return http.MethodPost,
		u,
		httpmock.BodyContainsString(requestBodyShouldContains),
		httpmock.NewStringResponder(http.StatusOK, responseBody.String())
}

func GetResponse(
	bucket string,
	object string,
	body string,
) (string, string, httpmock.Responder) {
	u := fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		bucket,
		object,
	)
	return http.MethodGet,
		u,
		httpmock.NewStringResponder(http.StatusOK, body)
}
