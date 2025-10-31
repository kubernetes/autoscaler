package transfermanager

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
)

const userAgentKey = "s3-transfer"

// defaultMaxUploadParts is the maximum allowed number of parts in a multi-part upload
// on Amazon S3.
const defaultMaxUploadParts = 10000

// defaultPartSizeBytes is the default part size when transferring objects to/from S3
const minPartSizeBytes = 1024 * 1024 * 8

// defaultMultipartUploadThreshold is the default size threshold in bytes indicating when to use multipart upload.
const defaultMultipartUploadThreshold = 1024 * 1024 * 16

// defaultTransferConcurrency is the default number of goroutines to spin up when
// using PutObject().
const defaultTransferConcurrency = 5

const defaultPartBodyMaxRetries = 3

const defaultGetBufferSize = 1024 * 1024 * 50

// Client provides the API client to make operations call for Amazon Simple
// Storage Service's Transfer Manager
// It is safe to call Client methods concurrently across goroutines.
type Client struct {
	options Options
}

// New returns an initialized Client from the client Options. Provide
// more functional options to further configure the Client
func New(s3Client S3APIClient, opts Options, optFns ...func(*Options)) *Client {
	opts.S3 = s3Client
	for _, fn := range optFns {
		fn(&opts)
	}

	resolveConcurrency(&opts)
	resolvePartSizeBytes(&opts)
	resolveChecksumAlgorithm(&opts)
	resolveMultipartUploadThreshold(&opts)
	resolveGetObjectType(&opts)
	resolvePartBodyMaxRetries(&opts)
	resolveGetBufferSize(&opts)

	return &Client{
		options: opts,
	}
}

// NewFromConfig returns a new Client from the provided s3 config
func NewFromConfig(s3Client S3APIClient, cfg aws.Config, optFns ...func(*Options)) *Client {
	return New(s3Client, Options{}, optFns...)
}
