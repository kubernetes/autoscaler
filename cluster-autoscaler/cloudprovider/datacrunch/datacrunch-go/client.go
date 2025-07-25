package datacrunchclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client is the main struct for interacting with the DataCrunch API.
type Client struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
	token        *tokenResponse
	tokenMu      sync.Mutex
}

// tokenResponse holds the OAuth2 token and related info.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	ObtainedAt   time.Time
}

// NewClient creates a new DataCrunch API client.
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      "https://api.datacrunch.io/v1",
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Login authenticates with the API and retrieves an access token.
func (c *Client) Login() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	body := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}
	b, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/oauth2/token", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("login failed: %s", resp.Status)
	}
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return err
	}
	tr.ObtainedAt = time.Now()
	c.token = &tr
	return nil
}

// RefreshToken refreshes the access token using the refresh token.
func (c *Client) RefreshToken() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token == nil || c.token.RefreshToken == "" {
		return errors.New("no refresh token available")
	}
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": c.token.RefreshToken,
	}
	b, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/oauth2/token", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("refresh token failed: %s", resp.Status)
	}
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return err
	}
	tr.ObtainedAt = time.Now()
	c.token = &tr
	return nil
}

// ensureToken ensures a valid access token is present, refreshing if needed.
func (c *Client) ensureToken() error {
	c.tokenMu.Lock()
	if c.token == nil || time.Since(c.token.ObtainedAt) > time.Duration(c.token.ExpiresIn-60)*time.Second {
		c.tokenMu.Unlock()
		return c.Login()
	}
	c.tokenMu.Unlock()
	return nil
}

// doRequest performs an HTTP request with authentication, handling token refresh on 401.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	c.tokenMu.Lock()
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	c.tokenMu.Unlock()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err

	}
	if resp.StatusCode == 401 {
		// Try refresh
		_ = resp.Body.Close()
		if err := c.RefreshToken(); err != nil {
			return nil, err
		}
		c.tokenMu.Lock()
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		c.tokenMu.Unlock()
		return c.httpClient.Do(req)
	}
	return resp, nil
}

// BalanceResponse represents the response from /v1/balance
// See datacrunch-models-docs.txt BalanceResponseDto
// Example: { "amount": 1000, "currency": "usd" }
type BalanceResponse struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// GetBalance returns the project balance.
func (c *Client) GetBalance() (*BalanceResponse, error) {
	endpoint := c.baseURL + "/balance"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var balance BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// ImageInfoResponseDto represents the response from /v1/images
// See datacrunch-models-docs.txt ImageInfoResponseDto
type ImageInfoResponseDto struct {
	ID        string   `json:"id"`         // Unique identifier for the image
	ImageType string   `json:"image_type"` // Type of the image
	Name      string   `json:"name"`       // Name of the image
	IsDefault bool     `json:"is_default"` // Indicates if this is the default image
	Category  string   `json:"category"`   // Category of the image
	IsCluster bool     `json:"is_cluster"` // Indicates if the image is for a cluster
	Details   []string `json:"details"`    // Additional details about the image
}

// ListImages returns the list of available image types.
func (c *Client) ListImages() ([]ImageInfoResponseDto, error) {
	endpoint := c.baseURL + "/images"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var images []ImageInfoResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, err
	}
	return images, nil
}

// Token returns a copy of the current tokenResponse (for use by provider)
func (c *Client) Token() tokenResponse {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token == nil {
		return tokenResponse{}
	}
	return *c.token
}
