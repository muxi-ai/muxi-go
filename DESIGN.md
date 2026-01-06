# MUXI Go SDK Design

**Status:** Design Document  
**Last Updated:** 2025-01-05  
**Version:** 1.0

---

## Overview

Go SDK for MUXI providing programmatic access to:
- **Server Management** (Server API) - Deploy, list, manage formations
- **Formation Usage** (Formation API) - Chat, health, agents, secrets, etc.

---

## Architecture

### Two Separate Clients

```go
package muxi

// ServerClient - Server Management (Server API)
type ServerClient struct {
    BaseURL    string  // https://muxi.company.com:7890
    KeyID      string  // MUXI_SERVER_KEY_ID
    SecretKey  string  // MUXI_SERVER_SECRET_KEY
    HTTPClient *http.Client
}

// FormationClient - Formation Usage (Formation API)
type FormationClient struct {
    FormationID string
    BaseURL     string  // Direct: http://localhost:8001
                      // Or Proxy: https://muxi.company.com:7890/api/my-formation
    AdminKey  string  // X-MUXI-ADMIN-KEY
    ClientKey string  // X-MUXI-CLIENT-KEY
    HTTPClient *http.Client
}
```

### Connection Modes

**Direct Mode** (development/testing):
```go
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    URL:         "http://localhost:8001",  // Direct to formation port
    ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
})
```

**Proxy Mode** (production, via server):
```go
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    ServerURL:   "https://muxi.company.com:7890",  // Via server proxy
    ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
})
// Uses: https://muxi.company.com:7890/api/my-formation/v1/*
```

---

## API Surface

### ServerClient Methods

```go
// Formation lifecycle
DeployFormation(ctx, req *DeployRequest) (*DeployResult, error)
DeployFormationStreaming(ctx, req *DeployRequest) (<-chan DeployEvent, error)
ListFormations(ctx) (*FormationList, error)
GetFormation(ctx, id string) (*Formation, error)
UpdateFormation(ctx, id, bundlePath string) (*UpdateResult, error)
StopFormation(ctx, id string) error
StartFormation(ctx, id string) error
RestartFormation(ctx, id string) error
RestartFormationStreaming(ctx, id string) (<-chan RestartEvent, error)
RollbackFormation(ctx, id string) (*RollbackResult, error)
RollbackFormationStreaming(ctx, id string) (<-chan RollbackEvent, error)
DeleteFormation(ctx, id string) error
CancelUpdate(ctx, id string) error

// Logs
GetFormationLogs(ctx, id string, opts *LogOptions) (*Logs, error)
StreamFormationLogs(ctx, id string, opts *LogOptions) (<-chan LogLine, error)
GetServerLogs(ctx, lines int) ([]string, error)

// Server management
Status(ctx) (*ServerStatus, error)
Health(ctx) (*Health, error)
Ping(ctx) error
```

### FormationClient Methods

