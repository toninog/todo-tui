package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/blackraven/todo-tui/internal/config"
)

// Credentials stores login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Client handles API communication
type Client struct {
	baseURL    string
	tokenPath  string
	credsPath  string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	c := &Client{
		baseURL:   cfg.APIURL,
		tokenPath: cfg.TokenPath,
		credsPath: cfg.CredsPath,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	c.loadToken()
	return c
}

// loadToken loads the JWT token from disk
func (c *Client) loadToken() {
	data, err := os.ReadFile(c.tokenPath)
	if err == nil {
		c.token = strings.TrimSpace(string(data))
	}
}

// saveToken saves the JWT token to disk
func (c *Client) saveToken(token string) error {
	c.token = token
	if err := config.EnsureDataDir(); err != nil {
		return err
	}
	return os.WriteFile(c.tokenPath, []byte(token), 0600)
}

// clearToken removes the saved token
func (c *Client) clearToken() error {
	c.token = ""
	return os.Remove(c.tokenPath)
}

// HasToken returns true if a token is loaded
func (c *Client) HasToken() bool {
	return c.token != ""
}

// SetToken sets the current token
func (c *Client) SetToken(token string) error {
	return c.saveToken(token)
}

// ClearToken clears the current token
func (c *Client) ClearToken() error {
	return c.clearToken()
}

// SaveCredentials stores login credentials for auto-login
func (c *Client) SaveCredentials(email, password string) error {
	if err := config.EnsureDataDir(); err != nil {
		return err
	}
	creds := Credentials{Email: email, Password: password}
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(c.credsPath, data, 0600)
}

// LoadCredentials loads stored credentials
func (c *Client) LoadCredentials() (*Credentials, error) {
	data, err := os.ReadFile(c.credsPath)
	if err != nil {
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// ClearCredentials removes stored credentials
func (c *Client) ClearCredentials() error {
	os.Remove(c.credsPath)
	return nil
}

// HasCredentials returns true if credentials are stored
func (c *Client) HasCredentials() bool {
	_, err := os.Stat(c.credsPath)
	return err == nil
}

// AutoLogin attempts to login with stored credentials
func (c *Client) AutoLogin() error {
	creds, err := c.LoadCredentials()
	if err != nil {
		return err
	}
	return c.Login(creds.Email, creds.Password)
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsUnauthorized returns true if the error is an auth error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// request makes an HTTP request to the API
func (c *Client) request(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Detail != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: errResp.Detail}
		}
		return &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Get makes a GET request
func (c *Client) Get(path string, result interface{}) error {
	return c.request("GET", path, nil, result)
}

// Post makes a POST request
func (c *Client) Post(path string, body, result interface{}) error {
	return c.request("POST", path, body, result)
}

// Patch makes a PATCH request
func (c *Client) Patch(path string, body, result interface{}) error {
	return c.request("PATCH", path, body, result)
}

// Delete makes a DELETE request
func (c *Client) Delete(path string, result interface{}) error {
	return c.request("DELETE", path, nil, result)
}
