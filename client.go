package nendesktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://desktop.api.getnen.ai"
	defaultTimeout = 30 * time.Second
	executeTimeout = 120 * time.Second
)

// Config configures a Client. APIKey is required; all other fields have
// sensible defaults.
type Config struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// Client is an HTTP client for the Nen Desktop API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a Client from the given Config.
func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// -- Desktops --

// CreateDesktop calls POST /desktops and returns the created desktop.
func (c *Client) CreateDesktop(ctx context.Context, desktopType string) (*Desktop, error) {
	body, err := json.Marshal(createDesktopRequest{DesktopType: desktopType})
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/desktops", bytes.NewReader(body), 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var desktop Desktop
	if err := json.NewDecoder(resp.Body).Decode(&desktop); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &desktop, nil
}

// ListDesktops calls GET /desktops and returns all active desktops.
func (c *Client) ListDesktops(ctx context.Context) ([]Desktop, error) {
	resp, err := c.do(ctx, http.MethodGet, "/desktops", nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var desktops []Desktop
	if err := json.NewDecoder(resp.Body).Decode(&desktops); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return desktops, nil
}

// GetDesktop calls GET /desktops/:id and returns a single desktop.
func (c *Client) GetDesktop(ctx context.Context, desktopID string) (*Desktop, error) {
	resp, err := c.do(ctx, http.MethodGet, "/desktops/"+desktopID, nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var desktop Desktop
	if err := json.NewDecoder(resp.Body).Decode(&desktop); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &desktop, nil
}

// UpdateDesktop calls PATCH /desktops/:id to update the desktop name.
func (c *Client) UpdateDesktop(ctx context.Context, desktopID, name string) (*Desktop, error) {
	body, err := json.Marshal(updateDesktopRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPatch, "/desktops/"+desktopID, bytes.NewReader(body), 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var desktop Desktop
	if err := json.NewDecoder(resp.Body).Decode(&desktop); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &desktop, nil
}

// DeleteDesktop calls DELETE /desktops/:id and destroys the desktop.
func (c *Client) DeleteDesktop(ctx context.Context, desktopID string) (*DeleteResponse, error) {
	resp, err := c.do(ctx, http.MethodDelete, "/desktops/"+desktopID, nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var dr DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &dr, nil
}

// -- Execute / Tools --

// Execute calls POST /desktops/:id/execute with the given tool action.
func (c *Client) Execute(ctx context.Context, desktopID, tool, action string, params map[string]any) (*ExecuteResult, error) {
	if params == nil {
		params = map[string]any{}
	}
	body, err := json.Marshal(executeRequest{
		Action: executeAction{Tool: tool, Action: action, Params: params},
	})
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/desktops/"+desktopID+"/execute", bytes.NewReader(body), executeTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result ExecuteResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// ListTools calls GET /desktops/:id/tools and returns the tool schemas.
func (c *Client) ListTools(ctx context.Context, desktopID string) ([]ToolSchema, error) {
	resp, err := c.do(ctx, http.MethodGet, "/desktops/"+desktopID+"/tools", nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var tools []ToolSchema
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("decoding tools: %w", err)
	}
	return tools, nil
}

// GetToolLogs calls GET /desktops/:id/tool-logs and returns the raw JSON response.
func (c *Client) GetToolLogs(ctx context.Context, desktopID string) (json.RawMessage, error) {
	resp, err := c.do(ctx, http.MethodGet, "/desktops/"+desktopID+"/tool-logs", nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return json.RawMessage(data), nil
}

// -- Sessions --

// CreateSession calls PUT /desktops/:id/session to create or reconnect an RDP session.
func (c *Client) CreateSession(ctx context.Context, desktopID string) (*SessionInfo, error) {
	resp, err := c.do(ctx, http.MethodPut, "/desktops/"+desktopID+"/session", nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var session SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &session, nil
}

// GetSession calls GET /desktops/:id/session and returns the session status.
func (c *Client) GetSession(ctx context.Context, desktopID string) (*SessionInfo, error) {
	resp, err := c.do(ctx, http.MethodGet, "/desktops/"+desktopID+"/session", nil, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var session SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &session, nil
}

// DeleteSession calls DELETE /desktops/:id/session to disconnect the session.
func (c *Client) DeleteSession(ctx context.Context, desktopID string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/desktops/"+desktopID+"/session", nil, 0)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return err
	}
	return nil
}

// -- Internal --

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, timeout time.Duration) (*http.Response, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		_ = cancel // caller closes resp.Body which unblocks the request
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       strings.TrimSpace(string(body)),
	}
}
