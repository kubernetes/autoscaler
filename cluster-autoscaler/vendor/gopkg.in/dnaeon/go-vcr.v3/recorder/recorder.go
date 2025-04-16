// Copyright (c) 2015-2022 Marin Atanasov Nikolov <dnaeon@gmail.com>
// Copyright (c) 2016 David Jack <davars@gmail.com>
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

package recorder

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

// Mode represents the mode of operation of the recorder
type Mode int

// Recorder states
const (
	// ModeRecordOnly specifies that VCR will run in recording
	// mode only. HTTP interactions will be recorded for each
	// interaction. If the cassette file is present, it will be
	// overwritten.
	ModeRecordOnly Mode = iota

	// ModeReplayOnly specifies that VCR will only replay
	// interactions from previously recorded cassette. If an
	// interaction is missing from the cassette it will return
	// ErrInteractionNotFound error. If the cassette file is
	// missing it will return ErrCassetteNotFound error.
	ModeReplayOnly

	// ModeReplayWithNewEpisodes starts the recorder in replay
	// mode, where existing interactions are returned from the
	// cassette, and missing ones will be recorded and added to
	// the cassette. This mode is useful in cases where you need
	// to update an existing cassette with new interactions, but
	// don't want to wipe out previously recorded interactions.
	// If the cassette file is missing it will create a new one.
	ModeReplayWithNewEpisodes

	// ModeRecordOnce will record new HTTP interactions once only.
	// This mode is useful in cases where you need to record a set
	// of interactions once only and replay only the known
	// interactions. Unknown/missing interactions will cause the
	// recorder to return an ErrInteractionNotFound error. If the
	// cassette file is missing, it will be created.
	ModeRecordOnce

	// ModePassthrough specifies that VCR will not record any
	// interactions at all. In this mode all HTTP requests will be
	// forwarded to the endpoints using the real HTTP transport.
	// In this mode no cassette will be created.
	ModePassthrough
)

// ErrInvalidMode is returned when attempting to start the recorder
// with invalid mode
var ErrInvalidMode = errors.New("invalid recorder mode")

// HookFunc represents a function, which will be invoked in different
// stages of the playback. The hook functions allow for plugging in to
// the playback and transform an interaction, if needed. For example a
// hook function might redact or remove sensitive data from a
// request/response before it is added to the in-memory cassette, or
// before it is saved on disk.  Another use case would be to transform
// the HTTP response before it is returned to the client during replay
// mode.
type HookFunc func(i *cassette.Interaction) error

// Hook kinds
type HookKind int

const (
	// AfterCaptureHook represents a hook, which will be invoked
	// after capturing a request/response pair.
	AfterCaptureHook HookKind = iota

	// BeforeSaveHook represents a hook, which will be invoked
	// right before the cassette is saved on disk.
	BeforeSaveHook

	// BeforeResponseReplayHook represents a hook, which will be
	// invoked before replaying a previously recorded response to
	// the client.
	BeforeResponseReplayHook

	// OnRecorderStopHook is a hook, which will be invoked when the recorder
	// is about to be stopped. This hook is useful for performing any
	// post-actions such as cleanup or reporting.
	OnRecorderStopHook
)

// Hook represents a function hook of a given kind. Depending on the
// hook kind, the function will be invoked in different stages of the
// playback.
type Hook struct {
	// Handler is the function which will be invoked
	Handler HookFunc

	// Kind represents the hook kind
	Kind HookKind
}

// NewHook creates a new hook
func NewHook(handler HookFunc, kind HookKind) *Hook {
	hook := &Hook{
		Handler: handler,
		Kind:    kind,
	}

	return hook
}

// PassthroughFunc is handler which determines whether a specific HTTP
// request is to be forwarded to the original endpoint. It should
// return true when a request needs to be passed through, and false
// otherwise.
type PassthroughFunc func(req *http.Request) bool

