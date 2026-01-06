package muxi

import "net/http"

// Basic tests for retry helper functions

func isRetry(method string, status int, max int, attempt int) bool {
    return shouldRetry(method, status, max, attempt)
}

func retryCodes() []int {
    return []int{http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout}
}

// note: backoffDelay uses rand jitter; we only assert range in tests elsewhere if needed
