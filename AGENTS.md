<coding_guidelines>
## AGENTS GUIDE (muxi-go)

Purpose: fast orientation for AI coding agents contributing to the Go SDK.

### Project structure
```
go/
├── src/                    # Go module root (run all commands here)
│   ├── server_client.go    # ServerClient: HMAC auth, formation lifecycle, server logs
│   ├── formation_client.go # FormationClient: key auth, chat/audio, memory, scheduler, etc.
│   ├── server_models.go    # Types for server API responses
│   ├── formation_models.go # Types for formation API responses
│   ├── transport.go        # HTTP transport, retries, timeouts
│   ├── streaming.go        # SSE streaming helpers
│   ├── auth.go             # HMAC signature generation
│   ├── errors.go           # Typed error hierarchy
│   ├── version.go          # SDK version constant
│   └── examples/           # Working examples
├── AGENTS.md
├── USER_GUIDE.md
└── README.md
```

### Quick commands
```bash
cd go/src
go test ./...              # Run all tests
gofmt -w .                 # Format code
go build ./...             # Verify compilation
```

### Key patterns
- **ServerClient**: HMAC auth with `keyId`/`secretKey` for `/rpc` endpoints
- **FormationClient**: `X-MUXI-CLIENT-KEY` or `X-MUXI-ADMIN-KEY` headers for `/api/{formation}/v1`
- **Streaming**: Returns `<-chan T` + `<-chan error`; context cancellation stops streams
- **Retries**: GET/DELETE only, exponential backoff, respects `Retry-After`
- **Idempotency**: Auto `X-Muxi-Idempotency-Key` on every request (never remove)

### Adding new endpoints
1. Add request/response types to `formation_models.go` or `server_models.go`
2. Add method using existing patterns (`formationRequest`, `do`, `doJSON`)
3. Include `RequestID`/`Timestamp` fields for metadata
4. Run `go test ./...` before committing

### Git workflow
```bash
cd go
git status --short         # Check changes
git add . && git commit -m "..."
git push origin develop
# Then from sdks root: git add go && git commit -m "Update go submodule"
```

### Rules
- Keep idempotency header on every request — no toggles
- Streaming uses infinite timeouts — no per-request deadlines
- Do not add dependencies without approval
- Do not edit README unless requested
- Format with `gofmt` before committing
</coding_guidelines>
