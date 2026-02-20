package muxi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

// FormationConfig configures FormationClient
type FormationConfig struct {
	FormationID string
	URL         string // direct mode (http://localhost:8001)
	ServerURL   string // proxy mode (https://server/api/{id}/v1)
	BaseURL     string // explicit override
	AdminKey    string
	ClientKey   string
	MaxRetries  int
	Timeout     time.Duration
	HTTPClient  *http.Client
	Debug       bool
	Logger      *log.Logger
	Mode        string // "live" (default) or "draft" for local dev
	App         string // internal: for Console telemetry (undocumented)
}

// FormationClient is an HTTP client for Formation API
type FormationClient struct {
	baseURL    string
	adminKey   string
	clientKey  string
	httpClient *http.Client
	maxRetries int
}

// NewFormationClient constructs a FormationClient
func NewFormationClient(cfg *FormationConfig) *FormationClient {
	if cfg == nil {
		panic("FormationConfig is required")
	}

	base := cfg.BaseURL
	if base == "" {
		switch {
		case cfg.URL != "":
			base = cfg.URL + "/v1"
		case cfg.ServerURL != "" && cfg.FormationID != "":
			prefix := "api"
			if cfg.Mode == "draft" {
				prefix = "draft"
			}
			base = fmt.Sprintf("%s/%s/%s/v1", cfg.ServerURL, prefix, cfg.FormationID)
		default:
			panic("must set BaseURL, URL, or ServerURL+FormationID")
		}
	}

	debug := cfg.Debug || os.Getenv("MUXI_DEBUG") != ""
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := newSDKTransport(http.DefaultTransport, logger, debug, cfg.App)
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout, Transport: transport}
	} else {
		baseTr := client.Transport
		if baseTr == nil {
			baseTr = http.DefaultTransport
		}
		client = &http.Client{Timeout: client.Timeout, Transport: newSDKTransport(baseTr, logger, debug, cfg.App)}
	}

	return &FormationClient{
		baseURL:    base,
		adminKey:   cfg.AdminKey,
		clientKey:  cfg.ClientKey,
		httpClient: client,
		maxRetries: cfg.MaxRetries,
	}
}

// Health checks formation health (no auth)
func (c *FormationClient) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, mapStatusToError(resp.StatusCode, resp.Body)
	}

	var h HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, err
	}
	return &h, nil
}

// GetStatus returns formation status
func (c *FormationClient) GetStatus(ctx context.Context) (*StatusResponse, error) {
	return formationRequest[StatusResponse](ctx, c, http.MethodGet, "/status", nil, true, "")
}

// GetConfig returns formation config
func (c *FormationClient) GetConfig(ctx context.Context) (*ConfigResponse, error) {
	return formationRequest[ConfigResponse](ctx, c, http.MethodGet, "/config", nil, true, "")
}

// GetFormationInfo returns basic formation info
func (c *FormationClient) GetFormationInfo(ctx context.Context) (*FormationInfoResponse, error) {
	return formationRequest[FormationInfoResponse](ctx, c, http.MethodGet, "/formation", nil, true, "")
}

// GetAgents lists agents
func (c *FormationClient) GetAgents(ctx context.Context) (*AgentListResponse, error) {
	return formationRequest[AgentListResponse](ctx, c, http.MethodGet, "/agents", nil, true, "")
}

// GetAgent gets a single agent (raw map)
func (c *FormationClient) GetAgent(ctx context.Context, id string) (map[string]interface{}, error) {
	res, err := formationRequest[map[string]interface{}](ctx, c, http.MethodGet, "/agents/"+id, nil, true, "")
	if err != nil {
		return nil, err
	}
	return *res, nil
}

// GetSecrets lists secrets (masked)
func (c *FormationClient) GetSecrets(ctx context.Context) (*SecretsListResponse, error) {
	return formationRequest[SecretsListResponse](ctx, c, http.MethodGet, "/secrets", nil, true, "")
}

// GetSecret gets a secret (masked value)
func (c *FormationClient) GetSecret(ctx context.Context, key string) (*SecretResponse, error) {
	return formationRequest[SecretResponse](ctx, c, http.MethodGet, "/secrets/"+key, nil, true, "")
}

