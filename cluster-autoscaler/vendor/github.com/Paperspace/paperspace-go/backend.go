package paperspace

import (
	"net/http"
)

type Backend interface {
	Request(method, url string, params, result interface{}, headers map[string]string) (*http.Response, error)
}
