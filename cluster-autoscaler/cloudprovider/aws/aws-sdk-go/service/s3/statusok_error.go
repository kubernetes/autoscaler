package s3

import (
	"bytes"
	"io"
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/awserr"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/internal/sdkio"
)

func copyMultipartStatusOKUnmarshalError(r *request.Request) {
	b, err := io.ReadAll(r.HTTPResponse.Body)
	r.HTTPResponse.Body.Close()
	if err != nil {
		r.Error = awserr.NewRequestFailure(
			awserr.New(request.ErrCodeSerialization, "unable to read response body", err),
			r.HTTPResponse.StatusCode,
			r.RequestID,
		)
		// Note, some middleware later in the stack like restxml.Unmarshal expect a valid, non-closed Body
		// even in case of an error, so we replace it with an empty Reader.
		r.HTTPResponse.Body = io.NopCloser(bytes.NewBuffer(nil))
		return
	}

	body := bytes.NewReader(b)
	r.HTTPResponse.Body = io.NopCloser(body)
	defer body.Seek(0, sdkio.SeekStart)

	unmarshalError(r)
	if err, ok := r.Error.(awserr.Error); ok && err != nil {
		if err.Code() == request.ErrCodeSerialization &&
			err.OrigErr() != io.EOF {
			r.Error = nil
			return
		}
		// if empty payload
		if err.OrigErr() == io.EOF {
			r.HTTPResponse.StatusCode = http.StatusInternalServerError
		} else {
			r.HTTPResponse.StatusCode = http.StatusServiceUnavailable
		}
	}
}
