package gmailext

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/api/gmail/v1"
)

func bodyData(message *gmail.Message) string {
	if message.Payload.Body.Data != "" {
		return message.Payload.Body.Data
	}
	for _, part := range message.Payload.Parts {
		return part.Body.Data
	}
	return ""
}

func StringBody(message *gmail.Message) (string, error) {
	rawBody := bodyData(message)
	body, err := base64.URLEncoding.DecodeString(rawBody)
	if err != nil {
		return "", fmt.Errorf("base64 url decode: %w", err)
	}
	return string(body), nil
}
