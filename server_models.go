package muxi

import "encoding/json"

// ServerAPIResponse is the server envelope
type ServerAPIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
	Code    int             `json:"code,omitempty"`
}

// ServerHealthResponse from GET /health
type ServerHealthResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Status     string `json:"status"`
		Formations int    `json:"formations"`
		PortPool   struct {
			Allocated int `json:"allocated"`
			Available int `json:"available"`
			Total     int `json:"total"`
		} `json:"port_pool"`
	} `json:"data"`
}

// ServerStatusResponse from GET /rpc/server/status
type ServerStatusResponse struct {
	Server struct {
		ServerID string `json:"server_id"`
		Version  string `json:"version"`
		Uptime   int64  `json:"uptime"`
		Port     int    `json:"port"`
	} `json:"server"`
	Formations struct {
		Total   int `json:"total"`
		Running int `json:"running"`
		Stopped int `json:"stopped"`
		Healthy int `json:"healthy"`
	} `json:"formations"`
	Ports struct {
		Allocated int    `json:"allocated"`
		Available int    `json:"available"`
		Range     string `json:"range"`
	} `json:"ports"`
	Runtime struct {
		Type     string   `json:"type"`
		Platform string   `json:"platform"`
		Versions []string `json:"versions"`
	} `json:"runtime"`
}

// FormationVersion contains version information
type FormationVersion struct {
	Semantic string `json:"semantic"`
	Current  string `json:"current"`
	Previous string `json:"previous"`
}

// FormationListItem is a formation in the list
type FormationListItem struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      *FormationVersion `json:"version,omitempty"`
	Status       string            `json:"status"`
	Port         int               `json:"port"`
	PID          int               `json:"pid"`
	Uptime       int64             `json:"uptime"`
	RestartCount int               `json:"restart_count"`
	Healthy      bool              `json:"healthy"`
}

// FormationDetail from GET /rpc/formations/{id}
type FormationDetail struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      *FormationVersion `json:"version,omitempty"`
	Status       string            `json:"status"`
	Port         int               `json:"port"`
	PID          int               `json:"pid"`
	Uptime       int64             `json:"uptime"`
	RestartCount int               `json:"restart_count"`
	CreatedAt    string            `json:"created_at"`
	DeployedAt   string            `json:"deployed_at"`
	UpdatedAt    string            `json:"updated_at"`
}

// ListFormationsResponse from GET /rpc/formations
type ListFormationsResponse struct {
	Formations []FormationListItem `json:"formations"`
	Total      int                 `json:"total"`
}

// DeployRequest for deployment
type DeployRequest struct {
	FormationID string
	BundlePath  string
	Version     string
}

// DeployResponse from deploy/update
type DeployResponse struct {
	ID      string `json:"id"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// RollbackResponse from rollback
type RollbackResponse struct {
	ID              string `json:"id"`
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
}

// LogsResponse from GET /rpc/formations/{id}/logs
type LogsResponse struct {
	FormationID string `json:"formation_id"`
	Logs        struct {
		Stdout []string `json:"stdout"`
		Stderr []string `json:"stderr"`
	} `json:"logs"`
}

// LogEvent represents a single log line in streaming mode
type LogEvent struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
	Time   string `json:"time,omitempty"`
}

// Deploy streaming events
type DeployProgressEvent struct {
	Stage       string `json:"stage"`
	Message     string `json:"message"`
	Progress    int    `json:"progress"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	Attempt     int    `json:"attempt"`
	MaxAttempts int    `json:"max_attempts"`
	StagingPort int    `json:"staging_port"`
}

type DeployCompleteEvent struct {
	FormationID     string `json:"formation_id"`
	Port            int    `json:"port"`
	Status          string `json:"status"`
	URL             string `json:"url"`
	HealthURL       string `json:"health_url"`
	PID             int    `json:"pid"`
	PreviousVersion string `json:"previous_version,omitempty"`
	NewVersion      string `json:"new_version,omitempty"`
}

type DeployErrorEvent struct {
	Error          string `json:"error"`
	Message        string `json:"message"`
	Stage          string `json:"stage"`
	RollbackStatus string `json:"rollback_status,omitempty"`
}

// DeployEvent is a typed union for streaming
type DeployEvent struct {
	Progress *DeployProgressEvent
	Complete *DeployCompleteEvent
	Error    *DeployErrorEvent
}
