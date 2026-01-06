# MUXI Go SDK

Official Go SDK for [MUXI](https://muxi.ai) - the infrastructure layer for AI agents.

> Need deeper usage notes? See [USER_GUIDE.md](./USER_GUIDE.md) for streaming, retries, and auth details.

## Installation

```bash
go get github.com/muxi-ai/muxi-go
```

## Quick Start

### ServerClient - Formation Management

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/muxi-ai/muxi-go"
)

func main() {
    ctx := context.Background()

    // Create server client for formation management
    server := muxi.NewServerClient(&muxi.ServerConfig{
        URL:       os.Getenv("MUXI_SERVER_URL"),
        KeyID:     os.Getenv("MUXI_KEY_ID"),
        SecretKey: os.Getenv("MUXI_SECRET_KEY"),
    })

    // List formations
    formations, err := server.ListFormations(ctx)
    if err != nil {
        panic(err)
    }
    for _, f := range formations.Formations {
        fmt.Printf("%s: %s (v%s)\n", f.ID, f.Status, f.Version)
    }

    // Deploy a formation
    result, err := server.DeployFormation(ctx, &muxi.DeployRequest{
        BundlePath: "my-bot.tar.gz",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Deployed: %s on port %d\n", result.FormationID, result.Port)
}
```

### FormationClient - Chat & Runtime API

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/muxi-ai/muxi-go"
)

func main() {
    ctx := context.Background()

    // Direct mode (development) - connect directly to formation
    client := muxi.NewFormationClient(&muxi.FormationConfig{
        FormationID: "my-bot",
        URL:         "http://localhost:8001",
        ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
    })

    // Or proxy mode (production) - via MUXI server
    client = muxi.NewFormationClient(&muxi.FormationConfig{
        FormationID: "my-bot",
        ServerURL:   os.Getenv("MUXI_SERVER_URL"),
        ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
    })

    // Non-streaming chat
    resp, err := client.Chat(ctx, &muxi.ChatRequest{
        Message: "Hello, how are you?",
        UserID:  "user-123",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Message)

    // Streaming chat
    stream, err := client.ChatStream(ctx, &muxi.ChatRequest{
        Message: "Tell me a story",
        UserID:  "user-123",
    })
    if err != nil {
        panic(err)
    }
    for chunk := range stream {
        if chunk.Type == "text" {
            fmt.Print(chunk.Text)
        }
    }
}
```

### Streaming Deployment

```go
// Deploy with progress updates
events, err := server.DeployFormationStreaming(ctx, &muxi.DeployRequest{
    BundlePath: "my-bot.tar.gz",
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
```

## Configuration

The SDK does **not** auto-read environment variables; pass values explicitly. Examples below show `os.Getenv` only as a convenience pattern.

### Passing config (example using env values)

| Variable (example) | Description |
|--------------------|-------------|
| `MUXI_SERVER_URL` | Server URL for ServerClient |
| `MUXI_KEY_ID` | HMAC key ID for server auth |
| `MUXI_SECRET_KEY` | HMAC secret key for server auth |
| `MUXI_CLIENT_KEY` | Formation client key |
| `MUXI_ADMIN_KEY` | Formation admin key (optional) |
| `MUXI_DEBUG` | Enable debug logging (optional) |

### Client Options

```go
// ServerClient with retry
server := muxi.NewServerClient(&muxi.ServerConfig{
    URL:        "https://muxi.company.com:7890",
    KeyID:      "MUXI_KEY",
    SecretKey:  "sk-...",
    MaxRetries: 3,
    Timeout:    60 * time.Second,
})

// FormationClient with custom timeout
client := muxi.NewFormationClient(&muxi.FormationConfig{
    FormationID: "my-bot",
    URL:         "http://localhost:8001",
    ClientKey:   "fmc-...",
    Timeout:     30 * time.Second,
})
```

## API Reference

### ServerClient Methods (implemented)

| Method | Description |
|--------|-------------|
| `DeployFormation` | Deploy a new formation |
| `DeployFormationStreaming` | Deploy with progress events |
| `ListFormations` | List all formations |
| `GetFormation` | Get formation details |
| `UpdateFormation` | Update existing formation |
| `StopFormation` | Stop a running formation |
| `StartFormation` | Start a stopped formation |
| `RestartFormation` | Restart a formation |
| `RollbackFormation` | Rollback to previous version |
| `DeleteFormation` | Delete a formation |
| `CancelUpdate` | Cancel an ongoing update |
| `Status` | Get server status |
| `Health` | Health check |
| `GetServerLogs` | Fetch server audit logs (text) |
| `GetFormationLogs` | Fetch formation logs (non-streaming) |
| `StreamFormationLogs` | Stream formation logs (SSE) |

### FormationClient Methods (implemented)

| Category | Methods |
|----------|---------|
| **Chat** | `Chat`, `ChatStream`, `AudioChat`, `AudioChatStream` |
| **Config** | `Health`, `GetStatus`, `GetConfig`, `GetFormationInfo` |
| **Agents** | `GetAgents`, `GetAgent` |
| **Secrets** | `GetSecrets`, `GetSecret`, `SetSecret`, `DeleteSecret` |
| **MCP** | `GetMCPServers`, `GetMCPServer`, `GetMCPTools` |
| **Sessions/Requests** | `GetSessions`, `GetSession`, `GetSessionMessages`, `RestoreSession`, `GetRequests`, `GetRequestStatus`, `CancelRequest`, `StreamRequest` |
| **Users/Identifiers** | `ResolveUser`, `GetUserIdentifiers`, `GetUserIdentifiersForUser`, `LinkUserIdentifier`, `UnlinkUserIdentifier` |
| **Credentials** | `ListCredentials`, `GetCredential`, `CreateCredential`, `DeleteCredential`, `ListCredentialServices` |
| **Triggers/SOPs/Audit** | `GetTriggers`, `GetTrigger`, `FireTrigger`, `GetSOPs`, `GetSOP`, `GetAuditLog`, `ClearAuditLog` |
| **Memory** | `GetMemoryConfig`, `GetMemories`, `AddMemory`, `DeleteMemory`, `GetUserBuffer`, `ClearUserBuffer`, `ClearSessionBuffer`, `ClearAllBuffers`, `GetMemoryBuffers`, `GetBufferStats` |
| **Scheduler** | `GetSchedulerConfig`, `GetSchedulerJobs`, `GetSchedulerJob`, `CreateSchedulerJob`, `DeleteSchedulerJob` |
| **Async/A2A/Logging** | `GetAsyncConfig`, `GetAsyncJobs`, `GetAsyncJob`, `CancelAsyncJob`, `GetA2AConfig`, `GetLoggingConfig`, `GetLoggingDestinations` |
| **Overlord/LLM** | `GetOverlordConfig`, `GetOverlordPersona`, `GetLLMSettings` |
| **Events/Logs Streaming** | `StreamEvents`, `StreamLogs` |

## Error Handling

```go
resp, err := client.Chat(ctx, req)
if err != nil {
    switch e := err.(type) {
    case *muxi.AuthenticationError:
        log.Fatal("Invalid credentials")
    case *muxi.RateLimitError:
        log.Printf("Rate limited, retry after %d seconds", e.RetryAfter)
    case *muxi.NotFoundError:
        log.Printf("Not found: %s", e.Message)
    case *muxi.ValidationError:
        log.Printf("Invalid request: %s", e.Message)
    default:
        log.Printf("Error: %v", err)
    }
}
```

## Response Metadata

All responses include metadata for debugging:

```go
resp, err := client.Chat(ctx, req)
if err == nil {
    fmt.Printf("Request ID: %s\n", resp.RequestID)   // For support tickets
    fmt.Printf("Timestamp: %d\n", resp.Timestamp)    // Server time
}
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Links

- [MUXI Documentation](https://muxi.org/docs)
- [API Reference](https://muxi.org/docs/api-reference)
- [GitHub](https://github.com/muxi-ai/muxi-go)
