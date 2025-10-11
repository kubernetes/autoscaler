package protocoltest

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// HTTPClient is a mock http client used by protocol test cases to
// respond success response back
type HTTPClient struct{}

// Do returns a mock success response to caller
func (*HTTPClient) Do(request *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Header:     request.Header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
		Request:    request,
	}, nil
}

// NewClient returns pointer of a new HTTPClient for protocol test client
func NewClient() *HTTPClient {
	return &HTTPClient{}
}
