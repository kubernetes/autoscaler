package transfermanager

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
)

// Options provides params needed for transfer api calls
type Options struct {
	// The client to use when uploading to S3.
	S3 S3APIClient

	// The buffer size (in bytes) to use when buffering data into chunks and
	// sending them as parts to S3. The minimum allowed part size is 5MB, and
	// if this value is set to zero, the DefaultUploadPartSize value will be used.
	PartSizeBytes int64

	// The threshold bytes to decide when the file should be multi-uploaded
	MultipartUploadThreshold int64

	// Option to disable checksum validation for download
	DisableChecksumValidation bool

	// Checksum algorithm to use for upload
	ChecksumAlgorithm types.ChecksumAlgorithm

	// The number of goroutines to spin up in parallel per call to Upload when
	// sending parts. If this is set to zero, the DefaultUploadConcurrency value
	// will be used.
	//
	// The concurrency pool is not shared between calls to Upload.
	Concurrency int

	// The type indicating if object is multi-downloaded in parts or ranges
	GetObjectType types.GetObjectType

	// PartBodyMaxRetries is the number of retry attempts to make for failed part downloads.
	PartBodyMaxRetries int

	// Max size for the get object buffer
	GetBufferSize int64

	// Registry of progress listener hooks.
	//
	// It is safe to modify the registry in per-operation functional options,
	// the original client-level registry will not be affected.
	ProgressListeners ProgressListeners
}

func (o *Options) init() {
}

func resolveConcurrency(o *Options) {
	if o.Concurrency == 0 {
		o.Concurrency = defaultTransferConcurrency
	}
}

func resolvePartSizeBytes(o *Options) {
	if o.PartSizeBytes == 0 {
		o.PartSizeBytes = minPartSizeBytes
	}
}

func resolveChecksumAlgorithm(o *Options) {
	if o.ChecksumAlgorithm == "" {
		o.ChecksumAlgorithm = types.ChecksumAlgorithmCrc32
	}
}

func resolveMultipartUploadThreshold(o *Options) {
	if o.MultipartUploadThreshold == 0 {
		o.MultipartUploadThreshold = defaultMultipartUploadThreshold
	}
}

func resolveGetObjectType(o *Options) {
	if o.GetObjectType == "" {
		o.GetObjectType = types.GetObjectParts
	}
}

func resolvePartBodyMaxRetries(o *Options) {
	if o.PartBodyMaxRetries == 0 {
		o.PartBodyMaxRetries = defaultPartBodyMaxRetries
	}
}

func resolveGetBufferSize(o *Options) {
	if o.GetBufferSize == 0 {
		o.GetBufferSize = defaultGetBufferSize
	}
}

// Copy returns new copy of the Options
func (o Options) Copy() Options {
	to := o
	to.ProgressListeners = to.ProgressListeners.Copy()
	return to
}