```go
// Core
Health(ctx) (*HealthStatus, error)
Chat(ctx, req *ChatRequest) (*ChatResponse, error)
ChatStream(ctx, req *ChatRequest) (<-chan ChatChunk, error)
AudioChat(ctx, req *AudioChatRequest) (*ChatResponse, error)
AudioChatStream(ctx, req *AudioChatRequest) (<-chan ChatChunk, error)

// Config
GetStatus(ctx) (*StatusResponse, error)
GetConfig(ctx) (*ConfigResponse, error)
GetFormationInfo(ctx) (*FormationInfo, error)

// Agents (read-only - modifications via deployment)
GetAgents(ctx) (*AgentList, error)
GetAgent(ctx, id string) (*Agent, error)

// Secrets
GetSecrets(ctx) (*SecretList, error)
GetSecret(ctx, key string) (*Secret, error)
SetSecret(ctx, key, value string) error
DeleteSecret(ctx, key string) error

// MCP
GetMCPServers(ctx) (*MCPList, error)
GetMCPServer(ctx, id string) (*MCP, error)
GetMCPTools(ctx) (*MCPTools, error)

// Memory
GetMemoryConfig(ctx) (*MemoryConfig, error)
GetMemories(ctx, userID string) (*MemoryList, error)
AddMemory(ctx, userID, memType, detail string) (*Memory, error)
DeleteMemory(ctx, userID, memoryID string) error
GetUserBuffer(ctx, userID string) (*UserBuffer, error)
ClearUserBuffer(ctx, userID string) error
ClearSessionBuffer(ctx, sessionID, userID string) error

// Scheduler
GetSchedulerConfig(ctx) (*SchedulerConfig, error)
GetSchedulerJobs(ctx, userID string) (*JobList, error)
GetSchedulerJob(ctx, jobID string) (*Job, error)
CreateSchedulerJob(ctx, jobType, schedule, message, userID string) (*Job, error)
DeleteSchedulerJob(ctx, jobID string) error

// Sessions
GetSessions(ctx, userID string, limit int) (*SessionList, error)
GetSession(ctx, sessionID, userID string) (*Session, error)
GetSessionMessages(ctx, sessionID, userID string) (*Messages, error)
RestoreSession(ctx, sessionID, userID string, messages []Message) error

// Requests
GetRequests(ctx, userID string) (*RequestList, error)
GetRequestStatus(ctx, requestID, userID string) (*RequestStatus, error)
CancelRequest(ctx, requestID, userID string) error
StreamRequest(ctx, sessionID, requestID, userID string) (<-chan SSEEvent, error)

// Users/Identifiers
GetUserIdentifiers(ctx) (*IdentifierList, error)
GetUserIdentifiersForUser(ctx, userID string) (*IdentifierList, error)
ResolveUser(ctx, identifier string, createUser bool) (*User, error)
LinkUserIdentifier(ctx, identifier, userID, idType string) error
UnlinkUserIdentifier(ctx, identifier string) error

// Credentials
ListCredentials(ctx, userID string) (*CredentialList, error)
GetCredential(ctx, credentialID, userID string) (*Credential, error)
CreateCredential(ctx, userID string, req *CreateCredentialRequest) (*Credential, error)
DeleteCredential(ctx, credentialID, userID string) error
ListCredentialServices(ctx) (*CredentialServices, error)

// Triggers
GetTriggers(ctx) (*TriggerList, error)
GetTrigger(ctx, name string) (*Trigger, error)
TriggerTrigger(ctx, name string, data json.RawMessage, async bool, userID string) (*TriggerResult, error)

// SOPs
GetSOPs(ctx) (*SOPList, error)
GetSOP(ctx, name string) (*SOP, error)

// Logging/Events
StreamEvents(ctx, userID string) (<-chan Event, error)
StreamLogs(ctx, opts *LogFilters) (<-chan LogEntry, error)

// Memory buffers and stats
GetMemoryBuffers(ctx) (*MemoryBufferList, error)
GetBufferStats(ctx) (*BufferStats, error)
ClearAllBuffers(ctx) error

// Async / A2A / Logging config
GetAsyncConfig(ctx) (*AsyncSettings, error)
GetAsyncJobs(ctx) (*AsyncJobList, error)
GetAsyncJob(ctx, jobID string) (*AsyncJob, error)
CancelAsyncJob(ctx, jobID string) error
GetA2AConfig(ctx) (*A2AConfig, error)
GetLoggingConfig(ctx) (*LoggingConfig, error)
GetLoggingDestinations(ctx) (*LoggingDestinations, error)

// Admin-only
GetOverlordConfig(ctx) (*OverlordConfig, error)
GetOverlordPersona(ctx) (*Persona, error)
GetLLMSettings(ctx) (*LLMSettings, error)
GetAuditLog(ctx) (*AuditLog, error)
ClearAuditLog(ctx) error
```

---

## Streaming Pattern

