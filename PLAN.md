## Go SDK Gap Remediation Plan (2026-01-06)

### Scope
Bring the Go SDK into alignment with SDK-DESIGN, SDK-CONVENTIONS, go/DESIGN, and current implementation reality. Covers gaps, mismatches, and the noted non-blocking observation.

### Items
1) Chat streaming chunk types
   - Extend `ChatChunk` to include convention types (`text`, `tool_call`, `tool_result`, `agent_handoff`, `thinking`, `error`, `done`) with appropriate payload fields.
   - Update SSE parsing to honor `event:` types and map into `ChatChunk.Type`, emitting `done` on clean close.

2) Debug logging toggle
   - Add `Debug bool` and optional `Logger` to `ServerConfig` / `FormationConfig`; default to `log.Default()` when `Debug` or `MUXI_DEBUG` is set.
   - Add lightweight request/response summaries in transports (method, URL, status, duration); no body logging.

3) Response metadata
   - Ensure all responses surface `RequestID` and `Timestamp` from envelopes (Server and Formation). Add fields and populate during decode.

4) README/API surface alignment
   - Update `go/README.md` to match implemented methods.
   - Move unimplemented areas (memory, scheduler, overlord/persona/LLM, events, etc.) to a "Planned / Not yet implemented" section.

5) HMAC query handling (non-blocking)
   - Keep current signer behavior (strip query when signing) but document it and ensure callers pass the request path consistently (streaming logs included).

6) Validation
   - Run `go test ./...` after changes.

### Sequence
1) Chat chunk/type expansion + SSE parser update.
2) Debug/logging toggle in transports.
3) Response metadata surfacing.
4) README alignment.
5) HMAC comment clarifications.
6) Tests and commits.
