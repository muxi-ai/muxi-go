package muxi

import (
    "encoding/json"
    "time"
)

// Formation API envelope
type FormationAPIResponse struct {
    Object    string          `json:"object"`
    Timestamp int64           `json:"timestamp"`
    Type      string          `json:"type"`
    Request   RequestInfo     `json:"request"`
    Success   bool            `json:"success"`
    Error     *APIError       `json:"error,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
}

type RequestInfo struct {
    ID             string `json:"id"`
    IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type APIError struct {
    Code    string          `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data,omitempty"`
}

// HealthResponse from GET /health
type HealthResponse struct {
    Status      string `json:"status"`
    FormationID string `json:"formation_id,omitempty"`
    Version     string `json:"version,omitempty"`
    Timestamp   string `json:"timestamp,omitempty"`
}

// StatusResponse from GET /status
type StatusResponse struct {
    Formation struct {
        ID          string `json:"id"`
        Name        string `json:"name"`
        Description string `json:"description"`
        Version     string `json:"version"`
    } `json:"formation"`
}

// ConfigResponse from GET /config
type ConfigResponse struct {
    FormationID   string `json:"formation_id"`
    Version       string `json:"version"`
    Description   string `json:"description"`
    SchemaVersion string `json:"schema_version"`
}

// FormationInfoResponse from GET /formation
type FormationInfoResponse struct {
    FormationID string `json:"formation_id"`
    Name        string `json:"name"`
    Version     string `json:"version"`
}

// Agent represents an agent
type Agent struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Role        string   `json:"role"`
    Description string   `json:"description,omitempty"`
    Model       string   `json:"model,omitempty"`
    Provider    string   `json:"provider,omitempty"`
    Enabled     bool     `json:"enabled"`
    Status      string   `json:"status,omitempty"`
    Tools       []string `json:"tools,omitempty"`
    MCPServers  []string `json:"mcp_servers,omitempty"`
}

type AgentListResponse struct {
    Agents []Agent `json:"agents"`
    Count  int     `json:"count"`
}

// ChatFile attachment
type ChatFile struct {
    Filename    string `json:"filename"`
    Content     string `json:"content"`      // Base64
    ContentType string `json:"content_type"` // MIME
    Size        int64  `json:"size,omitempty"`
}

// ChatRequest for /chat
type ChatRequest struct {
    Message          string     `json:"message"`
    UserID           string     `json:"user_id,omitempty"`
    SessionID        string     `json:"session_id,omitempty"`
    GroupID          string     `json:"group_id,omitempty"`
    Stream           bool       `json:"stream,omitempty"`
    UseAsync         *bool      `json:"use_async,omitempty"`
    WebhookURL       string     `json:"webhook_url,omitempty"`
    ThresholdSeconds int        `json:"threshold_seconds,omitempty"`
    Files            []ChatFile `json:"files,omitempty"`
}

// AudioChatRequest for /audiochat
type AudioChatRequest struct {
    Files     []ChatFile `json:"files"`
    UserID    string     `json:"user_id,omitempty"`
    SessionID string     `json:"session_id,omitempty"`
    AgentName string     `json:"agent_name,omitempty"`
    Stream    bool       `json:"stream,omitempty"`
}

// ChatResponse from chat/audiochat
type ChatResponse struct {
    RequestID string `json:"request_id"`
    SessionID string `json:"session_id"`
    Response  string `json:"response"`
    Agent     string `json:"agent,omitempty"`
    Model     string `json:"model,omitempty"`
    Usage     struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage,omitempty"`
}

// ChatChunk represents streaming chat chunk
type ChatChunk struct {
    Type       string          `json:"type"` // text, tool_call, tool_result, error, done
    Text       string          `json:"text,omitempty"`
    ToolCall   json.RawMessage `json:"tool_call,omitempty"`
    ToolResult json.RawMessage `json:"tool_result,omitempty"`
    Error      string          `json:"error,omitempty"`
}

// Secrets
type SecretsListResponse struct {
    Secrets map[string]string `json:"secrets"`
    Count   int               `json:"count"`
}

type SecretResponse struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// MCP
type MCPServer struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Status      string   `json:"status"`
    Enabled     bool     `json:"enabled"`
    Tools       []string `json:"tools,omitempty"`
    Description string   `json:"description,omitempty"`
}

type MCPListResponse struct {
    Servers []MCPServer `json:"servers"`
    Count   int         `json:"count"`
}

type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Server      string                 `json:"server"`
    InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

type MCPToolsResponse struct {
    Tools []MCPTool `json:"tools"`
    Count int       `json:"count"`
}

