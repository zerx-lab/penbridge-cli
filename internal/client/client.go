package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// ErrDryRun is returned by request methods when DryRun is enabled. It signals
// that the request was printed instead of sent and should be treated as a
// successful no-op by callers.
var ErrDryRun = errors.New("dry-run: request not sent")

// Client talks to the SiYuan kernel API.
type Client struct {
	cfg        Config
	httpClient *http.Client

	// DryRun prints the planned request (to Out) and does not send it.
	DryRun bool
	// Verbose logs request/response metadata to Log.
	Verbose bool
	// Out receives dry-run payloads (defaults to os.Stdout).
	Out io.Writer
	// Log receives verbose logs (defaults to os.Stderr).
	Log io.Writer
}

// New creates a Client from cfg.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			Transport: transport,
		},
		Out: os.Stdout,
		Log: os.Stderr,
	}
}

// Config returns the effective configuration.
func (c *Client) Config() Config { return c.cfg }

// APIResponse is the standard SiYuan response envelope.
type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

// APIError represents a non-zero API response code or an HTTP error.
type APIError struct {
	Path       string
	Code       int
	Msg        string
	HTTPStatus int
}

func (e *APIError) Error() string {
	msg := e.Msg
	if msg == "" {
		msg = "(no message)"
	}
	return fmt.Sprintf("API %s failed: %s [code=%d http=%d]", e.Path, msg, e.Code, e.HTTPStatus)
}

// NormalizePath converts a user-supplied endpoint into an absolute API path.
//
//	insertBlock          -> (left as-is, caller should pass a full path)
//	block/insertBlock    -> /api/block/insertBlock
//	api/block/insert     -> /api/block/insert
//	/api/block/insert    -> /api/block/insert
func NormalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "/") {
		return p
	}
	if strings.HasPrefix(p, "api/") {
		return "/" + p
	}
	return "/api/" + p
}

func (c *Client) url(path string) string {
	return strings.TrimRight(c.cfg.BaseURL, "/") + path
}

func (c *Client) setAuth(req *http.Request) {
	switch {
	case c.cfg.Token != "":
		req.Header.Set("Authorization", "Token "+c.cfg.Token)
	case c.cfg.User != "" || c.cfg.Password != "":
		req.SetBasicAuth(c.cfg.User, c.cfg.Password)
	}
}

func (c *Client) authPreview() string {
	switch {
	case c.cfg.Token != "":
		return "Token " + mask(c.cfg.Token)
	case c.cfg.User != "" || c.cfg.Password != "":
		return "Basic " + c.cfg.User + ":****"
	default:
		return "(none)"
	}
}

// Call POSTs a JSON body to the given API path and parses the response envelope.
// body may be nil, []byte, json.RawMessage, or any JSON-marshalable value.
// On a non-zero response code it returns the parsed response and an *APIError.
func (c *Client) Call(ctx context.Context, path string, body any) (*APIResponse, error) {
	bodyBytes, err := toJSONBytes(body)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		bodyBytes = []byte("{}")
	}
	return c.doJSON(ctx, http.MethodPost, NormalizePath(path), bodyBytes)
}

// CallMethod is like Call but with an explicit HTTP method.
func (c *Client) CallMethod(ctx context.Context, method, path string, body []byte) (*APIResponse, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		body = []byte("{}")
	}
	return c.doJSON(ctx, strings.ToUpper(method), NormalizePath(path), body)
}

func (c *Client) doJSON(ctx context.Context, method, path string, bodyBytes []byte) (*APIResponse, error) {
	u := c.url(path)
	if c.DryRun {
		c.emitDryRun(method, u, "application/json", bodyBytes, nil)
		return nil, ErrDryRun
	}
	if c.Verbose {
		c.logRequest(method, u, "application/json", bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response %s: %w", path, err)
	}
	if c.Verbose {
		fmt.Fprintf(c.Log, "< HTTP %d (%d bytes)\n", resp.StatusCode, len(raw))
	}
	return parseEnvelope(path, resp.StatusCode, raw)
}

func parseEnvelope(path string, status int, raw []byte) (*APIResponse, error) {
	var apiResp APIResponse
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &apiResp); err != nil {
			return nil, fmt.Errorf("API %s: HTTP %d, unexpected non-JSON response: %s",
				path, status, truncate(raw, 512))
		}
	}
	if apiResp.Code != 0 {
		return &apiResp, &APIError{Path: path, Code: apiResp.Code, Msg: apiResp.Msg, HTTPStatus: status}
	}
	if status >= 400 {
		msg := apiResp.Msg
		if msg == "" {
			msg = strings.TrimSpace(string(truncate(raw, 256)))
		}
		return &apiResp, &APIError{Path: path, Code: apiResp.Code, Msg: msg, HTTPStatus: status}
	}
	return &apiResp, nil
}

