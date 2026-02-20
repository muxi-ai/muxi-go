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
	RequestID   string `json:"-"`
	MetaTS      int64  `json:"-"`
}

// StatusResponse from GET /status
type StatusResponse struct {
	Formation struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"formation"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
}

// ConfigResponse from GET /config
type ConfigResponse struct {
	FormationID   string `json:"formation_id"`
	Version       string `json:"version"`
	Description   string `json:"description"`
	SchemaVersion string `json:"schema_version"`
	RequestID     string `json:"-"`
	Timestamp     int64  `json:"-"`
}

// FormationInfoResponse from GET /formation
type FormationInfoResponse struct {
	FormationID string `json:"formation_id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	RequestID   string `json:"-"`
	Timestamp   int64  `json:"-"`
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
	Agents    []Agent `json:"agents"`
	Count     int     `json:"count"`
	RequestID string  `json:"-"`
	Timestamp int64   `json:"-"`
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
	Timestamp int64 `json:"-"`
}

// ChatChunk represents streaming chat chunk
type ChatChunk struct {
	Type         string          `json:"type"` // text, tool_call, tool_result, agent_handoff, thinking, error, done
	Text         string          `json:"text,omitempty"`
	ToolCall     json.RawMessage `json:"tool_call,omitempty"`
	ToolResult   json.RawMessage `json:"tool_result,omitempty"`
	AgentHandoff json.RawMessage `json:"agent_handoff,omitempty"`
	Thinking     string          `json:"thinking,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// Secrets
type SecretsListResponse struct {
	Secrets   map[string]string `json:"secrets"`
	Count     int               `json:"count"`
	RequestID string            `json:"-"`
	Timestamp int64             `json:"-"`
}

type SecretResponse struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
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
	Servers   []MCPServer `json:"servers"`
	Count     int         `json:"count"`
	RequestID string      `json:"-"`
	Timestamp int64       `json:"-"`
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Server      string                 `json:"server"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

type MCPToolsResponse struct {
	Tools     []MCPTool `json:"tools"`
	Count     int       `json:"count"`
	RequestID string    `json:"-"`
	Timestamp int64     `json:"-"`
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
	Sessions  []Session `json:"sessions"`
	Count     int       `json:"count"`
	HasMore   bool      `json:"has_more"`
	RequestID string    `json:"-"`
	Timestamp int64     `json:"-"`
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
	RequestID string    `json:"-"`
	Timestamp int64     `json:"-"`
}

// Requests
type RequestItem struct {
	RequestID string     `json:"request_id"`
	Status    string     `json:"status"`
	Progress  int        `json:"progress,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type RequestsListResponse struct {
	Requests  []RequestItem `json:"requests"`
	Count     int           `json:"count"`
	RequestID string        `json:"-"`
	Timestamp int64         `json:"-"`
}

type RequestStatusResponse struct {
	RequestID   string     `json:"request_id"`
	Status      string     `json:"status"`
	Progress    string     `json:"progress,omitempty"`
	Error       string     `json:"error,omitempty"`
	Completed   *time.Time `json:"completed_at,omitempty"`
	RequestMeta string     `json:"-"`
	Timestamp   int64      `json:"-"`
}

// Memory
type MemoryContent struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

type Memory struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id,omitempty"`
	Content   MemoryContent `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
	RequestID string        `json:"-"`
	Timestamp int64         `json:"-"`
}

type MemoriesListResponse struct {
	Memories  []Memory `json:"memories"`
	Count     int      `json:"count"`
	RequestID string   `json:"-"`
	Timestamp int64    `json:"-"`
}

type MemoryConfigResponse struct {
	Buffer struct {
		Size         int     `json:"size"`
		Multiplier   float64 `json:"multiplier"`
		VectorSearch bool    `json:"vector_search"`
	} `json:"buffer"`
	Working struct {
		MaxMemoryMB     int `json:"max_memory_mb"`
		FIFOIntervalMin int `json:"fifo_interval_min"`
	} `json:"working"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
}

type BufferSession struct {
	SessionID    string    `json:"session_id"`
	MessageCount int       `json:"message_count"`
	LastActivity time.Time `json:"last_activity"`
}