All streaming operations use Go channels for idiomatic consumption:

```go
// Deploy with progress
events, err := server.DeployFormationStreaming(ctx, &muxi.DeployRequest{
    BundlePath: "my-bot.tar.gz",
    Metadata: map[string]string{"id": "my-bot"},
})
if err != nil {
    panic(err)
}

for event := range events {
    switch event.Type {
    case "progress":
        fmt.Printf("[%s] %s\n", event.Stage, event.Message)
    case "complete":
        fmt.Printf("Deployed: %s on port %d\n", event.FormationID, event.Port)
    case "error":
        panic(event.Error)
    }
}

// Chat streaming
chunks, err := client.ChatStream(ctx, &muxi.ChatRequest{
    Message: "Tell me a story",
    UserID:  "user-123",
})
if err != nil {
    panic(err)
}

for chunk := range chunks {
    if chunk.Type == "text" {
        fmt.Print(chunk.Text)
    }
}
```

---

## Package Structure

```
muxi-go/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── client.go              // Common client logic, constructors
├── server_client.go       // ServerClient implementation
├── formation_client.go    // FormationClient implementation
├── streaming.go           // SSE stream parsing helpers
├── auth.go                // HMAC signing for ServerClient
├── models.go              // Request/response types
├── errors.go              // Custom error types
├── options.go             // Functional options pattern
│
├── examples/
│   ├── deploy/
│   │   └── main.go       // Deploy formation
│   ├── streaming/
│   │   └── main.go       // Chat streaming
│   ├── monitoring/
│   │   └── main.go       // Health checks
│   └── testing/
│       └── main.go       // Integration test fixture
│
├── server_client_test.go
├── formation_client_test.go
├── streaming_test.go
└── auth_test.go
```

---

## Error Handling

Structured error types (not user-friendly messages):

```go
type MuxiError struct {
    Code       string
    Message    string
    StatusCode int
    Details    map[string]interface{}
}

// Specific error types
type AuthenticationError struct { *MuxiError }
type AuthorizationError struct { *MuxiError }
type NotFoundError struct { *MuxiError }
type ConflictError struct { *MuxiError }
type ValidationError struct { *MuxiError }
type RateLimitError struct {
    *MuxiError
    RetryAfter int
}
```

---

## Design Decisions

### 1. Two Clients vs One

**Decision:** Two separate clients (`ServerClient` and `FormationClient`)

**Rationale:**
- Clear separation: management vs usage
- Different auth mechanisms (HMAC vs API keys)
- Different URL patterns (server vs formation)
- Prevents accidental API misuse

### 2. Direct vs Proxy Mode

**Decision:** Support both with auto-detection based on config

**Rationale:**
- Development: connect directly to formation port (faster, no proxy overhead)
- Production: via server proxy (auth, routing, centralized logging)
- `FormationConfig.ServerURL` set → proxy mode
- `FormationConfig.URL` set → direct mode

### 3. Channel-Based Streaming

**Decision:** Return `<-chan Event` for all streaming operations

**Rationale:**
- Idiomatic Go (channels are standard for streaming)
- Easy to use with `for range`
- Supports context cancellation
- Caller can buffer, timeout, or process concurrently

### 4. Context-First Methods

**Decision:** All methods take `ctx context.Context` as first parameter

**Rationale:**
- Standard Go practice since Go 1.7
- Enables per-request timeouts
- Supports cancellation
- Tracing/observability integration

### 5. No Profile Management

**Decision:** SDK does NOT use ~/.muxi/profiles.yaml

**Rationale:**
- CLI-specific feature
- SDK users pass credentials explicitly
- Environment variables or application config
- No auto-detection magic

---

## Design Decisions (Resolved)

### 1. Two Clients vs One ✅

**Decision:** Two separate clients (`ServerClient` and `FormationClient`)

