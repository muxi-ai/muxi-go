package muxi

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "os"
    "time"
)

// ServerConfig configures ServerClient
type ServerConfig struct {
    URL        string
    KeyID      string
    SecretKey  string
    MaxRetries int           // 0 = no retries (default)
    Timeout    time.Duration // default 30s
    HTTPClient *http.Client  // optional custom client
}

// ServerClient is an HTTP client for MUXI Server (management API)
type ServerClient struct {
    baseURL    string
    keyID      string
    secretKey  string
    httpClient *http.Client
    maxRetries int
}

// NewServerClient constructs a ServerClient
func NewServerClient(cfg *ServerConfig) *ServerClient {
    if cfg == nil {
        panic("ServerConfig is required")
    }

    timeout := cfg.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second
    }

    transport := newSDKTransport(http.DefaultTransport)
    client := cfg.HTTPClient
    if client == nil {
        client = &http.Client{Timeout: timeout, Transport: transport}
    } else {
        // wrap provided transport to inject headers
        base := client.Transport
        if base == nil {
            base = http.DefaultTransport
        }
        client = &http.Client{
            Timeout:   client.Timeout,
            Transport: newSDKTransport(base),
        }
    }

    return &ServerClient{
        baseURL:    cfg.URL,
        keyID:      cfg.KeyID,
        secretKey:  cfg.SecretKey,
        httpClient: client,
        maxRetries: cfg.MaxRetries,
    }
}

// Ping tests connectivity (unauthenticated)
func (c *ServerClient) Ping(ctx context.Context) (int64, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ping", nil)
    if err != nil {
        return 0, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return 0, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return 0, mapStatusToError(resp.StatusCode, resp.Body)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }

    return int64(len(body)), nil
}