// GetFile fetches a workspace file via /api/file/getFile. On success it returns
// the raw bytes and content type. API errors are returned as *APIError.
func (c *Client) GetFile(ctx context.Context, workspacePath string) ([]byte, string, error) {
	const path = "/api/file/getFile"
	bodyBytes, _ := json.Marshal(map[string]string{"path": workspacePath})
	u := c.url(path)
	if c.DryRun {
		c.emitDryRun(http.MethodPost, u, "application/json", bodyBytes, nil)
		return nil, "", ErrDryRun
	}
	if c.Verbose {
		c.logRequest(http.MethodPost, u, "application/json", bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	ct := resp.Header.Get("Content-Type")
	// getFile returns the raw bytes with HTTP 200 on success; on failure it
	// returns a JSON envelope (often HTTP 202) with a non-zero code.
	if resp.StatusCode == http.StatusOK && !strings.Contains(ct, "application/json") {
		return raw, ct, nil
	}
	// Try to parse an error envelope; if it isn't JSON, treat 200 as the file.
	var apiResp APIResponse
	if json.Unmarshal(raw, &apiResp) == nil && (apiResp.Code != 0 || resp.StatusCode != http.StatusOK) {
		msg := apiResp.Msg
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, "", &APIError{Path: path, Code: apiResp.Code, Msg: msg, HTTPStatus: resp.StatusCode}
	}
	return raw, ct, nil
}

// PutFile uploads data to a workspace path via /api/file/putFile (multipart).
// If isDir is true, a directory is created and fileData is ignored.
func (c *Client) PutFile(ctx context.Context, workspacePath string, isDir bool, fileName string, fileData io.Reader) (*APIResponse, error) {
	const path = "/api/file/putFile"
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("path", workspacePath)
	_ = w.WriteField("isDir", fmt.Sprintf("%t", isDir))
	_ = w.WriteField("modTime", fmt.Sprintf("%d", time.Now().UnixMilli()))
	if !isDir {
		if fileName == "" {
			fileName = "file"
		}
		fw, err := w.CreateFormFile("file", fileName)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(fw, fileData); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	u := c.url(path)
	if c.DryRun {
		c.emitDryRun(http.MethodPost, u, w.FormDataContentType(),
			[]byte(fmt.Sprintf("(multipart) path=%s isDir=%t file=%s", workspacePath, isDir, fileName)), nil)
		return nil, ErrDryRun
	}
	if c.Verbose {
		fmt.Fprintf(c.Log, "> POST %s (multipart upload path=%s isDir=%t)\n", u, workspacePath, isDir)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	c.setAuth(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return parseEnvelope(path, resp.StatusCode, raw)
}

func (c *Client) logRequest(method, u, contentType string, body []byte) {
	fmt.Fprintf(c.Log, "> %s %s\n", method, u)
	fmt.Fprintf(c.Log, "> Authorization: %s\n", c.authPreview())
	fmt.Fprintf(c.Log, "> Content-Type: %s\n", contentType)
	if len(body) > 0 {
		fmt.Fprintf(c.Log, "> %s\n", truncate(body, 4096))
	}
}

// emitDryRun writes a machine-readable description of the request to Out.
func (c *Client) emitDryRun(method, u, contentType string, body []byte, _ any) {
	preview := map[string]any{
		"method":      method,
		"url":         u,
		"contentType": contentType,
		"authorization": c.authPreview(),
	}
	if json.Valid(body) {
		preview["body"] = json.RawMessage(body)
	} else {
		preview["body"] = string(body)
	}
	out, _ := json.MarshalIndent(preview, "", "  ")
	fmt.Fprintln(c.Out, string(out))
}

func toJSONBytes(body any) ([]byte, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case []byte:
		return b, nil
	case json.RawMessage:
		return b, nil
	case string:
		return []byte(b), nil
	default:
		out, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		return out, nil
	}
}

func truncate(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	out := make([]byte, 0, n+3)
	out = append(out, b[:n]...)
	out = append(out, "..."...)
	return out
}
