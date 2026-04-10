package muxi

import (
	"log"
	"net/http"
	"testing"
)

// mockRoundTripper captures the request
type mockRoundTripper struct {
	req *http.Request
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	m.req = r
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestSDKTransportHeaders(t *testing.T) {
	Version = "1.2.3"
	mock := &mockRoundTripper{}
	tr := newSDKTransport(mock, log.Default(), false, "")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if _, err := tr.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}

	if got := mock.req.Header.Get("X-Muxi-SDK"); got != "go/1.2.3" {
		t.Fatalf("X-Muxi-SDK = %s", got)
	}

	if got := mock.req.Header.Get("X-Muxi-Client"); got == "" {
		t.Fatalf("X-Muxi-Client missing")
	}

	if got := mock.req.Header.Get("X-Muxi-Idempotency-Key"); got == "" {
		t.Fatalf("X-Muxi-Idempotency-Key missing")
	}
}

func TestSDKTransportAppHeader(t *testing.T) {
	Version = "1.2.3"
	mock := &mockRoundTripper{}
	tr := newSDKTransport(mock, log.Default(), false, "console")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if _, err := tr.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}

	if got := mock.req.Header.Get("X-Muxi-App"); got != "console" {
		t.Fatalf("X-Muxi-App = %q, want %q", got, "console")
	}
}
