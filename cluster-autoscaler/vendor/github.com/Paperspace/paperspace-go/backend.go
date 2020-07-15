package paperspace

import (
	"context"
	"net/http"
)

type Backend interface {
	Request(ctx context.Context, method string, url string, params, result interface{}, headers map[string]string) (*http.Response, error)
}