// Option represents the Recorder options
type Options struct {
	// CassetteName is the name of the cassette
	CassetteName string

	// Mode is the operating mode of the Recorder
	Mode Mode

	// RealTransport is the underlying http.RoundTripper to make
	// the real requests
	RealTransport http.RoundTripper

	// SkipRequestLatency, if set to true will not simulate the
	// latency of the recorded interaction. When set to false
	// (default) it will block for the period of time taken by the
	// original request to simulate the latency between our
	// recorder and the remote endpoints.
	SkipRequestLatency bool
}

// Recorder represents a type used to record and replay
// client and server interactions
type Recorder struct {
	// Cassette used by the recorder
	cassette *cassette.Cassette

	// Recorder options
	options *Options

	// Passthrough handlers
	passthroughs []PassthroughFunc

	// hooks is a list of hooks, which are invoked in different
	// stages of the playback.
	hooks []*Hook
}

// New creates a new recorder
func New(cassetteName string) (*Recorder, error) {
	opts := &Options{
		CassetteName:       cassetteName,
		Mode:               ModeRecordOnce,
		SkipRequestLatency: false,
		RealTransport:      http.DefaultTransport,
	}

	return NewWithOptions(opts)
}

// NewWithOptions creates a new recorder based on the provided options
func NewWithOptions(opts *Options) (*Recorder, error) {
	if opts.RealTransport == nil {
		opts.RealTransport = http.DefaultTransport
	}

	rec := &Recorder{
		options:      opts,
		passthroughs: make([]PassthroughFunc, 0),
		hooks:        make([]*Hook, 0),
	}

	cassetteFile := cassette.New(opts.CassetteName).File
	_, err := os.Stat(cassetteFile)
	cassetteExists := !os.IsNotExist(err)

	switch {
	case opts.Mode == ModeRecordOnly:
		c := cassette.New(opts.CassetteName)
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModeReplayOnly && !cassetteExists:
		return nil, fmt.Errorf("%w: %s", cassette.ErrCassetteNotFound, cassetteFile)
	case opts.Mode == ModeReplayOnly && cassetteExists:
		c, err := cassette.Load(opts.CassetteName)
		if err != nil {
			return nil, err
		}
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModeReplayWithNewEpisodes && !cassetteExists:
		c := cassette.New(opts.CassetteName)
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModeReplayWithNewEpisodes && cassetteExists:
		c, err := cassette.Load(opts.CassetteName)
		if err != nil {
			return nil, err
		}
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModeRecordOnce && !cassetteExists:
		c := cassette.New(opts.CassetteName)
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModeRecordOnce && cassetteExists:
		c, err := cassette.Load(opts.CassetteName)
		if err != nil {
			return nil, err
		}
		rec.cassette = c
		return rec, nil
	case opts.Mode == ModePassthrough:
		c := cassette.New(opts.CassetteName)
		rec.cassette = c
		return rec, nil
	default:
		return nil, ErrInvalidMode
	}
}