// SetSecret sets a secret value
func (c *FormationClient) SetSecret(ctx context.Context, key, value string) error {
	body := map[string]string{"value": value}
	return formationRequestNoBody(ctx, c, http.MethodPut, "/secrets/"+key, body, true, "")
}

// DeleteSecret deletes a secret
func (c *FormationClient) DeleteSecret(ctx context.Context, key string) error {
	return formationRequestNoBody(ctx, c, http.MethodDelete, "/secrets/"+key, nil, true, "")
}

// GetMCPServers lists MCP servers
func (c *FormationClient) GetMCPServers(ctx context.Context) (*MCPListResponse, error) {
	return formationRequest[MCPListResponse](ctx, c, http.MethodGet, "/mcp/servers", nil, true, "")
}

// GetMCPServer gets an MCP server (raw map)
func (c *FormationClient) GetMCPServer(ctx context.Context, id string) (map[string]interface{}, error) {
	res, err := formationRequest[map[string]interface{}](ctx, c, http.MethodGet, "/mcp/servers/"+id, nil, true, "")
	if err != nil {
		return nil, err
	}
	return *res, nil
}

// GetMCPTools lists MCP tools
func (c *FormationClient) GetMCPTools(ctx context.Context) (*MCPToolsResponse, error) {
	return formationRequest[MCPToolsResponse](ctx, c, http.MethodGet, "/mcp/tools", nil, true, "")
}

