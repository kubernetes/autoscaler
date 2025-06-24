// Package eventstreamtesting implements helper utilities for event stream protocol testing.
package eventstreamtesting

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/protocol/eventstream/eventstreamapi"
	awshttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/transport/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/credentials"
)

const (
	errClientDisconnected = "client disconnected"
	errStreamClosed       = "http2: stream closed"

	// x/net had an exported StreamError type that we could assert against,
	// net/http's h2 implementation internalizes all of its error types but the
	// Error() text pattern remains identical
	http2StreamError = "stream error: stream ID"
)

func setupServer(server *httptest.Server) aws.HTTPClient {
	server.Config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	server.Config.TLSConfig.NextProtos = []string{"h2"}
	server.TLS = server.Config.TLSConfig

	server.StartTLS()

	buildableClient := awshttp.NewBuildableClient().WithTransportOptions(func(transport *http.Transport) {
		transport.ForceAttemptHTTP2 = true
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	})

	return buildableClient
}

// SetupEventStream configures an HTTPS event stream testing server.
func SetupEventStream(
	t *testing.T, handler http.Handler,
) (
	cfg aws.Config, cleanupFn func(), err error,
) {
	server := httptest.NewUnstartedServer(handler)

	client := setupServer(server)

	cleanupFn = func() {
		server.Close()
	}

	cfg.Credentials = credentials.NewStaticCredentialsProvider("KEYID", "SECRET", "TOKEN")
	cfg.HTTPClient = client
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{URL: server.URL}, nil
	})

	return cfg, cleanupFn, nil
}

// StaticResponse provides a way to define an HTTP event stream server that provides a fixed
// static response.
type StaticResponse struct {
	StatusCode int
	Body       []byte
}

// ServeEventStream provides serving EventStream messages from a HTTP server to
// the client. The events are sent sequentially to the client without delay.
type ServeEventStream struct {
	T             *testing.T
	BiDirectional bool

	StaticResponse *StaticResponse

	Events       []eventstream.Message
	ClientEvents []eventstream.Message

	ForceCloseAfter time.Duration

	requestsIdx int
}

// ServeHTTP serves an HTTP client request
func (s ServeEventStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.StaticResponse != nil {
		w.WriteHeader(s.StaticResponse.StatusCode)
		w.(http.Flusher).Flush()
		if _, err := w.Write(s.StaticResponse.Body); err != nil {
			s.T.Errorf("failed to write response body error: %v", err)
		}
		return
	}

	if s.BiDirectional {
		s.serveBiDirectionalStream(w, r)
	} else {
		s.serveReadOnlyStream(w, r)
	}
}

func (s *ServeEventStream) serveReadOnlyStream(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	encoder := eventstream.NewEncoder()

	for _, event := range s.Events {
		encoder.Encode(flushWriter{w}, event)
	}
}

func (s *ServeEventStream) serveBiDirectionalStream(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	ctx := context.Background()
	if s.ForceCloseAfter > 0 {
		var cancelFunc func()
		ctx, cancelFunc = context.WithTimeout(context.Background(), s.ForceCloseAfter)
		defer cancelFunc()
	}

	var (
		err error
		m   sync.Mutex
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		readErr := s.readEvents(ctx, r)
		m.Lock()
		if readErr != nil && err == nil {
			err = readErr
		}
		m.Unlock()
	}()

	w.(http.Flusher).Flush()

	writeErr := s.writeEvents(ctx, w)
	m.Lock()
	if writeErr != nil && err == nil {
		err = writeErr
	}
	m.Unlock()

	wg.Wait()

	if err != nil && isError(err) {
		s.T.Error(err.Error())
	}
}

func isError(err error) bool {
	for _, s := range []string{errClientDisconnected, errStreamClosed, http2StreamError} {
		if strings.Contains(err.Error(), s) {
			return false
		}
	}

	return true
}

func (s ServeEventStream) readEvents(ctx context.Context, r *http.Request) error {
	signBuffer := make([]byte, 1024)
	messageBuffer := make([]byte, 1024)
	decoder := eventstream.NewDecoder()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// unwrap signing envelope
		signedMessage, err := decoder.Decode(r.Body, signBuffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// empty payload is expected for the last signing message
		if len(signedMessage.Payload) == 0 {
			break
		}

		// get service event message from payload
		msg, err := decoder.Decode(bytes.NewReader(signedMessage.Payload), messageBuffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(s.ClientEvents) > 0 {
			i := s.requestsIdx
			s.requestsIdx++

			if e, a := s.ClientEvents[i], msg; !reflect.DeepEqual(e, a) {
				return fmt.Errorf("expected %v, got %v", e, a)
			}
		}
	}

	return nil
}

func (s *ServeEventStream) writeEvents(ctx context.Context, w http.ResponseWriter) error {
	encoder := eventstream.NewEncoder()

	var event eventstream.Message
	pendingEvents := s.Events

	for len(pendingEvents) > 0 {
		event, pendingEvents = pendingEvents[0], pendingEvents[1:]
		select {
		case <-ctx.Done():
			return nil
		default:
			err := encoder.Encode(flushWriter{w}, event)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("expected no error encoding event, got %v", err)
			}
		}
	}

	return nil
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

// EventMessageTypeHeader is an event message type header for specifying an
// event is an message type.
var EventMessageTypeHeader = eventstream.Header{
	Name:  eventstreamapi.MessageTypeHeader,
	Value: eventstream.StringValue(eventstreamapi.EventMessageType),
}

// EventExceptionTypeHeader is an event exception type header for specifying an
// event is an exception type.
var EventExceptionTypeHeader = eventstream.Header{
	Name:  eventstreamapi.MessageTypeHeader,
	Value: eventstream.StringValue(eventstreamapi.ExceptionMessageType),
}
