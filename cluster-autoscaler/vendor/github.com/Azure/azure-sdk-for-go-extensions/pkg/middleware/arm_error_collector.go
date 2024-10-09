package middleware

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

const (
	headerKeyRequestID                                    = "X-Ms-Client-Request-Id"
	headerKeyCorrelationID                                = "X-Ms-Correlation-Request-id"
	ArmErrorCodeCastToArmResponseErrorFailed ArmErrorCode = "CastToArmResponseErrorFailed"
	ArmErrorCodeTransportError               ArmErrorCode = "TransportError"
	ArmErrorCodeUnexpectedTransportError     ArmErrorCode = "UnexpectedTransportError"
	ArmErrorCodeContextCanceled              ArmErrorCode = "ContextCanceled"
	ArmErrorCodeContextDeadlineExceeded      ArmErrorCode = "ContextDeadlineExceeded"
)

// ArmError is unified Error Experience across AzureResourceManager, it contains Code Message.
type ArmError struct {
	Code    ArmErrorCode `json:"code"`
	Message string       `json:"message"`
}

type ArmErrorCode string

type RequestInfo struct {
	Request  *http.Request
	ArmResId *arm.ResourceID
}

func newRequestInfo(req *http.Request, resId *arm.ResourceID) *RequestInfo {
	return &RequestInfo{Request: req, ArmResId: resId}
}

type ResponseInfo struct {
	Response      *http.Response
	Error         *ArmError
	Latency       time.Duration
	RequestId     string
	CorrelationId string
	ConnTracking  *HttpConnTracking
}

type HttpConnTracking struct {
	TotalLatency string
	DnsLatency   string
	ConnLatency  string
	TlsLatency   string
	Protocol     string
	ReqConnInfo  *httptrace.GotConnInfo
}

// ArmRequestMetricCollector is a interface that collectors need to implement.
// TODO: use *policy.Request or *http.Request?
type ArmRequestMetricCollector interface {
	// RequestStarted is called when a request is about to be sent.
	// context is not provided, get it from RequestInfo.Request.Context()
	RequestStarted(*RequestInfo)

	// RequestCompleted is called when a request is finished
	// context is not provided, get it from RequestInfo.Request.Context()
	// if an error occurred, ResponseInfo.Error will be populated
	RequestCompleted(*RequestInfo, *ResponseInfo)
}

// ArmRequestMetricPolicy is a policy that collects metrics/telemetry for ARM requests.
type ArmRequestMetricPolicy struct {
	Collector ArmRequestMetricCollector
}

// Do implements the azcore/policy.Policy interface.
func (p *ArmRequestMetricPolicy) Do(req *policy.Request) (*http.Response, error) {
	httpReq := req.Raw()
	if httpReq == nil || httpReq.URL == nil {
		// not able to collect telemetry, just pass through
		return req.Next()
	}

	armResId, err := arm.ParseResourceID(httpReq.URL.Path)
	if err != nil {
		// TODO: error handling without break the request.
	}

	connTracking := &HttpConnTracking{}
	// have to add to the context at first - then clone the policy.Request struct
	// this allows the connection tracing to happen
	// otherwise we can't change the underlying http request of req, we have to use
	// newARMReq
	newCtx := addConnectionTracingToRequestContext(httpReq.Context(), connTracking)
	newARMReq := req.Clone(newCtx)
	requestInfo := newRequestInfo(httpReq, armResId)
	started := time.Now()

	p.requestStarted(requestInfo)

	var resp *http.Response
	var reqErr error

	// defer this function in case there's a panic somewhere down the pipeline.
	// It's the calling user's responsibility to handle panics, not this policy
	defer func() {
		latency := time.Since(started)
		respInfo := &ResponseInfo{
			Response:     resp,
			Latency:      latency,
			ConnTracking: connTracking,
		}

		if reqErr != nil {
			// either it's a transport error
			// or it is already handled by previous policy
			respInfo.Error = parseTransportError(reqErr)
		} else {
			respInfo.Error = parseArmErrorFromResponse(resp)
		}

		// need to get the request id and correlation id from the response.request header
		// because the headers were added by policy and might be called after this policy
		if resp != nil && resp.Request != nil {
			respInfo.RequestId = resp.Request.Header.Get(headerKeyRequestID)
			respInfo.CorrelationId = resp.Request.Header.Get(headerKeyCorrelationID)
		}

		p.requestCompleted(requestInfo, respInfo)
	}()

	resp, reqErr = newARMReq.Next()
	return resp, reqErr
}

