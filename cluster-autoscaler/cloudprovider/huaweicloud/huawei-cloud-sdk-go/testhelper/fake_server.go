package testhelper

import (
	"net/http"
	"net/http/httptest"
)

var (
	Mux *http.ServeMux
	Server *httptest.Server
)

func CreateServer() {
	Mux = http.NewServeMux()
	Server = httptest.NewServer(Mux)
}

func ShutDownServer() {
	Server.Close()
}

func GetEndpoint() string {
	return Server.URL
}