type UserBufferResponse struct {
	UserID        string          `json:"user_id"`
	TotalMessages int             `json:"total_messages"`
	Sessions      []BufferSession `json:"sessions"`
	BufferSizeKB  float64         `json:"buffer_size_kb"`
	RequestID     string          `json:"-"`
	Timestamp     int64           `json:"-"`
}

type BufferStatsResponse struct {
	TotalEntries  int     `json:"total_entries"`
	TotalUsers    int     `json:"total_users"`
	TotalSessions int     `json:"total_sessions"`
	BufferSizeKB  float64 `json:"buffer_size_kb"`
	MaxSize       int     `json:"max_size"`
	Utilization   float64 `json:"utilization"`
	RequestID     string  `json:"-"`
	Timestamp     int64   `json:"-"`
}

type BufferClearedResponse struct {
	Message         string `json:"message"`
	UserID          string `json:"user_id,omitempty"`
	MessagesCleared int    `json:"messages_cleared"`
	SessionsCleared int    `json:"sessions_cleared"`
	RequestID       string `json:"-"`
	Timestamp       int64  `json:"-"`
}

type SessionBufferClearedResponse struct {
	Message         string `json:"message"`
	UserID          string `json:"user_id,omitempty"`
	SessionID       string `json:"session_id"`
	MessagesCleared int    `json:"messages_cleared"`
	RequestID       string `json:"-"`
	Timestamp       int64  `json:"-"`
}

// Async / A2A / Logging
type AsyncSettingsResponse struct {
	ThresholdSeconds int    `json:"threshold_seconds"`
	EnableEstimation bool   `json:"enable_estimation"`
	WebhookURL       string `json:"webhook_url,omitempty"`
	WebhookRetries   int    `json:"webhook_retries"`
	WebhookTimeout   int    `json:"webhook_timeout"`
	RequestID        string `json:"-"`
	Timestamp        int64  `json:"-"`
}

type A2AConfigResponse struct {
	Inbound struct {
		Enabled bool `json:"enabled"`
	} `json:"inbound"`
	Outbound struct {
		Enabled               bool     `json:"enabled"`
		DefaultRetryAttempts  int      `json:"default_retry_attempts"`
		DefaultTimeoutSeconds int      `json:"default_timeout_seconds"`
		AllowedFormations     []string `json:"allowed_formations,omitempty"`
	} `json:"outbound"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
}

type LoggingConfigResponse struct {
	System       map[string]interface{} `json:"system"`
	Conversation map[string]interface{} `json:"conversation"`
	RequestID    string                 `json:"-"`
	Timestamp    int64                  `json:"-"`
}

type LoggingDestination struct {
	ID          string `json:"id"`
	Transport   string `json:"transport"`
	Destination string `json:"destination,omitempty"`
	Level       string `json:"level"`
	Format      string `json:"format"`
	Enabled     bool   `json:"enabled"`
}

type LoggingConversationDestinations struct {
	Destinations []LoggingDestination `json:"destinations"`
	Count        int                  `json:"count"`
}

type LoggingDestinationsResponse struct {
	System       LoggingSystemConfig             `json:"system"`
	Conversation LoggingConversationDestinations `json:"conversation"`
	RequestID    string                          `json:"-"`
	Timestamp    int64                           `json:"-"`
}

type LoggingSystemConfig struct {
	Level       string `json:"level"`
	Destination string `json:"destination"`
}

// Credential services
type CredentialService struct {
	Service     string `json:"service"`
	ServerID    string `json:"server_id"`
	Description string `json:"description"`
}

type CredentialServicesResponse struct {
	Services  []CredentialService `json:"services"`
	Count     int                 `json:"count"`
	RequestID string              `json:"-"`
	Timestamp int64               `json:"-"`
}

// Scheduler
type SchedulerConfigResponse struct {
	// fields per API; keep generic map for forward compatibility
	Config    map[string]interface{} `json:"config"`
	RequestID string                 `json:"-"`
	Timestamp int64                  `json:"-"`
}

type SchedulerJob struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Schedule     string    `json:"schedule,omitempty"`
	RunAt        string    `json:"run_at,omitempty"`
	Message      string    `json:"message"`
	UserID       string    `json:"user_id"`
	Enabled      bool      `json:"enabled"`
	NextRun      time.Time `json:"next_run,omitempty"`
	LastRun      time.Time `json:"last_run,omitempty"`
	FailureCount int       `json:"failure_count"`
}

type SchedulerJobsResponse struct {
	Jobs      []SchedulerJob `json:"jobs"`
	Count     int            `json:"count"`
	RequestID string         `json:"-"`
	Timestamp int64          `json:"-"`
}

type SchedulerJobDetail = SchedulerJob

// User identifiers
type UserIdentifiersResponse struct {
	Identifiers []interface{} `json:"identifiers"`
	Count       int           `json:"count"`
	RequestID   string        `json:"-"`
	Timestamp   int64         `json:"-"`
}

// Overlord / LLM
type OverlordConfigResponse struct {
	Persona       string                 `json:"persona,omitempty"`
	SystemNote    string                 `json:"system_note,omitempty"`
	Clarification map[string]interface{} `json:"clarification,omitempty"`
	Workflow      map[string]interface{} `json:"workflow,omitempty"`
	Response      map[string]interface{} `json:"response,omitempty"`
	LLM           map[string]interface{} `json:"llm,omitempty"`
	Caching       map[string]interface{} `json:"caching,omitempty"`
	RequestID     string                 `json:"-"`
	Timestamp     int64                  `json:"-"`
}

type OverlordSoulResponse struct {
	Soul      string `json:"soul"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
}