// Proxies client requests to their original destination
func (rec *Recorder) requestHandler(r *http.Request) (*cassette.Interaction, error) {
	if err := r.Context().Err(); err != nil {
		return nil, err
	}

	switch {
	case rec.options.Mode == ModeReplayOnly:
		return rec.cassette.GetInteraction(r)
	case rec.options.Mode == ModeReplayWithNewEpisodes:
		interaction, err := rec.cassette.GetInteraction(r)
		if err == nil {
			// Interaction found, return it
			return interaction, nil
		} else if err == cassette.ErrInteractionNotFound {
			// Interaction not found, we have a new episode
			break
		} else {
			// Any other error is an error
			return nil, err
		}
	case rec.options.Mode == ModeRecordOnce && !rec.cassette.IsNew:
		// We've got an existing cassette, return what we've got
		return rec.cassette.GetInteraction(r)
	case rec.options.Mode == ModePassthrough:
		// Passthrough requests always hit the original endpoint
		break
	case (rec.options.Mode == ModeRecordOnly || rec.options.Mode == ModeRecordOnce) && rec.cassette.ReplayableInteractions:
		// When running with replayable interactions look for
		// existing interaction first, so we avoid hitting
		// multiple times the same endpoint.
		interaction, err := rec.cassette.GetInteraction(r)
		if err == nil {
			// Interaction found, return it
			return interaction, nil
		} else if err == cassette.ErrInteractionNotFound {
			// Interaction not found, we have to record it
			break
		} else {
			// Any other error is an error
			return nil, err
		}
	default:
		// Anything else hits the original endpoint
		break
	}

	// Copy the original request, so we can read the form values
	reqBytes, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, err
	}

	reqBuffer := bytes.NewBuffer(reqBytes)
	copiedReq, err := http.ReadRequest(bufio.NewReader(reqBuffer))
	if err != nil {
		return nil, err
	}

	err = copiedReq.ParseForm()
	if err != nil {
		return nil, err
	}

	reqBody := &bytes.Buffer{}
	if r.Body != nil && r.Body != http.NoBody {
		// Record the request body so we can add it to the cassette
		r.Body = io.NopCloser(io.TeeReader(r.Body, reqBody))
	}

	// Perform request to it's original destination and record the interactions
	var start time.Time
	start = time.Now()
	resp, err := rec.options.RealTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	requestDuration := time.Since(start)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Add interaction to the cassette
	interaction := &cassette.Interaction{
		Request: cassette.Request{
			Proto:            r.Proto,
			ProtoMajor:       r.ProtoMajor,
			ProtoMinor:       r.ProtoMinor,
			ContentLength:    r.ContentLength,
			TransferEncoding: r.TransferEncoding,
			Trailer:          r.Trailer,
			Host:             r.Host,
			RemoteAddr:       r.RemoteAddr,
			RequestURI:       r.RequestURI,
			Body:             reqBody.String(),
			Form:             copiedReq.PostForm,
			Headers:          r.Header,
			URL:              r.URL.String(),
			Method:           r.Method,
		},
		Response: cassette.Response{
			Status:           resp.Status,
			Code:             resp.StatusCode,
			Proto:            resp.Proto,
			ProtoMajor:       resp.ProtoMajor,
			ProtoMinor:       resp.ProtoMinor,
			TransferEncoding: resp.TransferEncoding,
			Trailer:          resp.Trailer,
			ContentLength:    resp.ContentLength,
			Uncompressed:     resp.Uncompressed,
			Body:             string(respBody),
			Headers:          resp.Header,
			Duration:         requestDuration,
		},
	}

	// Apply after-capture hooks before we add the interaction to
	// the in-memory cassette.
	if err := rec.applyHooks(interaction, AfterCaptureHook); err != nil {
		return nil, err
	}

	rec.cassette.AddInteraction(interaction)

	return interaction, nil
}

// Stop is used to stop the recorder and save any recorded
// interactions if running in one of the recording modes. When
// running in ModePassthrough no cassette will be saved on disk.
func (rec *Recorder) Stop() error {
	cassetteFile := rec.cassette.File
	_, err := os.Stat(cassetteFile)
	cassetteExists := !os.IsNotExist(err)

	// Nothing to do for ModeReplayOnly and ModePassthrough here
	switch {
	case rec.options.Mode == ModeRecordOnly || rec.options.Mode == ModeReplayWithNewEpisodes:
		if err := rec.persistCassette(); err != nil {
			return err
		}

	case rec.options.Mode == ModeRecordOnce && !cassetteExists:
		if err := rec.persistCassette(); err != nil {
			return err
		}
	}

	// Apply on-recorder-stop hooks
	for _, interaction := range rec.cassette.Interactions {
		if err := rec.applyHooks(interaction, OnRecorderStopHook); err != nil {
			return err
		}
	}

	return nil
}

// persisteCassette persists the cassette on disk for future re-use
func (rec *Recorder) persistCassette() error {
	// Apply any before-save hooks
	for _, interaction := range rec.cassette.Interactions {
		if err := rec.applyHooks(interaction, BeforeSaveHook); err != nil {
			return err
		}
	}

	return rec.cassette.Save()
}

