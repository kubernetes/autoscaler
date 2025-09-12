package testing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3"
)

var etag = "myetag"

// TransferManagerLoggingClient is a mock client that can be used to record and stub responses for testing the transfer manager.
type TransferManagerLoggingClient struct {
	// params for upload test

	UploadInvocations []string
	Params            []interface{}

	ConsumeBody bool

	ignoredOperations []string

	PartNum int

	// params for download test

	Data       []byte
	PartsData  [][]byte
	PartsCount int32

	GetObjectInvocations int

	RetrievedRanges []string
	RetrievedParts  []int32
	Versions        []string
	Etags           []string

	ErrReaders []TestErrReader
	index      int

	m sync.Mutex

	PutObjectFn               func(*TransferManagerLoggingClient, *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	UploadPartFn              func(*TransferManagerLoggingClient, *s3.UploadPartInput) (*s3.UploadPartOutput, error)
	CreateMultipartUploadFn   func(*TransferManagerLoggingClient, *s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error)
	CompleteMultipartUploadFn func(*TransferManagerLoggingClient, *s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error)
	AbortMultipartUploadFn    func(*TransferManagerLoggingClient, *s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error)
	GetObjectFn               func(*TransferManagerLoggingClient, *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

func (c *TransferManagerLoggingClient) simulateHTTPClientOption(optFns ...func(*s3.Options)) error {

	o := s3.Options{
		HTTPClient: httpDoFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				Request: r,
			}, nil
		}),
	}

	for _, fn := range optFns {
		fn(&o)
	}

	_, err := o.HTTPClient.Do(&http.Request{
		URL: &url.URL{
			Scheme:   "https",
			Host:     "mock.amazonaws.com",
			Path:     "/key",
			RawQuery: "foo=bar",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

type httpDoFunc func(*http.Request) (*http.Response, error)

func (f httpDoFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

func (c *TransferManagerLoggingClient) traceOperation(name string, params interface{}) {
	if slices.Contains(c.ignoredOperations, name) {
		return
	}
	c.UploadInvocations = append(c.UploadInvocations, name)
	c.Params = append(c.Params, params)

}

// PutObject is the S3 PutObject API.
func (c *TransferManagerLoggingClient) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.ConsumeBody {
		io.Copy(ioutil.Discard, params.Body)
	}

	c.traceOperation("PutObject", params)

	if err := c.simulateHTTPClientOption(optFns...); err != nil {
		return nil, err
	}

	if c.PutObjectFn != nil {
		return c.PutObjectFn(c, params)
	}

	return &s3.PutObjectOutput{
		VersionId: aws.String("VERSION-ID"),
	}, nil
}

// UploadPart is the S3 UploadPart API.
func (c *TransferManagerLoggingClient) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.ConsumeBody {
		io.Copy(ioutil.Discard, params.Body)
	}

	c.traceOperation("UploadPart", params)

	if err := c.simulateHTTPClientOption(optFns...); err != nil {
		return nil, err
	}

	if c.UploadPartFn != nil {
		return c.UploadPartFn(c, params)
	}

	return &s3.UploadPartOutput{
		ETag: aws.String(fmt.Sprintf("ETAG%d", *params.PartNumber)),
	}, nil
}

// CreateMultipartUpload is the S3 CreateMultipartUpload API.
func (c *TransferManagerLoggingClient) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.traceOperation("CreateMultipartUpload", params)

	if err := c.simulateHTTPClientOption(optFns...); err != nil {
		return nil, err
	}

	if c.CreateMultipartUploadFn != nil {
		return c.CreateMultipartUploadFn(c, params)
	}

	return &s3.CreateMultipartUploadOutput{
		UploadId: aws.String("UPLOAD-ID"),
	}, nil
}

// CompleteMultipartUpload is the S3 CompleteMultipartUpload API.
func (c *TransferManagerLoggingClient) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.traceOperation("CompleteMultipartUpload", params)

	if err := c.simulateHTTPClientOption(optFns...); err != nil {
		return nil, err
	}

	if c.CompleteMultipartUploadFn != nil {
		return c.CompleteMultipartUploadFn(c, params)
	}

	return &s3.CompleteMultipartUploadOutput{
		Location:  aws.String("http://location"),
		VersionId: aws.String("VERSION-ID"),
	}, nil
}

// AbortMultipartUpload is the S3 AbortMultipartUpload API.
func (c *TransferManagerLoggingClient) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.traceOperation("AbortMultipartUpload", params)
	if err := c.simulateHTTPClientOption(optFns...); err != nil {
		return nil, err
	}

	if c.AbortMultipartUploadFn != nil {
		return c.AbortMultipartUploadFn(c, params)
	}

	return &s3.AbortMultipartUploadOutput{}, nil
}

// GetObject is the S3 GetObject API.
func (c *TransferManagerLoggingClient) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.GetObjectInvocations++

	if params.Range != nil {
		c.RetrievedRanges = append(c.RetrievedRanges, aws.ToString(params.Range))
	}
	if params.PartNumber != nil {
		c.RetrievedParts = append(c.RetrievedParts, aws.ToInt32(params.PartNumber))
	}
	c.Versions = append(c.Versions, aws.ToString(params.VersionId))
	c.Etags = append(c.Etags, aws.ToString(params.IfMatch))

	if c.GetObjectFn != nil {
		return c.GetObjectFn(c, params)
	}

	return &s3.GetObjectOutput{}, nil
}