type LLMSettingsResponse struct {
	APIKeys   map[string]string        `json:"api_keys,omitempty"`
	Models    []map[string]interface{} `json:"models,omitempty"`
	Settings  map[string]interface{}   `json:"settings,omitempty"`
	RequestID string                   `json:"-"`
	Timestamp int64                    `json:"-"`
}

// Sessions
type SessionDetailResponse struct {
	SessionID    string                 `json:"session_id"`
	UserID       string                 `json:"user_id"`
	CreatedAt    *time.Time             `json:"created_at,omitempty"`
	LastActivity *time.Time             `json:"last_activity,omitempty"`
	MessageCount int                    `json:"message_count,omitempty"`
	Active       bool                   `json:"active,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	RequestID    string                 `json:"-"`
	Timestamp    int64                  `json:"-"`
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
	RequestID      string `json:"-"`
	Timestamp      int64  `json:"-"`
}

// Triggers/SOPs
type TriggersListResponse struct {
	Triggers  []string `json:"triggers"`
	Count     int      `json:"count"`
	RequestID string   `json:"-"`
	Timestamp int64    `json:"-"`
}

type TriggerDetail struct {
	Name       string   `json:"name"`
	Content    string   `json:"content"`
	DataFields []string `json:"data_fields,omitempty"`
	RequestID  string   `json:"-"`
	Timestamp  int64    `json:"-"`
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
	Timestamp int64  `json:"-"`
}

type SOP struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Steps       int    `json:"steps,omitempty"`
	Content     string `json:"content,omitempty"`
	RequestID   string `json:"-"`
	Timestamp   int64  `json:"-"`
}

type SOPsListResponse struct {
	SOPs      []SOP  `json:"sops"`
	Count     int    `json:"count"`
	RequestID string `json:"-"`
	Timestamp int64  `json:"-"`
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
	RequestID    string       `json:"-"`
	Timestamp    int64        `json:"-"`
}

// Credentials
type Credential struct {
	CredentialID      string    `json:"credential_id"`
	Service           string    `json:"service"`
	Name              string    `json:"name"`
	CredentialPreview string    `json:"credential_preview"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
	RequestID         string    `json:"-"`
	Timestamp         int64     `json:"-"`
}

type CredentialsListResponse struct {
	Credentials []Credential `json:"credentials"`
	Count       int          `json:"count"`
	RequestID   string       `json:"-"`
	Timestamp   int64        `json:"-"`
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
	RequestID         string    `json:"-"`
	Timestamp         int64     `json:"-"`
}

type DeleteCredentialResponse struct {
	CredentialID string `json:"credential_id"`
	Deleted      bool   `json:"deleted"`
	RequestID    string `json:"-"`
	Timestamp    int64  `json:"-"`
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
