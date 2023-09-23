package httpmockext

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

func AssertCalled(
	t *testing.T,
	transport *httpmock.MockTransport,
) {
	t.Helper()

	callCounts := transport.GetCallCountInfo()
	for _, responder := range transport.Responders() {
		count := callCounts[responder]
		if count == 0 {
			t.Errorf("should be called: %s", responder)
		}
	}
}