**Rationale:**
- Clear separation: management vs usage
- Different auth mechanisms (HMAC vs API keys)
- Different URL patterns (server vs formation)
- Prevents accidental API misuse

---

### 2. Direct vs Proxy Mode ✅

**Decision:** Auto-detect based on which field is set in `FormationConfig`

**Options:**

```go
// Option A: Direct URL (development/testing)
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    URL:         "http://localhost:8001",  // Direct to formation
    ClientKey:   "fmc-...",
})
// BaseURL = "http://localhost:8001/v1"

// Option B: Server proxy (production)
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    ServerURL:   "https://muxi.company.com:7890",
    ClientKey:   "fmc-...",
})
// BaseURL = "https://muxi.company.com:7890/api/my-bot/v1"

// Option C: Explicit BaseURL (advanced)
client := muxi.NewFormationClient(&muxi.FormationConfig{
    BaseURL:  "http://custom-proxy:8080/my-formation/v1",
    ClientKey: "fmc-...",
})
```

**Implementation:**
```go
func NewFormationClient(cfg *FormationConfig) *FormationClient {
    var baseURL string
    
    if cfg.BaseURL != "" {
        baseURL = cfg.BaseURL
    } else if cfg.URL != "" {
        baseURL = cfg.URL + "/v1"
    } else if cfg.ServerURL != "" {
        baseURL = cfg.ServerURL + "/api/" + cfg.FormationID + "/v1"
    } else {
        panic("must provide URL, ServerURL, or BaseURL")
    }
    
    return &FormationClient{
        FormationID: cfg.FormationID,
        BaseURL:     baseURL,
        // ...
    }
}
```

**Priority:** `BaseURL` > `URL` > `ServerURL`

---

### 3. Pointer Types for Optional Fields ✅

**Decision:** Use pointer types (`*string`, `*int`) - standard Go practice

```go
type ChatRequest struct {
    Message   string  `json:"message"`
    UserID    string  `json:"user_id"`
    SessionID *string `json:"session_id,omitempty"`  // Optional
    GroupID   *string `json:"group_id,omitempty"`    // Optional
}
```

**Why not wrapper types?**
- Pointers are idiomatic in Go
- `omitempty` tag works naturally
- JSON marshaling handles nil correctly
- No need for custom MarshalJSON methods

---

### 4. Request/Response Structs ✅

**Decision:** Flat types in package namespace

```go
// ✅ Good - clean, importable
type ChatRequest struct { ... }
type ChatResponse struct { ... }

// ❌ Bad - verbose, harder to import
type ServerClient struct { ... }
type ServerClientChatRequest struct { ... }
```

**Organization:**
- Server API types: `DeployRequest`, `DeployResponse`, `FormationList`, etc.
- Formation API types: `ChatRequest`, `ChatResponse`, `Agent`, etc.
- Shared: `APIResponse`, `MuxiError`, etc.

---

### 5. Retry Logic ✅

**Decision:** Simple `MaxRetries` config (KISS for v1)

```go
// Usage
client := muxi.NewServerClient(&muxi.ServerConfig{
    ServerURL:  "https://muxi.company.com:7890",
    KeyID:      "MUXI_PROD_KEY",
    SecretKey:  "sk-...",
    MaxRetries: 3,  // 0 = disabled (default)
})
```

**Hardcoded behavior (v1):**
- 500ms initial delay, doubles each retry, max 30s
- Jitter ±10%
- Retryable: 408, 429, 500, 502, 503, 504
- Respects `Retry-After` header for 429s
- GET/DELETE retried, POST never (no idempotency guarantee)

Future: Expose full `RetryConfig` if users need fine-tuning.

---

### 6. Header Injection (X-Muxi-SDK) ✅

**Decision:** Yes, auto-inject via RoundTripper

```go
type sdkTransport struct {
    base    http.RoundTripper
    version string
}

func (t *sdkTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req.Header.Set("X-Muxi-SDK", "go/"+t.version)
    return t.base.RoundTrip(req)
}
```