// Health checks server health (unauthenticated)
func (c *ServerClient) Health(ctx context.Context) (*ServerHealthResponse, error) {
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

    var result ServerHealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// Status returns server status (authenticated)
func (c *ServerClient) Status(ctx context.Context) (*ServerStatusResponse, error) {
    resp, err := c.do(ctx, http.MethodGet, "/rpc/server/status", nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    apiResp, err := decodeServerAPI(resp)
    if err != nil {
        return nil, err
    }

    var status ServerStatusResponse
    if err := json.Unmarshal(apiResp.Data, &status); err != nil {
        return nil, err
    }
    return &status, nil
}

// ListFormations lists formations
func (c *ServerClient) ListFormations(ctx context.Context) (*ListFormationsResponse, error) {
    resp, err := c.do(ctx, http.MethodGet, "/rpc/formations", nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    apiResp, err := decodeServerAPI(resp)
    if err != nil {
        return nil, err
    }

    var list ListFormationsResponse
    if err := json.Unmarshal(apiResp.Data, &list); err != nil {
        return nil, err
    }
    return &list, nil
}

// GetFormation returns formation details
func (c *ServerClient) GetFormation(ctx context.Context, id string) (*FormationDetail, error) {
    resp, err := c.do(ctx, http.MethodGet, "/rpc/formations/"+id, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    apiResp, err := decodeServerAPI(resp)
    if err != nil {
        return nil, err
    }

    var detail FormationDetail
    if err := json.Unmarshal(apiResp.Data, &detail); err != nil {
        return nil, err
    }
    return &detail, nil
}

// StopFormation stops a formation
func (c *ServerClient) StopFormation(ctx context.Context, id string) error {
    resp, err := c.do(ctx, http.MethodPost, "/rpc/formations/"+id+"/stop", nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return checkServerResponse(resp)
}

// DeleteFormation deletes a formation
func (c *ServerClient) DeleteFormation(ctx context.Context, id string) error {
    resp, err := c.do(ctx, http.MethodDelete, "/rpc/formations/"+id, nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return checkServerResponse(resp)
}

// DeployFormation deploys a formation (non-streaming)
func (c *ServerClient) DeployFormation(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
    if req == nil {
        return nil, fmt.Errorf("DeployRequest is required")
    }

    file, err := os.Open(req.BundlePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open bundle: %w", err)
    }
    defer file.Close()

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/rpc/formations", file)
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/gzip")
    httpReq.Header.Set("X-Formation-ID", req.FormationID)
    if req.Version != "" {
        httpReq.Header.Set("X-Formation-Version", req.Version)
    }
    httpReq.Header.Set("Authorization", BuildAuthHeader(c.keyID, c.secretKey, http.MethodPost, "/rpc/formations"))

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
    }
    defer resp.Body.Close()

    if err := checkServerResponse(resp); err != nil {
        return nil, err
    }

    var apiResp ServerAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }

    var deploy DeployResponse
    if err := json.Unmarshal(apiResp.Data, &deploy); err != nil {
        return nil, err
    }
    return &deploy, nil
}

// DeployFormationStreaming deploys with SSE progress
func (c *ServerClient) DeployFormationStreaming(ctx context.Context, req *DeployRequest) (<-chan DeployEvent, <-chan error) {
    if req == nil {
        errCh := make(chan error, 1)
        evtCh := make(chan DeployEvent)
        errCh <- fmt.Errorf("DeployRequest is required")
        close(errCh)
        close(evtCh)
        return evtCh, errCh
    }

    headers := map[string]string{
        "Content-Type":      "application/gzip",
        "X-Formation-ID":    req.FormationID,
        "Authorization":     BuildAuthHeader(c.keyID, c.secretKey, http.MethodPost, "/rpc/formations"),
    }
    if req.Version != "" {
        headers["X-Formation-Version"] = req.Version
    }
    openBody := func() (io.ReadCloser, error) {
        return os.Open(req.BundlePath)
    }
    return c.streamDeploy(ctx, http.MethodPost, c.baseURL+"/rpc/formations", openBody, headers)
}

// UpdateFormation updates an existing formation (non-streaming)
func (c *ServerClient) UpdateFormation(ctx context.Context, id string, req *DeployRequest) (*DeployResponse, error) {
    if req == nil {
        return nil, fmt.Errorf("DeployRequest is required")
    }

    file, err := os.Open(req.BundlePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open bundle: %w", err)
    }
    defer file.Close()

    path := "/rpc/formations/" + id
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, file)
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/gzip")
    if req.Version != "" {
        httpReq.Header.Set("X-Formation-Version", req.Version)
    }
    httpReq.Header.Set("Authorization", BuildAuthHeader(c.keyID, c.secretKey, http.MethodPut, path))

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
    }
    defer resp.Body.Close()

    if err := checkServerResponse(resp); err != nil {
        return nil, err
    }

    var apiResp ServerAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }

    var update DeployResponse
    if err := json.Unmarshal(apiResp.Data, &update); err != nil {
        return nil, err
    }
    return &update, nil
}

// UpdateFormationStreaming updates with SSE progress
func (c *ServerClient) UpdateFormationStreaming(ctx context.Context, id string, req *DeployRequest) (<-chan DeployEvent, <-chan error) {
    if req == nil {
        errCh := make(chan error, 1)
        evtCh := make(chan DeployEvent)
        errCh <- fmt.Errorf("DeployRequest is required")
        close(errCh)
        close(evtCh)
        return evtCh, errCh
    }

    path := "/rpc/formations/" + id
    headers := map[string]string{
        "Content-Type":  "application/gzip",
        "Authorization": BuildAuthHeader(c.keyID, c.secretKey, http.MethodPut, path),
    }
    if req.Version != "" {
        headers["X-Formation-Version"] = req.Version
    }

    openBody := func() (io.ReadCloser, error) {
        return os.Open(req.BundlePath)
    }

    return c.streamDeploy(ctx, http.MethodPut, c.baseURL+path, openBody, headers)
}

// StartFormation starts a formation
func (c *ServerClient) StartFormation(ctx context.Context, id string) error {
    resp, err := c.do(ctx, http.MethodPost, "/rpc/formations/"+id+"/start", nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return checkServerResponse(resp)
}

// StartFormationStreaming starts with SSE progress
func (c *ServerClient) StartFormationStreaming(ctx context.Context, id string) (<-chan DeployEvent, <-chan error) {
    path := "/rpc/formations/" + id + "/start"
    headers := map[string]string{
        "Authorization": BuildAuthHeader(c.keyID, c.secretKey, http.MethodPost, path),
    }
    return c.streamDeploy(ctx, http.MethodPost, c.baseURL+path, nil, headers)
}

// RestartFormation restarts a formation
func (c *ServerClient) RestartFormation(ctx context.Context, id string) error {
    resp, err := c.do(ctx, http.MethodPost, "/rpc/formations/"+id+"/restart", nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return checkServerResponse(resp)
}

// RestartFormationStreaming restarts with SSE progress
func (c *ServerClient) RestartFormationStreaming(ctx context.Context, id string) (<-chan DeployEvent, <-chan error) {
    path := "/rpc/formations/" + id + "/restart"
    headers := map[string]string{
        "Authorization": BuildAuthHeader(c.keyID, c.secretKey, http.MethodPost, path),
    }
    return c.streamDeploy(ctx, http.MethodPost, c.baseURL+path, nil, headers)
}

// RollbackFormation rolls back a formation
func (c *ServerClient) RollbackFormation(ctx context.Context, id string) (*RollbackResponse, error) {
    resp, err := c.do(ctx, http.MethodPost, "/rpc/formations/"+id+"/rollback", nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    apiResp, err := decodeServerAPI(resp)
    if err != nil {
        return nil, err
    }
    var rb RollbackResponse
    if err := json.Unmarshal(apiResp.Data, &rb); err != nil {
        return nil, err
    }
    return &rb, nil
}

// RollbackFormationStreaming rolls back with SSE progress
func (c *ServerClient) RollbackFormationStreaming(ctx context.Context, id string) (<-chan DeployEvent, <-chan error) {
    path := "/rpc/formations/" + id + "/rollback"
    headers := map[string]string{
        "Authorization": BuildAuthHeader(c.keyID, c.secretKey, http.MethodPost, path),
    }
    return c.streamDeploy(ctx, http.MethodPost, c.baseURL+path, nil, headers)
}

// CancelUpdate cancels an ongoing update
func (c *ServerClient) CancelUpdate(ctx context.Context, id string) error {
    resp, err := c.do(ctx, http.MethodPost, "/rpc/formations/"+id+"/cancel-update", nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return checkServerResponse(resp)
}

// GetFormationLogs gets logs for a formation
func (c *ServerClient) GetFormationLogs(ctx context.Context, id string, lines int, stream string) (*LogsResponse, error) {
    path := fmt.Sprintf("/rpc/formations/%s/logs?lines=%d&stream=%s", id, lines, stream)
    resp, err := c.do(ctx, http.MethodGet, path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    apiResp, err := decodeServerAPI(resp)
    if err != nil {
        return nil, err
    }

    var logs LogsResponse
    if err := json.Unmarshal(apiResp.Data, &logs); err != nil {
        return nil, err
    }
    return &logs, nil
}

// StreamFormationLogs streams logs via SSE; returns channel of LogEvent
func (c *ServerClient) StreamFormationLogs(ctx context.Context, id string, stream string) (<-chan LogEvent, <-chan error) {
    path := fmt.Sprintf("/rpc/formations/%s/logs?stream=%s&follow=true", id, stream)
    headers := map[string]string{
        "Accept":       "text/event-stream",
        "Authorization": BuildAuthHeader(c.keyID, c.secretKey, http.MethodGet, fmt.Sprintf("/rpc/formations/%s/logs", id)),
    }
    return c.streamLogs(ctx, c.baseURL+path, headers)
}

// --- internal helpers ---

// do executes an authenticated request with retry (GET/DELETE only)
func (c *ServerClient) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
    url := c.baseURL + path

    attempt := 0
    for {
        req, err := http.NewRequestWithContext(ctx, method, url, body)
        if err != nil {
            return nil, err
        }

        if contentType != "" {
            req.Header.Set("Content-Type", contentType)
        }

        // Add auth
        req.Header.Set("Authorization", BuildAuthHeader(c.keyID, c.secretKey, method, path))

        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
        }

        // If success or non-retryable, return
        if !shouldRetry(method, resp.StatusCode, c.maxRetries, attempt) {
            return resp, nil
        }

        // Read and close body before retry to reuse connection
        io.Copy(io.Discard, resp.Body)
        resp.Body.Close()

        // Sleep with backoff
        delay := backoffDelay(attempt)
        if ra := retryAfter(resp); ra > 0 {
            delay = ra
        }
        time.Sleep(delay)
        attempt++
    }
}

func shouldRetry(method string, status int, maxRetries, attempt int) bool {
    if attempt >= maxRetries {
        return false
    }

    // Only retry idempotent methods
    if method != http.MethodGet && method != http.MethodDelete {
        return false
    }

    switch status {
    case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusInternalServerError,
        http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
        return true
    default:
        return false
    }
}

func backoffDelay(attempt int) time.Duration {
    base := 500 * time.Millisecond
    max := 30 * time.Second
    d := base * (1 << attempt)
    if d > max {
        d = max
    }
    // jitter ±10%
    jitter := 0.1 * rand.Float64()
    return time.Duration(float64(d) * (1 + jitter))
}

func retryAfter(resp *http.Response) time.Duration {
    if ra := resp.Header.Get("Retry-After"); ra != "" {
        if secs, err := time.ParseDuration(ra + "s"); err == nil {
            return secs
        }
    }
    return 0
}

// decodeServerAPI parses the server API envelope
func decodeServerAPI(resp *http.Response) (*ServerAPIResponse, error) {
    if err := checkServerResponse(resp); err != nil {
        return nil, err
    }

    var apiResp ServerAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }

    if !apiResp.Success {
        msg := apiResp.Message
        if msg == "" && apiResp.Error != "" {
            msg = apiResp.Error
        }
        return nil, &ServerError{newMuxiError(ErrServerError, msg, resp.StatusCode)}
    }

    return &apiResp, nil
}

// checkServerResponse maps HTTP status codes to errors
func checkServerResponse(resp *http.Response) error {
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return nil
    }
    return mapStatusToError(resp.StatusCode, resp.Body)
}

func mapStatusToError(status int, body io.Reader) error {
    // Try to parse server envelope
    var apiResp ServerAPIResponse
    data, _ := io.ReadAll(body)
    if err := json.Unmarshal(data, &apiResp); err == nil {
        if apiResp.Error != "" {
            return &MuxiError{Code: apiResp.Error, Message: apiResp.Message, StatusCode: status}
        }
        if apiResp.Message != "" {
            return &MuxiError{Code: ErrServerError, Message: apiResp.Message, StatusCode: status}
        }
    }

    switch status {
    case http.StatusUnauthorized:
        return &AuthenticationError{newMuxiError(ErrUnauthorized, "authentication failed", status)}
    case http.StatusForbidden:
        return &AuthorizationError{newMuxiError(ErrForbidden, "access denied", status)}
    case http.StatusNotFound:
        return &NotFoundError{newMuxiError(ErrNotFound, "not found", status)}
    case http.StatusConflict:
        return &ConflictError{newMuxiError(ErrConflict, "conflict", status)}
    default:
        return &ServerError{newMuxiError(ErrServerError, fmt.Sprintf("server error: %d", status), status)}
    }
}

// streamDeploy executes an SSE deploy/update/start/restart/rollback and parses events
func (c *ServerClient) streamDeploy(ctx context.Context, method, url string, openBody func() (io.ReadCloser, error), headers map[string]string) (<-chan DeployEvent, <-chan error) {
    out := make(chan DeployEvent)
    errs := make(chan error, 1)

    go func() {
        defer close(out)
        defer close(errs)

        var body io.ReadCloser
        if openBody != nil {
            var err error
            body, err = openBody()
            if err != nil {
                errs <- fmt.Errorf("failed to open body: %w", err)
                return
            }
            defer body.Close()
        }

        req, err := http.NewRequestWithContext(ctx, method, url, body)
        if err != nil {
            errs <- err
            return
        }
        req.Header.Set("Accept", "text/event-stream")
        for k, v := range headers {
            req.Header.Set(k, v)
        }

        client := &http.Client{Timeout: 10 * time.Minute, Transport: newSDKTransport(http.DefaultTransport)}
        resp, err := client.Do(req)
        if err != nil {
            errs <- &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 400 {
            errs <- mapStatusToError(resp.StatusCode, resp.Body)
            return
        }

        if err := parseDeploySSE(resp.Body, out); err != nil {
            errs <- err
        }
    }()

    return out, errs
}

// streamLogs streams log events via SSE
func (c *ServerClient) streamLogs(ctx context.Context, url string, headers map[string]string) (<-chan LogEvent, <-chan error) {
    out := make(chan LogEvent)
    errs := make(chan error, 1)

    go func() {
        defer close(out)
        defer close(errs)

        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil {
            errs <- err
            return
        }
        for k, v := range headers {
            req.Header.Set(k, v)
        }

        client := &http.Client{Timeout: 0, Transport: newSDKTransport(http.DefaultTransport)}
        resp, err := client.Do(req)
        if err != nil {
            errs <- &ConnectionError{newMuxiError(ErrConnectionError, err.Error(), 0)}
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 400 {
            errs <- mapStatusToError(resp.StatusCode, resp.Body)
            return
        }

        if err := parseLogSSE(resp.Body, out); err != nil {
            errs <- err
        }
    }()

    return out, errs
}
