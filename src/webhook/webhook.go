// Package webhook provides signature verification and payload parsing for MUXI async webhooks.
//
// Usage:
//
//	import "github.com/muxi-ai/muxi-go/webhook"
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    payload, _ := io.ReadAll(r.Body)
//	    sig := r.Header.Get("X-Muxi-Signature")
//
//	    if err := webhook.VerifySignature(payload, sig, secret); err != nil {
//	        http.Error(w, "Invalid signature", 401)
//	        return
//	    }
//
//	    event, _ := webhook.Parse(payload)
//	    fmt.Println(event.Status, event.Content)
//	}
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultTolerance is the default time tolerance for signature verification (5 minutes).
const DefaultTolerance = 5 * time.Minute

// ErrInvalidSignature is returned when the webhook signature is invalid.
var ErrInvalidSignature = errors.New("invalid webhook signature")

// ErrInvalidTimestamp is returned when the webhook timestamp is outside tolerance.
var ErrInvalidTimestamp = errors.New("webhook timestamp outside tolerance")

// ErrMissingSignature is returned when the signature header is missing.
var ErrMissingSignature = errors.New("missing webhook signature")

// ErrInvalidPayload is returned when the payload cannot be parsed.
var ErrInvalidPayload = errors.New("invalid webhook payload")

// ContentItem represents a content item in the webhook response.
type ContentItem struct {
	Type string                 `json:"type"`
	Text string                 `json:"text,omitempty"`
	File map[string]interface{} `json:"file,omitempty"`
}

// ErrorDetails contains error information when status is "failed".
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Trace   string `json:"trace,omitempty"`
}

// Clarification contains clarification details when status is "awaiting_clarification".
type Clarification struct {
	Question               string `json:"clarification_question"`
	ClarificationRequestID string `json:"clarification_request_id,omitempty"`
	OriginalMessage        string `json:"original_message,omitempty"`
}

// WebhookEvent represents a parsed webhook event from MUXI async completion.
type WebhookEvent struct {
	RequestID      string                 `json:"id"`
	Status         string                 `json:"status"` // "completed" | "failed" | "awaiting_clarification"
	Timestamp      int64                  `json:"timestamp"`
	Content        []ContentItem          `json:"response"`
	Error          *ErrorDetails          `json:"error,omitempty"`
	Clarification  *Clarification         `json:"-"`
	FormationID    string                 `json:"formation_id,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	ProcessingTime float64                `json:"processing_time,omitempty"`
	ProcessingMode string                 `json:"processing_mode,omitempty"`
	WebhookURL     string                 `json:"webhook_url,omitempty"`
	Raw            map[string]interface{} `json:"-"`
}

// VerifySignature verifies the webhook signature and checks timestamp tolerance.
//
// Parameters:
//   - payload: Raw request body
//   - signatureHeader: Value of X-Muxi-Signature header (format: "t=timestamp,v1=signature")
//   - secret: Webhook secret (typically admin_key or dedicated webhook secret)
//
// Returns an error if verification fails.
//
// Example:
//
//	if err := webhook.VerifySignature(payload, sig, secret); err != nil {
//	    http.Error(w, "Invalid signature", 401)
//	    return
//	}
func VerifySignature(payload []byte, signatureHeader, secret string) error {
	return VerifySignatureWithTolerance(payload, signatureHeader, secret, DefaultTolerance)
}

// VerifySignatureWithTolerance verifies the webhook signature with custom time tolerance.
func VerifySignatureWithTolerance(payload []byte, signatureHeader, secret string, tolerance time.Duration) error {
	if signatureHeader == "" {
		return ErrMissingSignature
	}
	if secret == "" {
		return errors.New("webhook secret is required")
	}

	// Parse signature header: "t=1234567890,v1=abc123..."
	parts := make(map[string]string)
	for _, part := range strings.Split(signatureHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}

	timestampStr, hasTimestamp := parts["t"]
	signature, hasSignature := parts["v1"]
	if !hasTimestamp || !hasSignature {
		return ErrInvalidSignature
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}

	// Check timestamp tolerance (prevent replay attacks)
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > int64(tolerance.Seconds()) {
		return ErrInvalidTimestamp
	}

	// Compute expected signature: HMAC-SHA256(secret, "timestamp.payload")
	message := fmt.Sprintf("%d.", timestamp) + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return ErrInvalidSignature
	}

	return nil
}

// Parse parses webhook payload into a typed WebhookEvent.
//
// Parameters:
//   - payload: Raw request body (JSON bytes)
//
// Returns the parsed WebhookEvent or an error.
//
// Example:
//
//	event, err := webhook.Parse(payload)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if event.Status == "completed" {
//	    for _, item := range event.Content {
//	        if item.Type == "text" {
//	            fmt.Println(item.Text)
//	        }
//	    }
//	}
func Parse(payload []byte) (*WebhookEvent, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	event.Raw = raw

	// Handle clarification status
	if event.Status == "awaiting_clarification" {
		event.Clarification = &Clarification{}
		if q, ok := raw["clarification_question"].(string); ok {
			event.Clarification.Question = q
		}
		if id, ok := raw["clarification_request_id"].(string); ok {
			event.Clarification.ClarificationRequestID = id
		}
		if msg, ok := raw["original_message"].(string); ok {
			event.Clarification.OriginalMessage = msg
		}
	}

	return &event, nil
}