**Version from:** `var Version = "dev"` (set via ldflags at build time)

---

## How to Fork CLI into SDK

### Overview

The Go SDK will be the **reference implementation** for all other SDKs. The CLI already has working implementations of both Server API and Formation API clients. We'll extract, clean up, and adapt them into the SDK.

### Step 1: Extract Server Client

**Source:** `cli/src/pkg/server/client.go` (672 lines)

**What to keep:**
- HMAC auth logic (`BuildAuthHeader`)
- Request/response types
- SSE stream parsing (`parseSSEStream`)
- Formation lifecycle methods
- Health/status endpoints

**What to remove:**
- Profile management (SDK uses explicit creds)
- Local server auto-detection
- CLI-specific error messages

**What to add:**
- Retry transport wrapper
- X-Muxi-SDK header injection
- Context propagation
- Configurable timeouts

**Mapping:**
```
cli/src/pkg/server/client.go                    → muxi-go/server_client.go
cli/src/pkg/server/auth.go                      → muxi-go/auth.go
cli/src/pkg/server/types.go                     → muxi-go/models.go (partial)
```

### Step 2: Extract Formation Client

**Source:** `cli/src/pkg/formation/client.go` (900+ lines)

**What to keep:**
- All HTTP methods (`Get`, `Post`, `DoWithUserID`, etc.)
- Request/response types
- SSE streaming (`ChatStream`, `StreamEvents`, `StreamLogs`)
- All API endpoints (agents, secrets, sessions, etc.)

**What to remove:**
- CLI-specific helpers
- Profile lookups

**What to add:**
- Direct vs Proxy mode auto-detection
- Context support
- Retry transport
- X-Muxi-SDK header

**Mapping:**
```
cli/src/pkg/formation/client.go               → muxi-go/formation_client.go
cli/src/pkg/formation/types.go                → muxi-go/models.go (partial)
cli/src/pkg/formation/auth.go                  → muxi-go/auth.go (partial)
```

### Step 3: Unified Models

**Source:** `cli/src/pkg/server/types.go` + `cli/src/pkg/formation/types.go`

**Merge into:** `muxi-go/models.go`

**Strategy:**
1. Server API types (formation lifecycle, server status)
2. Formation API types (chat, agents, sessions, etc.)
3. Shared types (APIResponse envelope, errors)

**Example:**
```go
// Server API
type DeployRequest struct { ... }
type FormationList struct { ... }
type ServerStatus struct { ... }

// Formation API
type ChatRequest struct { ... }
type Agent struct { ... }
type Session struct { ... }

// Shared
type APIResponse struct {
    Object    string          `json:"object"`
    Timestamp int64           `json:"timestamp"`
    Type      string          `json:"type"`
    Request   RequestMeta     `json:"request"`
    Success   bool            `json:"success"`
    Error     *ErrorInfo      `json:"error,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
}
```

### Step 4: Streaming Helpers

**Source:** SSE parsing in both clients

**Extract to:** `muxi-go/streaming.go`

**Common patterns:**
```go
// Parse SSE stream into channel
func parseSSEStream(reader io.Reader) (<-chan SSEEvent, <-chan error)

// Read SSE lines
func readSSELines(reader io.Reader) <-chan string

// Parse SSE event line
func parseSSEEvent(line string) (eventType, data string)
```

### Step 5: Remove CLI Dependencies

**Delete these patterns:**
```go
// ❌ Profile management
profile, _ := cmd.Flags().GetString("profile")
client, _ := server.NewClient(profile)

// ❌ Context detection
formationID := context.DetectFormationID()

// ❌ CLI output
ui.Success("Deployed!")
ui.ErrorBlock("Formation not found")
```

**Replace with:**
```go
// ✅ Explicit creds
client := muxi.NewServerClient(&muxi.ServerConfig{
    URL:       os.Getenv("MUXI_SERVER_URL"),
    KeyID:     os.Getenv("MUXI_KEY_ID"),
    SecretKey: os.Getenv("MUXI_SECRET_KEY"),
})