// Sessions
type Session struct {
    ID           string     `json:"session_id"`
    UserID       string     `json:"user_id,omitempty"`
    LastActivity *time.Time `json:"last_activity,omitempty"`
    Active       bool       `json:"active,omitempty"`
    CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type SessionsListResponse struct {
    Sessions []Session `json:"sessions"`
    Count    int       `json:"count"`
    HasMore  bool      `json:"has_more"`
}

type Message struct {
    ID        string     `json:"id,omitempty"`
    Text      string     `json:"text"`
    Content   string     `json:"content,omitempty"`
    Timestamp *time.Time `json:"timestamp,omitempty"`
    Metadata  *struct {
        Role      string `json:"role"`
        UserID    string `json:"user_id,omitempty"`
        SessionID string `json:"session_id,omitempty"`
        AgentID   string `json:"agent_id,omitempty"`
    } `json:"metadata,omitempty"`
}

type SessionMessagesResponse struct {
    SessionID string    `json:"session_id"`
    Messages  []Message `json:"messages"`
    Count     int       `json:"count"`
}

// Requests
type RequestItem struct {
    RequestID string     `json:"request_id"`
    Status    string     `json:"status"`
    Progress  int        `json:"progress,omitempty"`
    CreatedAt *time.Time `json:"created_at,omitempty"`
}

type RequestsListResponse struct {
    Requests []RequestItem `json:"requests"`
    Count    int           `json:"count"`
}

type RequestStatusResponse struct {
    RequestID string     `json:"request_id"`
    Status    string     `json:"status"`
    Progress  string     `json:"progress,omitempty"`
    Error     string     `json:"error,omitempty"`
    Completed *time.Time `json:"completed_at,omitempty"`
}

// User resolve
type UserResolveRequest struct {
    Identifier string `json:"identifier"`
    CreateUser bool   `json:"create_user,omitempty"`
}

type UserResolveResponse struct {
    Identifier     string `json:"identifier"`
    MuxiUserID     string `json:"muxi_user_id"`
    InternalUserID int    `json:"internal_user_id"`
}

// Triggers/SOPs
type TriggersListResponse struct {
    Triggers []string `json:"triggers"`
    Count    int      `json:"count"`
}

type TriggerDetail struct {
    Name       string   `json:"name"`
    Content    string   `json:"content"`
    DataFields []string `json:"data_fields,omitempty"`
}

type TriggerRequest struct {
    Data      json.RawMessage `json:"data"`
    SessionID string          `json:"session_id,omitempty"`
    UseAsync  bool            `json:"use_async"`
}

type TriggerResponse struct {
    RequestID string `json:"-"`
    Status    string `json:"status"`
    Content   string `json:"content,omitempty"`
}

type SOP struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Type        string `json:"type,omitempty"`
    Steps       int    `json:"steps,omitempty"`
    Content     string `json:"content,omitempty"`
}

type SOPsListResponse struct {
    SOPs  []SOP `json:"sops"`
    Count int   `json:"count"`
}

// Audit
type AuditEntry struct {
    Timestamp    *time.Time `json:"timestamp,omitempty"`
    RequestID    string     `json:"request_id,omitempty"`
    Action       string     `json:"action"`
    ResourceType string     `json:"resource_type,omitempty"`
    ResourceID   string     `json:"resource_id,omitempty"`
    User         string     `json:"user,omitempty"`
    IP           string     `json:"ip,omitempty"`
    Result       string     `json:"result,omitempty"`
    StatusCode   int        `json:"status_code,omitempty"`
    Message      string     `json:"message,omitempty"`
}

type AuditLogResponse struct {
    Entries      []AuditEntry `json:"entries"`
    Count        int          `json:"count"`
    TotalEntries int          `json:"total_entries,omitempty"`
}

// Credentials
type Credential struct {
    CredentialID      string    `json:"credential_id"`
    Service           string    `json:"service"`
    Name              string    `json:"name"`
    CredentialPreview string    `json:"credential_preview"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

type CredentialsListResponse struct {
    Credentials []Credential `json:"credentials"`
    Count       int          `json:"count"`
}

type CreateCredentialRequest struct {
    Service    string                 `json:"service"`
    Name       string                 `json:"name,omitempty"`
    Credential map[string]interface{} `json:"credential"`
}

type CreateCredentialResponse struct {
    CredentialID      string    `json:"credential_id"`
    Service           string    `json:"service"`
    Name              string    `json:"name"`
    CredentialPreview string    `json:"credential_preview"`
    CreatedAt         time.Time `json:"created_at"`
}

type DeleteCredentialResponse struct {
    CredentialID string `json:"credential_id"`
    Deleted      bool   `json:"deleted"`
}

// Logging stream
type LogStreamEvent struct {
    Timestamp int64                  `json:"timestamp"`
    Level     string                 `json:"level"`
    EventType string                 `json:"event_type,omitempty"`
    UserID    string                 `json:"user_id,omitempty"`
    SessionID string                 `json:"session_id,omitempty"`
    RequestID string                 `json:"request_id,omitempty"`
    AgentID   string                 `json:"agent_id,omitempty"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
}
