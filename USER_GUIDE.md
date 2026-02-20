## MUXI Go SDK — User Guide

This guide is the concise, Go-specific manual for developers and AI agents using `github.com/muxi-ai/muxi-go`. It complements (not replaces) muxi.org docs.

### Installation
```bash
go get github.com/muxi-ai/muxi-go
```

### Clients
- **ServerClient** (management, HMAC): deploy/list/update formations, server health/status, server logs.
- **FormationClient** (runtime, client/admin keys): chat/audio (streaming), agents, secrets, MCP, memory, scheduler, sessions/requests, identifiers, credentials, triggers/SOPs/audit, async/A2A/logging config, overlord/LLM settings, events/logs/request streaming.

### Minimal examples
**ServerClient**
```go
server := muxi.NewServerClient(&muxi.ServerConfig{
    URL:       os.Getenv("MUXI_SERVER_URL"),
    KeyID:     os.Getenv("MUXI_KEY_ID"),
    SecretKey: os.Getenv("MUXI_SECRET_KEY"),
})

res, err := server.DeployFormation(ctx, &muxi.DeployRequest{BundlePath: "my-bot.tar.gz"})
if err != nil { log.Fatal(err) }
fmt.Println("deployed", res.FormationID)

logs, _ := server.GetServerLogs(ctx, 200)
fmt.Println(len(logs), "server log lines")
```

**FormationClient (direct or proxy)**
```go
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    ServerURL:   os.Getenv("MUXI_SERVER_URL"), // or URL: http://localhost:8001
    ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
})

resp, err := client.Chat(ctx, &muxi.ChatRequest{Message: "Hello", UserID: "u1"})
if err != nil { log.Fatal(err) }
fmt.Println(resp.Response)

stream, errs := client.ChatStream(ctx, &muxi.ChatRequest{Message: "Tell me a story", UserID: "u1"})
for stream != nil {
    select {
    case chunk, ok := <-stream:
        if !ok { stream = nil; continue }
        if chunk.Type == "text" { fmt.Print(chunk.Text) }
    case err := <-errs:
        if err != nil { log.Fatal(err) }
        errs = nil
    }
}
```

### Streaming patterns
- All streaming returns `<-chan ...` + `<-chan error`; context cancellation stops streams.
- Buffers: SSE scanner starts at 256KB, grows to 10MB.
- No timeouts on streams; rely on context.

### Idempotency & headers
- SDK auto-adds `X-Muxi-Idempotency-Key` (UUID) on **every** request; no opt-out.
- Also adds `X-Muxi-SDK: go/{version}`, `X-Muxi-Client: {os-arch}/go{ver}`.

### Auth
- ServerClient: HMAC (`MUXI-HMAC key=<id>, timestamp=<sec>, signature=<b64>`), query stripped when signing.
- FormationClient: `X-MUXI-CLIENT-KEY` required; `X-MUXI-ADMIN-KEY` for admin endpoints.

### Retries & timeouts
- Default timeout: 30s (non-streaming). Streams are infinite.
- Retries: GET/DELETE only, backoff 500ms *2, max 30s, jitter ±10%, respects Retry-After for 429.

### Response metadata
- Most responses expose `RequestID` and `Timestamp`; useful for support/log correlation.

### Notable endpoints (FormationClient)
- Chat: `Chat`, `ChatStream`, `AudioChat`, `AudioChatStream` (chunk types: text, tool_call, tool_result, agent_handoff, thinking, error, done).
- Memory: `GetMemoryConfig`, `GetMemories`, `AddMemory`, `DeleteMemory`, `GetUserBuffer`, `ClearUserBuffer`, `ClearSessionBuffer`, `ClearAllBuffers`, `GetMemoryBuffers`, `GetBufferStats`.
- Scheduler: `GetSchedulerConfig`, `GetSchedulerJobs`, `GetSchedulerJob`, `CreateSchedulerJob`, `DeleteSchedulerJob`.
- Requests: `GetRequests`, `GetRequestStatus`, `CancelRequest`, `StreamRequest`.
- Logs/Events: `StreamLogs(filters)`, `StreamEvents(userID)`.
- Config/admin: `GetAsyncConfig`, `GetAsyncJobs`, `CancelAsyncJob`, `GetA2AConfig`, `GetLoggingConfig`, `GetLoggingDestinations`, `GetOverlordConfig`, `GetOverlordSoul`, `GetLLMSettings`.

### Webhook Verification

For async operations, MUXI delivers results via webhooks. The SDK provides helpers to verify signatures and parse payloads.

```go
import "github.com/muxi-ai/muxi-go/webhook"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("X-Muxi-Signature")

    // Verify signature
    if err := webhook.VerifySignature(payload, sig, secret); err != nil {
        http.Error(w, "Invalid signature", 401)
        return
    }

    // Parse into typed struct
    event, err := webhook.Parse(payload)
    if err != nil {
        http.Error(w, "Invalid payload", 400)
        return
    }

    switch event.Status {
    case "completed":
        for _, item := range event.Content {
            if item.Type == "text" {
                fmt.Println(item.Text)
            }
        }
    case "failed":
        fmt.Printf("Error: %s\n", event.Error.Message)
    case "awaiting_clarification":
        fmt.Printf("Question: %s\n", event.Clarification.Question)
    }

    w.WriteHeader(http.StatusOK)
}
```

**Webhook Functions:**
- `webhook.VerifySignature(payload, signature, secret)` - Verify HMAC-SHA256 signature
- `webhook.VerifySignatureWithTolerance(payload, signature, secret, tolerance)` - Custom time tolerance
- `webhook.Parse(payload)` - Parse into `*webhook.WebhookEvent`

**WebhookEvent Fields:** `RequestID`, `Status`, `Timestamp`, `Content`, `Error`, `Clarification`, `ProcessingTime`, `Raw`

### Troubleshooting
- Connection errors: ensure URL/baseURL and keys are set; for streaming, check proxies/firewalls.
- 401/403: verify client/admin keys (Formation) or key/secret (Server).
- 429: retries respect `Retry-After`; consider lowering call rate.
- Hanging streams: ensure context cancellation on shutdown.

### Testing locally
```bash
cd go/src
go test ./...
```

### Contributing notes (quick)
- Format with `gofmt` before commit.
- Do not remove idempotency header injection.
- Preserve streaming transport reuse and infinite timeouts.