// HeadObject is the S3 HeadObject API
func (c *TransferManagerLoggingClient) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	c.m.Lock()
	defer c.m.Unlock()

	return &s3.HeadObjectOutput{
		PartsCount:    aws.Int32(c.PartsCount),
		ContentLength: aws.Int64(int64(len(c.Data))),
		ETag:          aws.String(etag),
	}, nil
}

// NewUploadLoggingClient returns a new TransferManagerLoggingClient for upload testing.
func NewUploadLoggingClient(ignoredOps []string) (*TransferManagerLoggingClient, *[]string, *[]interface{}) {
	c := &TransferManagerLoggingClient{
		ignoredOperations: ignoredOps,
	}

	return c, &c.UploadInvocations, &c.Params
}

// NewDownloadClient returns a new TransferManagerLoggingClient for download testing
func NewDownloadClient() (*TransferManagerLoggingClient, *int, *[]int32, *[]string, *[]string, *[]string) {
	c := &TransferManagerLoggingClient{}

	return c, &c.GetObjectInvocations, &c.RetrievedParts, &c.RetrievedRanges, &c.Versions, &c.Etags
}

var rangeValueRegex = regexp.MustCompile(`bytes=(\d+)-(\d+)`)

func parseRange(rangeValue string) (start, fin int64) {
	rng := rangeValueRegex.FindStringSubmatch(rangeValue)
	start, _ = strconv.ParseInt(rng[1], 10, 64)
	fin, _ = strconv.ParseInt(rng[2], 10, 64)
	return start, fin
}

// RangeGetObjectFn mocks getobject behavior of s3 client to return object in ranges
var RangeGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	start, fin := parseRange(aws.ToString(params.Range))
	fin++

	if fin >= int64(len(c.Data)) {
		fin = int64(len(c.Data))
	}

	bodyBytes := c.Data[start:fin]

	return &s3.GetObjectOutput{
		Body:          ioutil.NopCloser(bytes.NewReader(bodyBytes)),
		ContentRange:  aws.String(fmt.Sprintf("bytes %d-%d/%d", start, fin-1, len(c.Data))),
		ContentLength: aws.Int64(int64(len(bodyBytes))),
		ETag:          aws.String(etag),
	}, nil
}

// ErrRangeGetObjectFn mocks getobject behavior of s3 client to return service error when certain number of range get is called from s3 client
var ErrRangeGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	out, err := RangeGetObjectFn(c, params)
	c.index++
	if c.index > 1 {
		return &s3.GetObjectOutput{}, fmt.Errorf("s3 service error")
	}
	return out, err
}

// MismatchRangeGetObjectFn mocks getobject behavior of s3 client to return mismatch error when object is updated during ranges GET
var MismatchRangeGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	out, err := RangeGetObjectFn(c, params)
	c.index++
	if c.index > 1 {
		return &s3.GetObjectOutput{}, fmt.Errorf("PreconditionFailed")
	}
	return out, err
}

// NonRangeGetObjectFn mocks getobject behavior of s3 client to return the whole object
var NonRangeGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:          ioutil.NopCloser(bytes.NewReader(c.Data[:])),
		ContentLength: aws.Int64(int64(len(c.Data))),
	}, nil
}

// ErrReaderFn mocks getobject behavior of s3 client to return object parts triggering different readerror
var ErrReaderFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	r := c.ErrReaders[c.index]
	out := &s3.GetObjectOutput{
		Body:          ioutil.NopCloser(&r),
		ContentRange:  aws.String(fmt.Sprintf("bytes %d-%d/%d", 0, r.Len-1, r.Len)),
		ContentLength: aws.Int64(r.Len),
		PartsCount:    aws.Int32(c.PartsCount),
	}
	c.index++
	return out, nil
}

// PartGetObjectFn mocks getobject behavior of s3 client to return object parts and total parts count
var PartGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:          ioutil.NopCloser(bytes.NewReader(c.Data)),
		ContentLength: aws.Int64(int64(len(c.Data))),
		PartsCount:    aws.Int32(c.PartsCount),
		ETag:          aws.String(etag),
	}, nil
}

// ReaderPartGetObjectFn mocks getobject behavior of s3 client to return object parts according to params.PartNumber
var ReaderPartGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	index := aws.ToInt32(params.PartNumber) - 1
	return &s3.GetObjectOutput{
		Body:          ioutil.NopCloser(bytes.NewReader(c.PartsData[index])),
		ContentLength: aws.Int64(int64(len(c.PartsData[index]))),
		PartsCount:    aws.Int32(c.PartsCount),
	}, nil
}

// ErrPartGetObjectFn mocks getobject behavior of s3 client to return service error when certain number of part get is called from s3 client
var ErrPartGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	out, err := PartGetObjectFn(c, params)
	c.index++
	if c.index > 1 {
		return &s3.GetObjectOutput{}, fmt.Errorf("s3 service error")
	}
	return out, err
}

// MismatchPartGetObjectFn mocks getobject behavior of s3 client to return mismatch error when object is updated during parts GET
var MismatchPartGetObjectFn = func(c *TransferManagerLoggingClient, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	out, err := PartGetObjectFn(c, params)
	c.index++
	if c.index > 1 {
		return &s3.GetObjectOutput{}, fmt.Errorf("PreconditionFailed")
	}
	return out, err
}

// TestErrReader mocks response's object body triggering specified error when read
type TestErrReader struct {
	Buf []byte
	Err error
	Len int64

	off int
}

// Read implements io.Reader.Read()
func (r *TestErrReader) Read(p []byte) (int, error) {
	to := len(r.Buf) - r.off

	n := copy(p, r.Buf[r.off:to])
	r.off += n

	if n < len(p) {
		return n, r.Err

	}

	return n, nil
}
