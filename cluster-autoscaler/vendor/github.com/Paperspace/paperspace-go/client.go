package paperspace

import (
	"context"
	"net/http"
)

type Client struct {
	APIKey  string
	Backend Backend
}

// client that makes requests to Gradient API
func NewClient() *Client {
	return &Client{
		Backend: NewAPIBackend(),
	}
}

func NewClientWithBackend(backend Backend) *Client {
	return &Client{
		Backend: backend,
	}
}

func (c *Client) Request(ctx context.Context, method string, url string, params, result interface{}) (*http.Response, error) {
	headers := map[string]string{
		"x-api-key": c.APIKey,
	}
	return c.Backend.Request(ctx, method, url, params, result, headers)
}
