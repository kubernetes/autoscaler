// Copyright (c) 2015-2022 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package cassette

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Cassette format versions
const (
	// Version 1 of the cassette format
	CassetteFormatV1 = 1

	// Version 2 of the cassette format
	CassetteFormatV2 = 2
)

var (
	// ErrInteractionNotFound indicates that a requested
	// interaction was not found in the cassette file
	ErrInteractionNotFound = errors.New("requested interaction not found")

	// ErrCassetteNotFound indicates that a requested
	// casette doesn't exist (only in Replaying mode)
	ErrCassetteNotFound = errors.New("requested cassette not found")

	// ErrUnsupportedCassetteFormat is returned when attempting to
	// use an older and potentially unsupported format of a
	// cassette
	ErrUnsupportedCassetteFormat = fmt.Errorf("required version of cassette is v%d", CassetteFormatV2)
)

// Request represents a client request as recorded in the
// cassette file
type Request struct {
	Proto            string      `yaml:"proto"`
	ProtoMajor       int         `yaml:"proto_major"`
	ProtoMinor       int         `yaml:"proto_minor"`
	ContentLength    int64       `yaml:"content_length"`
	TransferEncoding []string    `yaml:"transfer_encoding"`
	Trailer          http.Header `yaml:"trailer"`
	Host             string      `yaml:"host"`
	RemoteAddr       string      `yaml:"remote_addr"`
	RequestURI       string      `yaml:"request_uri"`

	// Body of request
	Body string `yaml:"body"`

	// Form values
	Form url.Values `yaml:"form"`

	// Request headers
	Headers http.Header `yaml:"headers"`

	// Request URL
	URL string `yaml:"url"`

	// Request method
	Method string `yaml:"method"`
}

// Response represents a server response as recorded in the
// cassette file
type Response struct {
	Proto            string      `yaml:"proto"`
	ProtoMajor       int         `yaml:"proto_major"`
	ProtoMinor       int         `yaml:"proto_minor"`
	TransferEncoding []string    `yaml:"transfer_encoding"`
	Trailer          http.Header `yaml:"trailer"`
	ContentLength    int64       `yaml:"content_length"`
	Uncompressed     bool        `yaml:"uncompressed"`

	// Body of response
	Body string `yaml:"body"`

	// Response headers
	Headers http.Header `yaml:"headers"`

	// Response status message
	Status string `yaml:"status"`

	// Response status code
	Code int `yaml:"code"`

	// Response duration (something like "100ms" or "10s")
	Duration time.Duration `yaml:"duration"`
}

// Interaction type contains a pair of request/response for a
// single HTTP interaction between a client and a server
type Interaction struct {
	// ID is the id of the interaction
	ID int `yaml:"id"`

	// Request is the recorded request
	Request Request `yaml:"request"`

	// Response is the recorded response
	Response Response `yaml:"response"`

	// DiscardOnSave if set to true will discard the interaction
	// as a whole and it will not be part of the final
	// interactions when saving the cassette on disk.
	DiscardOnSave bool `yaml:"-"`

	// replayed is true when this interaction has been played
	// already.
	replayed bool `yaml:"-"`
}

// WasReplayed returns a boolean indicating whether the given interaction was
// already replayed.
func (i *Interaction) WasReplayed() bool {
	return i.replayed
}

// GetHTTPRequest converts the recorded interaction request to
// http.Request instance
func (i *Interaction) GetHTTPRequest() (*http.Request, error) {
	url, err := url.Parse(i.Request.URL)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Proto:            i.Request.Proto,
		ProtoMajor:       i.Request.ProtoMajor,
		ProtoMinor:       i.Request.ProtoMinor,
		ContentLength:    i.Request.ContentLength,
		TransferEncoding: i.Request.TransferEncoding,
		Trailer:          i.Request.Trailer,
		Host:             i.Request.Host,
		RemoteAddr:       i.Request.RemoteAddr,
		RequestURI:       i.Request.RequestURI,
		Body:             io.NopCloser(strings.NewReader(i.Request.Body)),
		Form:             i.Request.Form,
		Header:           i.Request.Headers,
		URL:              url,
		Method:           i.Request.Method,
	}

	return req, nil
}

