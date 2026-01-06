package muxi

import (
	"net/http"
	"runtime"

	"github.com/google/uuid"
)

// sdkTransport wraps http.RoundTripper to add SDK headers
type sdkTransport struct {
	base http.RoundTripper
}

func (t *sdkTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add SDK identification header
	req.Header.Set("X-Muxi-SDK", "go/"+Version)

	// Add client info header
	req.Header.Set("X-Muxi-Client", runtime.GOOS+"-"+runtime.GOARCH+"/go"+runtime.Version()[2:])

	// Add idempotency key for every request
	if req.Header.Get("X-Muxi-Idempotency-Key") == "" {
		req.Header.Set("X-Muxi-Idempotency-Key", uuid.New().String())
	}

	return t.base.RoundTrip(req)
}

// newSDKTransport creates a new transport with SDK headers
func newSDKTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &sdkTransport{base: base}
}
