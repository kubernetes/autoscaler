package paperspace

import (
	"context"
	"net/http"
	"os"
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
	client := Client{
		Backend: NewAPIBackend(),
	}

	apiKey := os.Getenv("PAPERSPACE_APIKEY")
	if apiKey != "" {
		client.APIKey = apiKey
	}

	return &client
}

func NewClientWithBackend(backend Backend) *Client {
	client := NewClient()
	client.Backend = backend

	return client
}

func (c *Client) Request(method string, url string, params, result interface{}, requestParams RequestParams) (*http.Response, error) {
	if requestParams.Headers == nil {
		requestParams.Headers = make(map[string]string)
	}
	requestParams.Headers["x-api-key"] = c.APIKey

	return c.Backend.Request(method, url, params, result, requestParams)
}
