package paperspace

import (
	"context"
	"net/http"
)

type RequestParams struct {
	Context context.Context   `json:"-"`
	Headers map[string]string `json:"-"`
}

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

func (c *Client) Request(method string, url string, params, result interface{}, requestParams RequestParams) (*http.Response, error) {
	if requestParams.Headers == nil {
		requestParams.Headers = make(map[string]string)
	}
	requestParams.Headers["x-api-key"] = c.APIKey

	return c.Backend.Request(method, url, params, result, requestParams)
}
