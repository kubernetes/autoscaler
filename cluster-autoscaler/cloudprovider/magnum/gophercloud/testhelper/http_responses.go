package testhelper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// Mux is a multiplexer that can be used to register handlers.
	Mux *http.ServeMux

	// Server is an in-memory HTTP server for testing.
	Server *httptest.Server
)

// SetupPersistentPortHTTP prepares the Mux and Server listening specific port.
func SetupPersistentPortHTTP(t *testing.T, port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Errorf("Failed to listen to 127.0.0.1:%d: %s", port, err)
	}
	Mux = http.NewServeMux()
	Server = httptest.NewUnstartedServer(Mux)
	Server.Listener = l
	Server.Start()
}

// SetupHTTP prepares the Mux and Server.
func SetupHTTP() {
	Mux = http.NewServeMux()
	Server = httptest.NewServer(Mux)
}

// TeardownHTTP releases HTTP-related resources.
func TeardownHTTP() {
	Server.Close()
}

// Endpoint returns a fake endpoint that will actually target the Mux.
func Endpoint() string {
	return Server.URL + "/"
}

// TestFormValues ensures that all the URL parameters given to the http.Request are the same as values.
func TestFormValues(t *testing.T, r *http.Request, values map[string]string) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(want, r.Form) {
		t.Errorf("Request parameters = %v, want %v", r.Form, want)
	}
}

// TestMethod checks that the Request has the expected method (e.g. GET, POST).
func TestMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

// TestHeader checks that the header on the http.Request matches the expected value.
func TestHeader(t *testing.T, r *http.Request, header string, expected string) {
	if len(r.Header.Values(header)) == 0 {
		t.Errorf("Header %s not found, expected %q", header, expected)
		return
	}
	for _, actual := range r.Header.Values(header) {
		if expected != actual {
			t.Errorf("Header %s = %q, expected %q", header, actual, expected)
		}
	}
}

// TestHeaderUnset checks that the header on the http.Request doesn't exist.
func TestHeaderUnset(t *testing.T, r *http.Request, header string) {
	if len(r.Header.Values(header)) > 0 {
		t.Errorf("Header %s is not expected", header)
	}
}

// TestBody verifies that the request body matches an expected body.
func TestBody(t *testing.T, r *http.Request, expected string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Unable to read body: %v", err)
	}
	str := string(b)
	if expected != str {
		t.Errorf("Body = %s, expected %s", str, expected)
	}
}

// TestJSONRequest verifies that the JSON payload of a request matches an expected structure, without asserting things about
// whitespace or ordering.
func TestJSONRequest(t *testing.T, r *http.Request, expected string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Unable to read request body: %v", err)
	}

	var actualJSON interface{}
	err = json.Unmarshal(b, &actualJSON)
	if err != nil {
		t.Errorf("Unable to parse request body as JSON: %v", err)
	}

	CheckJSONEquals(t, expected, actualJSON)
}
