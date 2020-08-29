package paperspace

import (
	"net/http"
)

type Backend interface {
	Request(method string, url string, params, result interface{}, requestParams RequestParams) (*http.Response, error)
}