// ✅ Return structured data
return &DeployResult{FormationID: "my-bot", Port: 8000}

// ✅ Return errors
return fmt.Errorf("formation not found: %s", id)
```

### Step 6: Add SDK-Specific Features

**Not in CLI, add to SDK:**

1. **Functional Options Pattern**
```go
client := muxi.NewServerClient(
    muxi.WithURL("https://muxi.company.com:7890"),
    muxi.WithCredentials("MUXI_KEY", "sk-..."),
    muxi.WithRetry(&muxi.RetryConfig{MaxRetries: 3}),
    muxi.WithTimeout(60*time.Second),
)
```

2. **Context Helpers**
```go
// Per-request timeout
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
client.Chat(ctx, req)

// Per-request retry override
ctx := muxi.WithRetry(ctx, &muxi.RetryConfig{MaxRetries: 5})
client.Chat(ctx, req)
```

3. **Middleware/Hooks (future)**
```go
client := muxi.NewServerClient(
    muxi.WithRequestHook(func(req *http.Request) {
        log.Printf("[MUXI] %s %s", req.Method, req.URL)
    }),
    muxi.WithResponseHook(func(resp *http.Response) {
        log.Printf("[MUXI] %d", resp.StatusCode)
    }),
)
```

### Step 7: Test Structure

**Adopt CLI tests, but sandbox:**

```go
// From: cli/src/pkg/formation/client_test.go
func TestGetAgents(t *testing.T) {
    // ❌ Remove: Profile lookups
    // ✅ Keep: HTTP mocking
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/agents" {
            t.Errorf("expected /agents, got %s", r.URL.Path)
        }
        // Return mock response
    }))
    defer server.Close()

    client := muxi.NewFormationClient(&muxi.FormationConfig{
        BaseURL:  server.URL,
        ClientKey: "test-key",
    })

    agents, err := client.GetAgents(context.Background())
    // Assert...
}
```

### File Mapping Summary

| CLI File | SDK File | Changes |
|----------|----------|---------|
| `pkg/server/client.go` | `server_client.go` | Remove profiles, add retry, add context |
| `pkg/server/auth.go` | `auth.go` | Keep as-is (HMAC logic) |
| `pkg/server/types.go` | `models.go` (partial) | Merge with formation types |
| `pkg/formation/client.go` | `formation_client.go` | Remove profiles, add direct/proxy modes |
| `pkg/formation/types.go` | `models.go` (partial) | Merge with server types |
| `pkg/formation/auth.go` | (merge into auth.go) | API key logic |
| `cmd/formation.go` | `examples/` | Usage examples |

### New Files (not in CLI)

```
muxi-go/
├── client.go         # Constructors, functional options
├── errors.go         # Error type definitions
├── options.go        # Functional option helpers
├── retry.go          # Retry logic
└── middleware.go     # Request/response hooks (future)
```

### Dependencies to Remove

```go
// ❌ CLI-only
import (
    "github.com/muxi-ai/cli/pkg/defaults"  // No profile management
    "github.com/muxi-ai/cli/pkg/ui"         // No pretty output
    "github.com/spf13/cobra"                // No CLI framework
)

// ✅ SDK only
import (
    "context"      // Add
    "time"         // Add
)
```

### Build Tags

```go
//go:build !integration
// +build !integration

// Unit tests (no server required)

//go:build integration
// +build integration