// GetHTTPResponse converts the recorded interaction response to
// http.Response instance
func (i *Interaction) GetHTTPResponse() (*http.Response, error) {
	req, err := i.GetHTTPRequest()
	if err != nil {
		return nil, err
	}

	resp := &http.Response{
		Status:           i.Response.Status,
		StatusCode:       i.Response.Code,
		Proto:            i.Response.Proto,
		ProtoMajor:       i.Response.ProtoMajor,
		ProtoMinor:       i.Response.ProtoMinor,
		TransferEncoding: i.Response.TransferEncoding,
		Trailer:          i.Response.Trailer,
		ContentLength:    i.Response.ContentLength,
		Uncompressed:     i.Response.Uncompressed,
		Body:             io.NopCloser(strings.NewReader(i.Response.Body)),
		Header:           i.Response.Headers,
		Close:            true,
		Request:          req,
	}

	return resp, nil
}

// MatcherFunc function returns true when the actual request matches a
// single HTTP interaction's request according to the function's own
// criteria.
type MatcherFunc func(*http.Request, Request) bool

// DefaultMatcher is used when a custom matcher is not defined and
// compares only the method and of the HTTP request.
func DefaultMatcher(r *http.Request, i Request) bool {
	return r.Method == i.Method && r.URL.String() == i.URL
}

// Cassette type
type Cassette struct {
	// Name of the cassette
	Name string `yaml:"-"`

	// File name of the cassette as written on disk
	File string `yaml:"-"`

	// Cassette format version
	Version int `yaml:"version"`

	// Mutex to lock accessing Interactions. omitempty is set to
	// prevent the mutex appearing in the recorded YAML.
	Mu sync.RWMutex `yaml:"mu,omitempty"`

	// Interactions between client and server
	Interactions []*Interaction `yaml:"interactions"`

	// ReplayableInteractions defines whether to allow
	// interactions to be replayed or not
	ReplayableInteractions bool `yaml:"-"`

	// Matches actual request with interaction requests.
	Matcher MatcherFunc `yaml:"-"`

	// IsNew specifies whether this is a newly created cassette.
	// Returns false, when the cassette was loaded from an
	// existing source, e.g. a file.
	IsNew bool `yaml:"-"`

	nextInteractionId int `yaml:"-"`
}

// New creates a new empty cassette
func New(name string) *Cassette {
	c := &Cassette{
		Name:                   name,
		File:                   fmt.Sprintf("%s.yaml", name),
		Version:                CassetteFormatV2,
		Interactions:           make([]*Interaction, 0),
		Matcher:                DefaultMatcher,
		ReplayableInteractions: false,
		IsNew:                  true,
		nextInteractionId:      0,
	}

	return c
}

// Load reads a cassette file from disk
func Load(name string) (*Cassette, error) {
	c := New(name)
	data, err := os.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	c.IsNew = false
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	if c.Version != CassetteFormatV2 {
		return nil, ErrUnsupportedCassetteFormat
	}
	c.nextInteractionId = len(c.Interactions)

	return c, err
}

// AddInteraction appends a new interaction to the cassette
func (c *Cassette) AddInteraction(i *Interaction) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	i.ID = c.nextInteractionId
	c.nextInteractionId += 1
	c.Interactions = append(c.Interactions, i)
}

// GetInteraction retrieves a recorded request/response interaction
func (c *Cassette) GetInteraction(r *http.Request) (*Interaction, error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	for _, i := range c.Interactions {
		if (c.ReplayableInteractions || !i.replayed) && c.Matcher(r, i.Request) {
			i.replayed = true
			return i, nil
		}
	}

	return nil, ErrInteractionNotFound
}

// Save writes the cassette data on disk for future re-use
func (c *Cassette) Save() error {
	c.Mu.RLock()
	defer c.Mu.RUnlock()

	// Create directory for cassette if missing
	cassetteDir := filepath.Dir(c.File)
	if _, err := os.Stat(cassetteDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cassetteDir, 0755); err != nil {
			return err
		}
	}

	// Filter out interactions which should be discarded. While
	// discarding interactions we should also fix the interaction
	// IDs, so that we don't introduce gaps in the final results.
	nextId := 0
	interactions := make([]*Interaction, 0)
	for _, i := range c.Interactions {
		if !i.DiscardOnSave {
			i.ID = nextId
			interactions = append(interactions, i)
			nextId += 1
		}
	}
	c.Interactions = interactions

	// Marshal to YAML and save interactions
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	f, err := os.Create(c.File)
	if err != nil {
		return err
	}

	defer f.Close()

	// Honor the YAML structure specification
	// http://www.yaml.org/spec/1.2/spec.html#id2760395
	_, err = f.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}
