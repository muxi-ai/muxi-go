package muxi

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/google/uuid"
)

// sdkTransport wraps http.RoundTripper to add SDK headers
type sdkTransport struct {
	base   http.RoundTripper
	logger *log.Logger
	debug  bool
}

func (t *sdkTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Add SDK identification header
	req.Header.Set("X-Muxi-SDK", "go/"+Version)

	// Add client info header
	req.Header.Set("X-Muxi-Client", runtime.GOOS+"-"+runtime.GOARCH+"/go"+runtime.Version()[2:])

	// Add idempotency key for every request
	if req.Header.Get("X-Muxi-Idempotency-Key") == "" {
		req.Header.Set("X-Muxi-Idempotency-Key", uuid.New().String())
	}

	resp, err := t.base.RoundTrip(req)

	if t.debug && t.logger != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.logger.Printf("[muxi] %s %s -> %d (%v)", req.Method, req.URL.String(), status, time.Since(start))
	}

	return resp, err
}

// newSDKTransport creates a new transport with SDK headers
func newSDKTransport(base http.RoundTripper, logger *log.Logger, debug bool) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &sdkTransport{base: base, logger: logger, debug: debug}
}