// shortcut function to handle nil collector
func (p *ArmRequestMetricPolicy) requestStarted(iReq *RequestInfo) {
	if p.Collector != nil {
		p.Collector.RequestStarted(iReq)
	}
}

// shortcut function to handle nil collector
func (p *ArmRequestMetricPolicy) requestCompleted(iReq *RequestInfo, iResp *ResponseInfo) {
	if p.Collector != nil {
		p.Collector.RequestCompleted(iReq, iResp)
	}
}

func parseArmErrorFromResponse(resp *http.Response) *ArmError {
	if resp == nil {
		return &ArmError{Code: ArmErrorCodeUnexpectedTransportError, Message: "nil response"}
	}
	if resp.StatusCode > 399 {
		// for 4xx, 5xx response, ARM should include {error:{code, message}} in the body
		err := runtime.NewResponseError(resp)
		respErr := &azcore.ResponseError{}
		if errors.As(err, &respErr) {
			return &ArmError{Code: ArmErrorCode(respErr.ErrorCode), Message: respErr.Error()}
		}
		return &ArmError{Code: ArmErrorCodeCastToArmResponseErrorFailed, Message: fmt.Sprintf("Response body is not in ARM error form: {error:{code, message}}: %s", err.Error())}
	}
	return nil
}

// distinguash
// - Context Cancelled (request configured context to have timeout)
// - ClientTimeout (context still valid, http client have timeout configured)
// - Transport Error (DNS/Dial/TLS/ServerTimeout)
func parseTransportError(err error) *ArmError {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return &ArmError{Code: ArmErrorCodeContextCanceled, Message: err.Error()}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &ArmError{Code: ArmErrorCodeContextDeadlineExceeded, Message: err.Error()}
	}
	return &ArmError{Code: ArmErrorCodeTransportError, Message: err.Error()}
}

func addConnectionTracingToRequestContext(ctx context.Context, connTracking *HttpConnTracking) context.Context {
	var getConn, dnsStart, connStart, tlsStart *time.Time

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			getConn = to.Ptr(time.Now())
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			if getConn != nil {
				connTracking.TotalLatency = fmt.Sprintf("%dms", time.Now().Sub(*getConn).Milliseconds())
			}

			connTracking.ReqConnInfo = &connInfo
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = to.Ptr(time.Now())
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			if dnsInfo.Err == nil {
				if dnsStart != nil {
					connTracking.DnsLatency = fmt.Sprintf("%dms", time.Now().Sub(*dnsStart).Milliseconds())
				}
			} else {
				connTracking.DnsLatency = dnsInfo.Err.Error()
			}
		},
		ConnectStart: func(_, _ string) {
			connStart = to.Ptr(time.Now())
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				if connStart != nil {
					connTracking.ConnLatency = fmt.Sprintf("%dms", time.Now().Sub(*connStart).Milliseconds())
				}
			} else {
				connTracking.ConnLatency = err.Error()
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = to.Ptr(time.Now())
		},
		TLSHandshakeDone: func(t tls.ConnectionState, err error) {
			if err == nil {
				if tlsStart != nil {
					connTracking.TlsLatency = fmt.Sprintf("%dms", time.Now().Sub(*tlsStart).Milliseconds())
				}
				connTracking.Protocol = t.NegotiatedProtocol
			} else {
				connTracking.TlsLatency = err.Error()
			}
		},
	}
	ctx = httptrace.WithClientTrace(ctx, trace)
	return ctx
}