// applyHooks applies the registered hooks of the given kind with the
// specified interaction
func (rec *Recorder) applyHooks(i *cassette.Interaction, kind HookKind) error {
	for _, hook := range rec.hooks {
		if hook.Kind == kind {
			if err := hook.Handler(i); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetRealTransport can be used to configure the real HTTP transport
// of the recorder.
func (rec *Recorder) SetRealTransport(t http.RoundTripper) {
	rec.options.RealTransport = t
}

// RoundTrip implements the http.RoundTripper interface
func (rec *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	// Passthrough mode, use real transport
	if rec.options.Mode == ModePassthrough {
		return rec.options.RealTransport.RoundTrip(req)
	}

	// Apply passthrough handler functions
	for _, passthroughFunc := range rec.passthroughs {
		if passthroughFunc(req) {
			return rec.options.RealTransport.RoundTrip(req)
		}
	}

	interaction, err := rec.requestHandler(req)
	if err != nil {
		return nil, err
	}

	// Apply before-response-replay hooks
	if err := rec.applyHooks(interaction, BeforeResponseReplayHook); err != nil {
		return nil, err
	}

	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
		// Apply the duration defined in the interaction
		if !rec.options.SkipRequestLatency {
			<-time.After(interaction.Response.Duration)
		}

		return interaction.GetHTTPResponse()
	}
}

// CancelRequest implements the
// github.com/coreos/etcd/client.CancelableTransport interface
func (rec *Recorder) CancelRequest(req *http.Request) {
	type cancelableTransport interface {
		CancelRequest(req *http.Request)
	}
	if ct, ok := rec.options.RealTransport.(cancelableTransport); ok {
		ct.CancelRequest(req)
	}
}

// SetMatcher sets a function to match requests against recorded HTTP
// interactions.
func (rec *Recorder) SetMatcher(matcher cassette.MatcherFunc) {
	rec.cassette.Matcher = matcher
}

// SetReplayableInteractions defines whether to allow interactions to
// be replayed or not. This is useful in cases when you need to hit
// the same endpoint multiple times and want to replay the interaction
// from the cassette, instead of hiting the endpoint.
func (rec *Recorder) SetReplayableInteractions(replayable bool) {
	rec.cassette.ReplayableInteractions = replayable
}

// AddPassthrough appends a hook to determine if a request should be
// ignored by the recorder.
func (rec *Recorder) AddPassthrough(pass PassthroughFunc) {
	rec.passthroughs = append(rec.passthroughs, pass)
}

// AddHook appends a hook to the recorder. Depending on the hook kind,
// the handler will be invoked in different stages of the playback.
func (rec *Recorder) AddHook(handler HookFunc, kind HookKind) {
	hook := NewHook(handler, kind)
	rec.hooks = append(rec.hooks, hook)
}

// Mode returns recorder state
func (rec *Recorder) Mode() Mode {
	return rec.options.Mode
}

// GetDefaultClient returns an HTTP client with a pre-configured
// transport
func (rec *Recorder) GetDefaultClient() *http.Client {
	client := &http.Client{
		Transport: rec,
	}

	return client
}

// IsNewCassette returns true, if the recorder was started with a
// new/empty cassette. Returns false, if it was started using an
// existing cassette, which was loaded.
func (rec *Recorder) IsNewCassette() bool {
	return rec.cassette.IsNew
}

// IsRecording returns true, if the recorder is recording
// interactions, returns false otherwise. Note, that in some modes
// (e.g. ModeReplayWithNewEpisodes and ModeRecordOnce) the recorder
// might be recording new interactions. For example in ModeRecordOnce,
// we are replaying interactions only if there was an existing
// cassette, and we are recording it, if the cassette is a new one.
// ModeReplayWithNewEpisodes would replay interactions, if they are
// present in the cassette, but will also record new ones, if they are
// not part of the cassette already. In these cases the recorder is
// considered to be recording for these modes.
func (rec *Recorder) IsRecording() bool {
	switch {
	case rec.options.Mode == ModeRecordOnly || rec.options.Mode == ModeReplayWithNewEpisodes:
		return true
	case rec.options.Mode == ModeReplayOnly || rec.options.Mode == ModePassthrough:
		return false
	case rec.options.Mode == ModeRecordOnce && rec.IsNewCassette():
		return true
	default:
		return false
	}
}
