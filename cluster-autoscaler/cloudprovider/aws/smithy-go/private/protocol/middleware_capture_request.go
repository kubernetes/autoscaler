package protocol

import (
	"context"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
	"net/http"
	"strconv"
)

const captureRequestID = "CaptureProtocolTestRequest"

// AddCaptureRequestMiddleware captures serialized http request during protocol test for check
func AddCaptureRequestMiddleware(stack *middleware.Stack, req *http.Request) error {
	return stack.Build.Add(&captureRequestMiddleware{
		req: req,
	}, middleware.After)
}

type captureRequestMiddleware struct {
	req *http.Request
}

func (*captureRequestMiddleware) ID() string {
	return captureRequestID
}

func (m *captureRequestMiddleware) HandleBuild(ctx context.Context, input middleware.BuildInput, next middleware.BuildHandler,
) (
	output middleware.BuildOutput, metadata middleware.Metadata, err error,
) {
	request, ok := input.Request.(*smithyhttp.Request)
	if !ok {
		return output, metadata, fmt.Errorf("error while retrieving http request")
	}

	*m.req = *request.Build(ctx)
	if len(m.req.URL.RawPath) == 0 {
		m.req.URL.RawPath = m.req.URL.Path
	}
	if v := m.req.ContentLength; v != 0 {
		m.req.Header.Set("Content-Length", strconv.FormatInt(v, 10))
	}

	return next.HandleBuild(ctx, input)
}
