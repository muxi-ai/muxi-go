## AGENTS GUIDE (muxi-go)

Purpose: fast orientation for AI coding agents contributing to the Go SDK.

### Project layout
- Go module root: `go/src` (run all Go commands from this directory).
- Key files: `server_client.go`, `formation_client.go`, `streaming.go`, `transport.go`, `auth.go`, `errors.go`, `formation_models.go`, `server_models.go`.

### Required practices
- Always format: `gofmt -w <files>` from `go/src`.
- Always test after code changes: `go test ./...` from `go/src`.
- Keep auto idempotency header (`X-Muxi-Idempotency-Key`) on **every** request; no opt-outs.
- Preserve streaming semantics: infinite timeouts for SSE; reuse configured transport; large scanner buffers already set.
- Respect auth rules: HMAC strips query when signing; client key/admin key headers must remain.

### Common tasks
- Add methods using existing patterns (`formationRequest`, `formationRequestNoBody`, `do`, `doJSON`, streaming helpers).
- Use `decodeFormation`/`decodeServerAPI` to surface `RequestID`/`Timestamp` metadata.
- For new endpoints, add models in `formation_models.go` or `server_models.go` with `RequestID`/`Timestamp` fields.

### Git hygiene
- Check status inside submodule: `git status --short` (from `go`).
- Commit here first, then update parent pointer from repo root (`git add go`).

### Cautions
- Do not introduce new dependencies without approval.
- Avoid editing README unless requested.
- Keep debug logging minimal (method/URL/status/duration only).