// Chat sends a chat request (non-streaming)
func (c *FormationClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := c.doJSON(ctx, http.MethodPost, "/chat", req, false, req.UserID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[ChatResponse](resp)
}

// ChatStream streams chat responses
func (c *FormationClient) ChatStream(ctx context.Context, req *ChatRequest) (<-chan ChatChunk, <-chan error) {
	req.Stream = true
	return c.streamChat(ctx, "/chat", req, req.UserID)
}

// AudioChat sends audio (non-streaming)
func (c *FormationClient) AudioChat(ctx context.Context, req *AudioChatRequest) (*ChatResponse, error) {
	resp, err := c.doJSON(ctx, http.MethodPost, "/audiochat", req, false, req.UserID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[ChatResponse](resp)
}

// AudioChatStream streams audio chat responses
func (c *FormationClient) AudioChatStream(ctx context.Context, req *AudioChatRequest) (<-chan ChatChunk, <-chan error) {
	req.Stream = true
	return c.streamChat(ctx, "/audiochat", req, req.UserID)
}

// GetSessions lists sessions for a user
func (c *FormationClient) GetSessions(ctx context.Context, userID string, limit int) (*SessionsListResponse, error) {
	path := "/sessions"
	if limit > 0 {
		path = fmt.Sprintf("/sessions?limit=%d", limit)
	}
	return formationRequest[SessionsListResponse](ctx, c, http.MethodGet, path, nil, false, userID)
}

// GetSessionMessages gets messages for a session
func (c *FormationClient) GetSessionMessages(ctx context.Context, sessionID, userID string) (*SessionMessagesResponse, error) {
	return formationRequest[SessionMessagesResponse](ctx, c, http.MethodGet, "/sessions/"+sessionID+"/messages", nil, false, userID)
}

// RestoreSession restores messages into a session
func (c *FormationClient) RestoreSession(ctx context.Context, sessionID, userID string, messages []Message) error {
	payload := struct {
		Messages []Message `json:"messages"`
	}{Messages: messages}
	return formationRequestNoBody(ctx, c, http.MethodPost, "/sessions/"+sessionID+"/restore", payload, false, userID)
}

// GetRequests lists requests for a user
func (c *FormationClient) GetRequests(ctx context.Context, userID string) (*RequestsListResponse, error) {
	return formationRequest[RequestsListResponse](ctx, c, http.MethodGet, "/requests", nil, false, userID)
}

// CancelRequest cancels a request
func (c *FormationClient) CancelRequest(ctx context.Context, requestID, userID string) error {
	return formationRequestNoBody(ctx, c, http.MethodDelete, "/requests/"+requestID, nil, false, userID)
}

// GetRequestStatus returns request status
func (c *FormationClient) GetRequestStatus(ctx context.Context, requestID, userID string) (*RequestStatusResponse, error) {
	return formationRequest[RequestStatusResponse](ctx, c, http.MethodGet, "/requests/"+requestID, nil, false, userID)
}

// Memory APIs
func (c *FormationClient) GetMemoryConfig(ctx context.Context) (*MemoryConfigResponse, error) {
	return formationRequest[MemoryConfigResponse](ctx, c, http.MethodGet, "/memory", nil, true, "")
}

func (c *FormationClient) GetMemories(ctx context.Context, userID string) (*MemoriesListResponse, error) {
	path := "/memories"
	if userID != "" {
		path += "?user_id=" + userID
	}
	return formationRequest[MemoriesListResponse](ctx, c, http.MethodGet, path, nil, false, "")
}

func (c *FormationClient) AddMemory(ctx context.Context, userID, memType, detail string) (*Memory, error) {
	body := map[string]string{"user_id": userID, "type": memType, "detail": detail}
	resp, err := c.doJSON(ctx, http.MethodPost, "/memories", body, false, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[Memory](resp)
}

// DeleteMemory removes a memory by ID, optionally scoped to a user.
func (c *FormationClient) DeleteMemory(ctx context.Context, userID, memoryID string) error {
	path := "/memories/" + memoryID
	if userID != "" {
		path += "?user_id=" + userID
	}
	return formationRequestNoBody(ctx, c, http.MethodDelete, path, nil, false, "")
}

func (c *FormationClient) GetUserBuffer(ctx context.Context, userID string) (*UserBufferResponse, error) {
	path := "/memory/buffer"
	if userID != "" {
		path += "?user_id=" + userID
	}
	return formationRequest[UserBufferResponse](ctx, c, http.MethodGet, path, nil, false, "")
}

// ClearUserBuffer clears all buffered content for a user (admin-only when no user ID).
func (c *FormationClient) ClearUserBuffer(ctx context.Context, userID string) (*BufferClearedResponse, error) {
	path := "/memory/buffer"
	if userID != "" {
		path += "?user_id=" + userID
	}
	resp, err := c.do(ctx, http.MethodDelete, path, nil, false, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[BufferClearedResponse](resp)
}

// ClearAllBuffers clears all buffers across users (admin).
func (c *FormationClient) ClearAllBuffers(ctx context.Context) (*BufferClearedResponse, error) {
	resp, err := c.do(ctx, http.MethodDelete, "/memory/buffer", nil, true, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[BufferClearedResponse](resp)
}

// ClearSessionBuffer clears a single session buffer, optionally scoped to a user.
func (c *FormationClient) ClearSessionBuffer(ctx context.Context, userID, sessionID string) (*SessionBufferClearedResponse, error) {
	path := "/memory/buffer/" + sessionID
	if userID != "" {
		path += "?user_id=" + userID
	}
	resp, err := c.do(ctx, http.MethodDelete, path, nil, false, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[SessionBufferClearedResponse](resp)
}

// GetBufferStats returns aggregate buffer stats (admin).
func (c *FormationClient) GetBufferStats(ctx context.Context) (*BufferStatsResponse, error) {
	return formationRequest[BufferStatsResponse](ctx, c, http.MethodGet, "/memory/stats", nil, true, "")
}

// Scheduler APIs
func (c *FormationClient) GetSchedulerConfig(ctx context.Context) (*SchedulerConfigResponse, error) {
	return formationRequest[SchedulerConfigResponse](ctx, c, http.MethodGet, "/scheduler", nil, true, "")
}

func (c *FormationClient) GetSchedulerJobs(ctx context.Context, userID string) (*SchedulerJobsResponse, error) {
	path := "/scheduler/jobs"
	if userID != "" {
		path += "?user_id=" + userID
	}
	return formationRequest[SchedulerJobsResponse](ctx, c, http.MethodGet, path, nil, true, "")
}

func (c *FormationClient) GetSchedulerJob(ctx context.Context, jobID string) (*SchedulerJobDetail, error) {
	return formationRequest[SchedulerJobDetail](ctx, c, http.MethodGet, "/scheduler/jobs/"+jobID, nil, true, "")
}

func (c *FormationClient) CreateSchedulerJob(ctx context.Context, jobType, schedule, message, userID string) (*SchedulerJobDetail, error) {
	body := map[string]string{
		"type":     jobType,
		"schedule": schedule,
		"message":  message,
	}
	if userID != "" {
		body["user_id"] = userID
	}
	resp, err := c.doJSON(ctx, http.MethodPost, "/scheduler/jobs", body, true, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[SchedulerJobDetail](resp)
}

func (c *FormationClient) DeleteSchedulerJob(ctx context.Context, jobID string) error {
	return formationRequestNoBody(ctx, c, http.MethodDelete, "/scheduler/jobs/"+jobID, nil, true, "")
}

// Async / A2A / Logging
func (c *FormationClient) GetAsyncConfig(ctx context.Context) (*AsyncSettingsResponse, error) {
	return formationRequest[AsyncSettingsResponse](ctx, c, http.MethodGet, "/async", nil, true, "")
}

func (c *FormationClient) GetA2AConfig(ctx context.Context) (*A2AConfigResponse, error) {
	return formationRequest[A2AConfigResponse](ctx, c, http.MethodGet, "/a2a", nil, true, "")
}

func (c *FormationClient) GetLoggingConfig(ctx context.Context) (*LoggingConfigResponse, error) {
	return formationRequest[LoggingConfigResponse](ctx, c, http.MethodGet, "/logging", nil, true, "")
}

func (c *FormationClient) GetLoggingDestinations(ctx context.Context) (*LoggingDestinationsResponse, error) {
	return formationRequest[LoggingDestinationsResponse](ctx, c, http.MethodGet, "/logging/destinations", nil, true, "")
}

// Credential services
func (c *FormationClient) ListCredentialServices(ctx context.Context) (*CredentialServicesResponse, error) {
	return formationRequest[CredentialServicesResponse](ctx, c, http.MethodGet, "/credentials/services", nil, true, "")
}

// User identifiers
func (c *FormationClient) GetUserIdentifiersForUser(ctx context.Context, userID string) (*UserIdentifiersResponse, error) {
	return formationRequest[UserIdentifiersResponse](ctx, c, http.MethodGet, "/users/identifiers/"+userID, nil, true, "")
}

func (c *FormationClient) LinkUserIdentifier(ctx context.Context, muxiUserID string, identifiers []interface{}) (*UserIdentifiersResponse, error) {
	body := map[string]interface{}{"muxi_user_id": muxiUserID, "identifiers": identifiers}
	resp, err := c.doJSON(ctx, http.MethodPost, "/users/identifiers", body, true, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[UserIdentifiersResponse](resp)
}

func (c *FormationClient) UnlinkUserIdentifier(ctx context.Context, identifier string) error {
	return formationRequestNoBody(ctx, c, http.MethodDelete, "/users/identifiers/"+identifier, nil, true, "")
}

// Overlord / LLM
func (c *FormationClient) GetOverlordConfig(ctx context.Context) (*OverlordConfigResponse, error) {
	return formationRequest[OverlordConfigResponse](ctx, c, http.MethodGet, "/overlord", nil, true, "")
}

func (c *FormationClient) GetOverlordSoul(ctx context.Context) (*OverlordSoulResponse, error) {
	return formationRequest[OverlordSoulResponse](ctx, c, http.MethodGet, "/overlord/soul", nil, true, "")
}

func (c *FormationClient) GetLLMSettings(ctx context.Context) (*LLMSettingsResponse, error) {
	return formationRequest[LLMSettingsResponse](ctx, c, http.MethodGet, "/llm/settings", nil, true, "")
}

// Sessions
func (c *FormationClient) GetSession(ctx context.Context, sessionID, userID string) (*SessionDetailResponse, error) {
	return formationRequest[SessionDetailResponse](ctx, c, http.MethodGet, "/sessions/"+sessionID, nil, false, userID)
}

type LogStreamFilters struct {
	UserID    string
	SessionID string
	RequestID string
	AgentID   string
	Level     string
	EventType string
}

func (c *FormationClient) streamLogEvents(ctx context.Context, path string, headers map[string]string) (<-chan LogStreamEvent, <-chan error) {
	out := make(chan LogStreamEvent)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			errs <- err
			return
		}
		if c.clientKey == "" {
			errs <- fmt.Errorf("client key required")
			return
		}
		req.Header.Set("X-MUXI-CLIENT-KEY", c.clientKey)
		req.Header.Set("Accept", "text/event-stream")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		baseTr := c.httpClient.Transport
		if baseTr == nil {
			baseTr = http.DefaultTransport
		}
		client := &http.Client{Timeout: 0, Transport: baseTr}
		resp, err := client.Do(req)
		if err != nil {
			errs <- &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			errs <- checkFormationHTTP(resp)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)
		var dataBuf []string

		flush := func() error {
			if len(dataBuf) == 0 {
				return nil
			}
			payload := strings.Join(dataBuf, "")
			dataBuf = dataBuf[:0]
			var ev LogStreamEvent
			if err := json.Unmarshal([]byte(payload), &ev); err != nil {
				return fmt.Errorf("failed to parse event: %w", err)
			}
			out <- ev
			return nil
		}

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				dataBuf = append(dataBuf, data)
			}
			if line == "" {
				if err := flush(); err != nil {
					errs <- err
					return
				}
			}
		}
		if err := flush(); err != nil {
			errs <- err
			return
		}
		if err := scanner.Err(); err != nil {
			errs <- err
			return
		}
	}()

	return out, errs
}

// Events streaming
func (c *FormationClient) StreamEvents(ctx context.Context, userID string) (<-chan LogStreamEvent, <-chan error) {
	path := "/events"
	if userID != "" {
		path += "?user_id=" + userID
	}

	return c.streamLogEvents(ctx, path, nil)
}

// ResolveUser resolves an identifier
func (c *FormationClient) ResolveUser(ctx context.Context, identifier string, createUser bool) (*UserResolveResponse, error) {
	body := UserResolveRequest{Identifier: identifier, CreateUser: createUser}
	resp, err := c.doJSON(ctx, http.MethodPost, "/users/resolve", body, false, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[UserResolveResponse](resp)
}

// StreamRequest streams SSE events for a specific request
func (c *FormationClient) StreamRequest(ctx context.Context, userID, sessionID, requestID string) (<-chan LogStreamEvent, <-chan error) {
	path := "/events/" + sessionID + "/" + requestID
	headers := map[string]string{}
	if userID != "" {
		headers["X-Muxi-User-ID"] = userID
	}
	return c.streamLogEvents(ctx, path, headers)
}

// StreamLogs streams runtime logs with optional filters
func (c *FormationClient) StreamLogs(ctx context.Context, filters *LogStreamFilters) (<-chan LogStreamEvent, <-chan error) {
	params := url.Values{}
	if filters != nil {
		if filters.UserID != "" {
			params.Set("user_id", filters.UserID)
		}
		if filters.SessionID != "" {
			params.Set("session_id", filters.SessionID)
		}
		if filters.RequestID != "" {
			params.Set("request_id", filters.RequestID)
		}
		if filters.AgentID != "" {
			params.Set("agent_id", filters.AgentID)
		}
		if filters.Level != "" {
			params.Set("level", filters.Level)
		}
		if filters.EventType != "" {
			params.Set("event_type", filters.EventType)
		}
	}
	path := "/logs"
	if qs := params.Encode(); qs != "" {
		path += "?" + qs
	}
	return c.streamLogEvents(ctx, path, nil)
}

// Triggers
func (c *FormationClient) GetTriggers(ctx context.Context) (*TriggersListResponse, error) {
	return formationRequest[TriggersListResponse](ctx, c, http.MethodGet, "/triggers", nil, false, "")
}

func (c *FormationClient) GetTrigger(ctx context.Context, name string) (*TriggerDetail, error) {
	return formationRequest[TriggerDetail](ctx, c, http.MethodGet, "/triggers/"+name, nil, false, "")
}

func (c *FormationClient) FireTrigger(ctx context.Context, name string, data json.RawMessage, async bool, userID string) (*TriggerResponse, error) {
	body := TriggerRequest{Data: data, UseAsync: async}
	resp, err := c.doJSON(ctx, http.MethodPost, "/triggers/"+name, body, false, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	apiResp, err := decodeFormationResp(resp)
	if err != nil {
		return nil, err
	}
	var tr TriggerResponse
	if err := json.Unmarshal(apiResp.Data, &tr); err != nil {
		return nil, err
	}
	tr.RequestID = apiResp.Request.ID
	return &tr, nil
}

// SOPs
func (c *FormationClient) GetSOPs(ctx context.Context) (*SOPsListResponse, error) {
	return formationRequest[SOPsListResponse](ctx, c, http.MethodGet, "/sops", nil, false, "")
}

func (c *FormationClient) GetSOP(ctx context.Context, name string) (*SOP, error) {
	return formationRequest[SOP](ctx, c, http.MethodGet, "/sops/"+name, nil, false, "")
}

// Audit
func (c *FormationClient) GetAuditLog(ctx context.Context) (*AuditLogResponse, error) {
	return formationRequest[AuditLogResponse](ctx, c, http.MethodGet, "/audit", nil, true, "")
}

func (c *FormationClient) ClearAuditLog(ctx context.Context) error {
	return formationRequestNoBody(ctx, c, http.MethodDelete, "/audit?confirm=clear-audit-log", nil, true, "")
}

// Credentials
func (c *FormationClient) ListCredentials(ctx context.Context, userID string) (*CredentialsListResponse, error) {
	return formationRequest[CredentialsListResponse](ctx, c, http.MethodGet, "/credentials", nil, false, userID)
}

func (c *FormationClient) GetCredential(ctx context.Context, credentialID, userID string) (*Credential, error) {
	return formationRequest[Credential](ctx, c, http.MethodGet, "/credentials/"+credentialID, nil, false, userID)
}

func (c *FormationClient) CreateCredential(ctx context.Context, userID string, req *CreateCredentialRequest) (*CreateCredentialResponse, error) {
	resp, err := c.doJSON(ctx, http.MethodPost, "/credentials", req, false, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[CreateCredentialResponse](resp)
}

func (c *FormationClient) DeleteCredential(ctx context.Context, credentialID, userID string) (*DeleteCredentialResponse, error) {
	resp, err := c.do(ctx, http.MethodDelete, "/credentials/"+credentialID, nil, false, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[DeleteCredentialResponse](resp)
}

// --- helpers ---

func (c *FormationClient) doJSON(ctx context.Context, method, path string, body interface{}, useAdmin bool, userID string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	return c.do(ctx, method, path, reader, useAdmin, userID)
}

// do executes a request with retry logic
func (c *FormationClient) do(ctx context.Context, method, path string, body io.Reader, useAdmin bool, userID string) (*http.Response, error) {
	url := c.baseURL + path

	attempt := 0
	for {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}

		// Auth headers
		if useAdmin {
			if c.adminKey == "" {
				return nil, fmt.Errorf("admin key required")
			}
			req.Header.Set("X-MUXI-ADMIN-KEY", c.adminKey)
		} else {
			if c.clientKey == "" {
				return nil, fmt.Errorf("client key required")
			}
			req.Header.Set("X-MUXI-CLIENT-KEY", c.clientKey)
		}
		if userID != "" {
			req.Header.Set("X-Muxi-User-ID", userID)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
		}

		if !shouldRetry(method, resp.StatusCode, c.maxRetries, attempt) {
			return resp, nil
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		time.Sleep(backoffDelay(attempt))
		attempt++
	}
}

func decodeFormation[T any](resp *http.Response) (*T, error) {
	apiResp, err := decodeFormationResp(resp)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(apiResp.Data, &out); err != nil {
		return nil, err
	}
	setMetadata(&out, apiResp.Request.ID, apiResp.Timestamp)
	return &out, nil
}

// generic request helper
func formationRequest[T any](ctx context.Context, c *FormationClient, method, path string, body interface{}, useAdmin bool, userID string) (*T, error) {
	var resp *http.Response
	var err error
	if body != nil && method != http.MethodGet && method != http.MethodDelete {
		resp, err = c.doJSON(ctx, method, path, body, useAdmin, userID)
	} else {
		resp, err = c.do(ctx, method, path, nil, useAdmin, userID)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeFormation[T](resp)
}

func formationRequestNoBody(ctx context.Context, c *FormationClient, method, path string, body interface{}, useAdmin bool, userID string) error {
	var resp *http.Response
	var err error
	if body != nil {
		resp, err = c.doJSON(ctx, method, path, body, useAdmin, userID)
	} else {
		resp, err = c.do(ctx, method, path, nil, useAdmin, userID)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkFormationHTTP(resp)
}

func decodeFormationResp(resp *http.Response) (*FormationAPIResponse, error) {
	if err := checkFormationHTTP(resp); err != nil {
		return nil, err
	}
	var apiResp FormationAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, &MuxiError{Code: apiResp.Error.Code, Message: apiResp.Error.Message, StatusCode: resp.StatusCode}
		}
		return nil, &MuxiError{Code: ErrServerError, Message: "request failed", StatusCode: resp.StatusCode}
	}
	return &apiResp, nil
}

func checkFormationHTTP(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Try parse error envelope
	var apiResp FormationAPIResponse
	data, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(data, &apiResp) == nil && apiResp.Error != nil {
		return &MuxiError{Code: apiResp.Error.Code, Message: apiResp.Error.Message, StatusCode: resp.StatusCode}
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{newMuxiError(ErrUnauthorized, "authentication failed", resp.StatusCode)}
	case http.StatusForbidden:
		return &AuthorizationError{newMuxiError(ErrForbidden, "access denied", resp.StatusCode)}
	case http.StatusNotFound:
		return &NotFoundError{newMuxiError(ErrNotFound, "not found", resp.StatusCode)}
	case http.StatusConflict:
		return &ConflictError{newMuxiError(ErrConflict, "conflict", resp.StatusCode)}
	default:
		return &ServerError{newMuxiError(ErrServerError, fmt.Sprintf("server error: %d", resp.StatusCode), resp.StatusCode)}
	}
}

// setMetadata injects request metadata into response structs when fields exist
func setMetadata(dst interface{}, reqID string, ts int64) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return
	}
	if f := rv.FieldByName("RequestID"); f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
		f.SetString(reqID)
	}
	if f := rv.FieldByName("Timestamp"); f.IsValid() && f.CanSet() && f.Kind() == reflect.Int64 {
		f.SetInt(ts)
	}
	if f := rv.FieldByName("MetaTS"); f.IsValid() && f.CanSet() && f.Kind() == reflect.Int64 {
		f.SetInt(ts)
	}
}

// streaming helper
func (c *FormationClient) streamChat(ctx context.Context, path string, req interface{}, userID string) (<-chan ChatChunk, <-chan error) {
	out := make(chan ChatChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		// marshal
		data, err := json.Marshal(req)
		if err != nil {
			errs <- err
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
		if err != nil {
			errs <- err
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")
		httpReq.Header.Set("Cache-Control", "no-cache")
		httpReq.Header.Set("Connection", "keep-alive")
		if c.clientKey == "" {
			errs <- fmt.Errorf("client key required")
			return
		}
		httpReq.Header.Set("X-MUXI-CLIENT-KEY", c.clientKey)
		if userID != "" {
			httpReq.Header.Set("X-Muxi-User-ID", userID)
		}

		// no timeout for streaming; reuse configured transport
		baseTr := c.httpClient.Transport
		if baseTr == nil {
			baseTr = http.DefaultTransport
		}
		streamClient := &http.Client{Timeout: 0, Transport: baseTr}
		resp, err := streamClient.Do(httpReq)
		if err != nil {
			errs <- &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			errs <- checkFormationHTTP(resp)
			return
		}

		// parse SSE
		if err := parseChatSSE(resp.Body, out); err != nil {
			errs <- err
			return
		}
	}()

	return out, errs
}