// Integration tests (requires server)
```

---

## Implementation Checklist

### Foundation
- [ ] `go.mod` with minimal dependencies
- [ ] `client.go` - constructors, functional options
- [ ] `errors.go` - error type hierarchy
- [ ] `models.go` - extract from CLI types
- [ ] `auth.go` - extract HMAC + API key logic

### ServerClient
- [ ] `server_client.go` - extract from `pkg/server/client.go`
- [ ] Remove profile management
- [ ] Add context to all methods
- [ ] Add retry transport
- [ ] Add X-Muxi-SDK header

### FormationClient
- [ ] `formation_client.go` - extract from `pkg/formation/client.go`
- [ ] Add direct/proxy mode detection
- [ ] Add context to all methods
- [ ] Add retry transport
- [ ] Add X-Muxi-SDK header

### Streaming
- [ ] `streaming.go` - SSE parsing helpers
- [ ] Channel-based event delivery
- [ ] Context cancellation support

### Testing
- [ ] Unit tests (HTTP mocking)
- [ ] Integration tests (with test server)
- [ ] Examples (`examples/` directory)

### Documentation
- [ ] README with quick start
- [ ] Godoc comments on all exports
- [ ] Examples for common use cases

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Package structure, go.mod
- [ ] Base client structs, constructors
- [ ] HMAC auth (ServerClient)
- [ ] Error types
- [ ] Context handling

### Phase 2: ServerClient (Week 2)
- [ ] Formation lifecycle (deploy, list, get, delete, stop, start)
- [ ] Streaming deploy (SSE parsing)
- [ ] Health/status endpoints

### Phase 3: FormationClient - Core (Week 3)
- [ ] Health, status, config
- [ ] Chat (non-streaming)
- [ ] Chat streaming (SSE parsing)
- [ ] Agents, secrets, MCP

### Phase 4: FormationClient - Advanced (Week 4)
- [ ] Sessions, requests
- [ ] Memory, scheduler
- [ ] Users, credentials
- [ ] Triggers, SOPs

### Phase 5: Polish (Week 5)
- [ ] Examples
- [ ] Tests (unit + integration)
- [ ] Documentation
- [ ] CI/CD

---

## Resolved Issues

1. **Server audit logs** ✅ - Yes, add `GetServerLogs()` to ServerClient
2. **Idempotency keys** ✅ - Auto-generate UUID, send as `X-Muxi-Idempotency-Key` on every request
3. **File uploads** ✅ - Base64 encoded in JSON body (not multipart)
4. **Webhook support** ✅ - Yes, `webhook_url` field in ChatRequest for async
5. **OpenTelemetry** ✅ - Skip, just `X-Muxi-SDK` + `X-Muxi-Idempotency-Key` headers
6. **Connection pooling** ✅ - Skip in v1
7. **Pagination** ✅ - Simple `limit` param, `has_more` in response
8. **Timeouts** ✅ - 30s default (configurable), streaming infinite
9. **Response metadata** ✅ - Expose `request_id` and `timestamp` on all responses
10. **Retry config** ✅ - KISS: just `MaxRetries int` for v1

---

## References

- [CLI Implementation](../../cli/src) - Reference for auth, streaming patterns
- [SDK Design](../SDK-DESIGN.md) - Overall SDK philosophy
- [Server API Spec](../../schemas/api/server-api-v1-final.yaml)
- [Formation API Spec](../../schemas/api/formation-api-v1-final.yaml)

## Implementation Timeline

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| 1 | Foundation | Package setup, clients, auth, errors, retry |
| 2 | ServerClient | Deploy, list, update, delete, streaming |
| 3 | FormationClient Core | Chat (streaming), health, agents, secrets |
| 4 | FormationClient Advanced | Sessions, memory, scheduler, users, etc. |
| 5 | Polish | Tests, examples, docs, CI/CD |

---

## References

- [CLI Implementation](../../cli/src) - Source code to fork from
- [SDK Design](../SDK-DESIGN.md) - Overall SDK philosophy  
- [Server API Spec](../../schemas/api/server-api-v1-final.yaml)
- [Formation API Spec](../../schemas/api/formation-api-v1-final.yaml)
- [Go API Design](https://go.dev/doc/effective_go#documentation) - Go conventions
