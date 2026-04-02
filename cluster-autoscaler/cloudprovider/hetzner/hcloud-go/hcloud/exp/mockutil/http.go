package mockutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Request describes a http request that a [httptest.Server] should receive, and the
// corresponding response to return.
//
// Additional checks on the request (e.g. request body) may be added using the
// [Request.Want] function.
//
// The response body is populated from either a JSON struct, or a JSON string.
type Request struct {
	Method string
	Path   string
	Want   func(t *testing.T, r *http.Request)

	Status  int
	JSON    any
	JSONRaw string
	TextRaw string
}

// Handler is using a [Server] to mock http requests provided by the user.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func Handler(t *testing.T, requests []Request) http.HandlerFunc {
	t.Helper()

	server := NewServer(t, requests)
	t.Cleanup(server.close)

	return server.handler
}

// NewServer returns a new mock server that closes itself at the end of the test.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func NewServer(t *testing.T, requests []Request) *Server {
	t.Helper()

	o := &Server{t: t}
	o.Server = httptest.NewServer(http.HandlerFunc(o.handler))
	t.Cleanup(o.close)

	o.Expect(requests)

	return o
}

// Server embeds a [httptest.Server] that answers HTTP calls with a list of expected [Request].
//
// Request matching is based on the request count, and the user provided request will be
// iterated over.
//
// A Server must be created using the [NewServer] function.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
type Server struct {
	*httptest.Server

	t *testing.T

	requests []Request
	index    int
}

// Expect adds requests to the list of requests expected by the [Server].
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func (m *Server) Expect(requests []Request) {
	m.requests = append(m.requests, requests...)
}

func (m *Server) close() {
	m.t.Helper()

	m.Server.Close()

	assert.Equal(m.t, len(m.requests), m.index, "expected more calls")
}

func (m *Server) handler(w http.ResponseWriter, r *http.Request) {
	if testing.Verbose() {
		m.t.Logf("call %d: %s %s\n", m.index, r.Method, r.RequestURI)
	}

	if m.index >= len(m.requests) {
		m.t.Fatalf("received unknown call %d", m.index)
	}

	expected := m.requests[m.index]

	expectedCall := expected.Method
	foundCall := r.Method
	if expected.Path != "" {
		expectedCall += " " + expected.Path
		foundCall += " " + r.RequestURI
	}
	require.Equal(m.t, expectedCall, foundCall) // nolint: testifylint

	if expected.Want != nil {
		expected.Want(m.t, r)
	}

	switch {
	case expected.JSON != nil:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(expected.Status)
		if err := json.NewEncoder(w).Encode(expected.JSON); err != nil {
			m.t.Fatal(err)
		}
	case expected.JSONRaw != "":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(expected.Status)
		_, err := w.Write([]byte(expected.JSONRaw))
		if err != nil {
			m.t.Fatal(err)
		}
	case expected.TextRaw != "":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(expected.Status)
		_, err := w.Write([]byte(expected.TextRaw))
		if err != nil {
			m.t.Fatal(err)
		}
	default:
		w.WriteHeader(expected.Status)
	}

	m.index++
}
