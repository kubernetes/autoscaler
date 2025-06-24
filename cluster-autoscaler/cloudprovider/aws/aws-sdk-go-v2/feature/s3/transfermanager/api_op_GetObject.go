package transfermanager

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3"
	s3types "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3/types"
	smithymiddleware "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

type errReadingBody struct {
	err error
}

func (e *errReadingBody) Error() string {
	return fmt.Sprintf("failed to read part body: %v", e.err)
}

type errInvalidRange struct {
	max int64
}

func (e *errInvalidRange) Error() string {
	return fmt.Sprintf("invalid input range, must be between 0 and %d", e.max)
}

// GetObjectInput represents a request to the GetObject() or DownloadObject() call. It contains common fields
// of s3 GetObject input
type GetObjectInput struct {
	// Bucket where the object is downloaded from
	Bucket string

	// Key of the object to get.
	Key string

	// To retrieve the checksum, this mode must be enabled.
	//
	// General purpose buckets - In addition, if you enable checksum mode and the
	// object is uploaded with a [checksum] and encrypted with an Key Management Service (KMS)
	// key, you must have permission to use the kms:Decrypt action to retrieve the
	// checksum.
	//
	// [checksum]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_Checksum.html
	ChecksumMode types.ChecksumMode

	// The account ID of the expected bucket owner. If the account ID that you provide
	// does not match the actual owner of the bucket, the request fails with the HTTP
	// status code 403 Forbidden (access denied).
	ExpectedBucketOwner string

	// Return the object only if its entity tag (ETag) is the same as the one
	// specified in this header; otherwise, return a 412 Precondition Failed error.
	//
	// If both of the If-Match and If-Unmodified-Since headers are present in the
	// request as follows: If-Match condition evaluates to true , and;
	// If-Unmodified-Since condition evaluates to false ; then, S3 returns 200 OK and
	// the data requested.
	//
	// For more information about conditional requests, see [RFC 7232].
	//
	// [RFC 7232]: https://tools.ietf.org/html/rfc7232
	IfMatch string

	// Return the object only if it has been modified since the specified time;
	// otherwise, return a 304 Not Modified error.
	//
	// If both of the If-None-Match and If-Modified-Since headers are present in the
	// request as follows: If-None-Match condition evaluates to false , and;
	// If-Modified-Since condition evaluates to true ; then, S3 returns 304 Not
	// Modified status code.
	//
	// For more information about conditional requests, see [RFC 7232].
	//
	// [RFC 7232]: https://tools.ietf.org/html/rfc7232
	IfModifiedSince time.Time

	// Return the object only if its entity tag (ETag) is different from the one
	// specified in this header; otherwise, return a 304 Not Modified error.
	//
	// If both of the If-None-Match and If-Modified-Since headers are present in the
	// request as follows: If-None-Match condition evaluates to false , and;
	// If-Modified-Since condition evaluates to true ; then, S3 returns 304 Not
	// Modified HTTP status code.
	//
	// For more information about conditional requests, see [RFC 7232].
	//
	// [RFC 7232]: https://tools.ietf.org/html/rfc7232
	IfNoneMatch string

	// Return the object only if it has not been modified since the specified time;
	// otherwise, return a 412 Precondition Failed error.
	//
	// If both of the If-Match and If-Unmodified-Since headers are present in the
	// request as follows: If-Match condition evaluates to true , and;
	// If-Unmodified-Since condition evaluates to false ; then, S3 returns 200 OK and
	// the data requested.
	//
	// For more information about conditional requests, see [RFC 7232].
	//
	// [RFC 7232]: https://tools.ietf.org/html/rfc7232
	IfUnmodifiedSince time.Time

	// Part number of the object being read. This is a positive integer between 1 and
	// 10,000. Effectively performs a 'ranged' GET request for the part specified.
	// Useful for downloading just a part of an object.
	PartNumber int32

	// Downloads the specified byte range of an object. For more information about the
	// HTTP Range header, see [https://www.rfc-editor.org/rfc/rfc9110.html#name-range].
	//
	// Amazon S3 doesn't support retrieving multiple ranges of data per GET request.
	//
	// [https://www.rfc-editor.org/rfc/rfc9110.html#name-range]: https://www.rfc-editor.org/rfc/rfc9110.html#name-range
	Range string

	// Confirms that the requester knows that they will be charged for the request.
	// Bucket owners need not specify this parameter in their requests. If either the
	// source or destination S3 bucket has Requester Pays enabled, the requester will
	// pay for corresponding charges to copy the object. For information about
	// downloading objects from Requester Pays buckets, see [Downloading Objects in Requester Pays Buckets]in the Amazon S3 User
	// Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Downloading Objects in Requester Pays Buckets]: https://docs.aws.amazon.com/AmazonS3/latest/dev/ObjectsinRequesterPaysBuckets.html
	RequestPayer types.RequestPayer

	// Sets the Cache-Control header of the response.
	ResponseCacheControl string

	// Sets the Content-Disposition header of the response.
	ResponseContentDisposition string

	// Sets the Content-Encoding header of the response.
	ResponseContentEncoding string

	// Sets the Content-Language header of the response.
	ResponseContentLanguage string

	// Sets the Content-Type header of the response.
	ResponseContentType string

	// Sets the Expires header of the response.
	ResponseExpires time.Time

	// Specifies the algorithm to use when decrypting the object (for example, AES256 ).
	//
	// If you encrypt an object by using server-side encryption with customer-provided
	// encryption keys (SSE-C) when you store the object in Amazon S3, then when you
	// GET the object, you must use the following headers:
	//
	//   - x-amz-server-side-encryption-customer-algorithm
	//
	//   - x-amz-server-side-encryption-customer-key
	//
	//   - x-amz-server-side-encryption-customer-key-MD5
	//
	// For more information about SSE-C, see [Server-Side Encryption (Using Customer-Provided Encryption Keys)] in the Amazon S3 User Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Server-Side Encryption (Using Customer-Provided Encryption Keys)]: https://docs.aws.amazon.com/AmazonS3/latest/dev/ServerSideEncryptionCustomerKeys.html
	SSECustomerAlgorithm string

	// Specifies the customer-provided encryption key that you originally provided for
	// Amazon S3 to encrypt the data before storing it. This value is used to decrypt
	// the object when recovering it and must match the one used when storing the data.
	// The key must be appropriate for use with the algorithm specified in the
	// x-amz-server-side-encryption-customer-algorithm header.
	//
	// If you encrypt an object by using server-side encryption with customer-provided
	// encryption keys (SSE-C) when you store the object in Amazon S3, then when you
	// GET the object, you must use the following headers:
	//
	//   - x-amz-server-side-encryption-customer-algorithm
	//
	//   - x-amz-server-side-encryption-customer-key
	//
	//   - x-amz-server-side-encryption-customer-key-MD5
	//
	// For more information about SSE-C, see [Server-Side Encryption (Using Customer-Provided Encryption Keys)] in the Amazon S3 User Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Server-Side Encryption (Using Customer-Provided Encryption Keys)]: https://docs.aws.amazon.com/AmazonS3/latest/dev/ServerSideEncryptionCustomerKeys.html
	SSECustomerKey string

	// Specifies the 128-bit MD5 digest of the customer-provided encryption key
	// according to RFC 1321. Amazon S3 uses this header for a message integrity check
	// to ensure that the encryption key was transmitted without error.
	//
	// If you encrypt an object by using server-side encryption with customer-provided
	// encryption keys (SSE-C) when you store the object in Amazon S3, then when you
	// GET the object, you must use the following headers:
	//
	//   - x-amz-server-side-encryption-customer-algorithm
	//
	//   - x-amz-server-side-encryption-customer-key
	//
	//   - x-amz-server-side-encryption-customer-key-MD5
	//
	// For more information about SSE-C, see [Server-Side Encryption (Using Customer-Provided Encryption Keys)] in the Amazon S3 User Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Server-Side Encryption (Using Customer-Provided Encryption Keys)]: https://docs.aws.amazon.com/AmazonS3/latest/dev/ServerSideEncryptionCustomerKeys.html
	SSECustomerKeyMD5 string

	// Version ID used to reference a specific version of the object.
	//
	// By default, the GetObject operation returns the current version of an object.
	// To return a different version, use the versionId subresource.
	//
	//   - If you include a versionId in your request header, you must have the
	//   s3:GetObjectVersion permission to access a specific version of an object. The
	//   s3:GetObject permission is not required in this scenario.
	//
	//   - If you request the current version of an object without a specific versionId
	//   in the request header, only the s3:GetObject permission is required. The
	//   s3:GetObjectVersion permission is not required in this scenario.
	//
	//   - Directory buckets - S3 Versioning isn't enabled and supported for directory
	//   buckets. For this API operation, only the null value of the version ID is
	//   supported by directory buckets. You can only specify null to the versionId
	//   query parameter in the request.
	//
	// For more information about versioning, see [PutBucketVersioning].
	//
	// [PutBucketVersioning]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketVersioning.html
	VersionID string
}

func (i GetObjectInput) mapGetObjectInput(enableChecksumValidation bool) *s3.GetObjectInput {
	input := &s3.GetObjectInput{
		Bucket: aws.String(i.Bucket),
		Key:    aws.String(i.Key),
	}

	if i.ChecksumMode != "" {
		input.ChecksumMode = s3types.ChecksumMode(i.ChecksumMode)
	} else if enableChecksumValidation {
		input.ChecksumMode = s3types.ChecksumModeEnabled
	}

	if i.RequestPayer != "" {
		input.RequestPayer = s3types.RequestPayer(i.RequestPayer)
	}

	input.ExpectedBucketOwner = nzstring(i.ExpectedBucketOwner)
	input.IfMatch = nzstring(i.IfMatch)
	input.IfNoneMatch = nzstring(i.IfNoneMatch)
	input.IfModifiedSince = nztime(i.IfModifiedSince)
	input.IfUnmodifiedSince = nztime(i.IfUnmodifiedSince)
	input.ResponseCacheControl = nzstring(i.ResponseCacheControl)
	input.ResponseContentDisposition = nzstring(i.ResponseContentDisposition)
	input.ResponseContentEncoding = nzstring(i.ResponseContentEncoding)
	input.ResponseContentLanguage = nzstring(i.ResponseContentLanguage)
	input.ResponseContentType = nzstring(i.ResponseContentType)
	input.ResponseExpires = nztime(i.ResponseExpires)
	input.SSECustomerAlgorithm = nzstring(i.SSECustomerAlgorithm)
	input.SSECustomerKey = nzstring(i.SSECustomerKey)
	input.SSECustomerKeyMD5 = nzstring(i.SSECustomerKeyMD5)
	input.VersionId = nzstring(i.VersionID)

	return input
}

// GetObjectOutput represents a response from GetObject() or DownloadObject() call. It contains common fields
// of s3 GetObject output
type GetObjectOutput struct {
	// Indicates that a range of bytes was specified in the request.
	AcceptRanges string

	// Object data.
	Body io.Reader

	// Indicates whether the object uses an S3 Bucket Key for server-side encryption
	// with Key Management Service (KMS) keys (SSE-KMS).
	BucketKeyEnabled bool

	// Specifies caching behavior along the request/reply chain.
	CacheControl string

	// Specifies if the response checksum validation is enabled
	ChecksumMode types.ChecksumMode

	// The base64-encoded, 32-bit CRC-32 checksum of the object. This will only be
	// present if it was uploaded with the object. For more information, see [Checking object integrity]in the
	// Amazon S3 User Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	ChecksumCRC32 string

	// The base64-encoded, 32-bit CRC-32C checksum of the object. This will only be
	// present if it was uploaded with the object. For more information, see [Checking object integrity]in the
	// Amazon S3 User Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	ChecksumCRC32C string

	// The base64-encoded, 160-bit SHA-1 digest of the object. This will only be
	// present if it was uploaded with the object. For more information, see [Checking object integrity]in the
	// Amazon S3 User Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	ChecksumSHA1 string

	// The base64-encoded, 256-bit SHA-256 digest of the object. This will only be
	// present if it was uploaded with the object. For more information, see [Checking object integrity]in the
	// Amazon S3 User Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	ChecksumSHA256 string

	// Specifies presentational information for the object.
	ContentDisposition string

	// Indicates what content encodings have been applied to the object and thus what
	// decoding mechanisms must be applied to obtain the media-type referenced by the
	// Content-Type header field.
	ContentEncoding string

	// The language the content is in.
	ContentLanguage string

	// Size of the body in bytes.
	ContentLength int64

	// The portion of the object returned in the response.
	ContentRange string

	// A standard MIME type describing the format of the object data.
	ContentType string

	// Indicates whether the object retrieved was (true) or was not (false) a Delete
	// Marker. If false, this response header does not appear in the response.
	//
	//   - If the current version of the object is a delete marker, Amazon S3 behaves
	//   as if the object was deleted and includes x-amz-delete-marker: true in the
	//   response.
	//
	//   - If the specified version in the request is a delete marker, the response
	//   returns a 405 Method Not Allowed error and the Last-Modified: timestamp
	//   response header.
	DeleteMarker bool

	// An entity tag (ETag) is an opaque identifier assigned by a web server to a
	// specific version of a resource found at a URL.
	ETag string

	// If the object expiration is configured (see [PutBucketLifecycleConfiguration]PutBucketLifecycleConfiguration ),
	// the response includes this header. It includes the expiry-date and rule-id
	// key-value pairs providing object expiration information. The value of the
	// rule-id is URL-encoded.
	//
	// This functionality is not supported for directory buckets.
	//
	// [PutBucketLifecycleConfiguration]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html
	Expiration string

	// The date and time at which the object is no longer cacheable.
	//
	// Deprecated: This field is handled inconsistently across AWS SDKs. Prefer using
	// the ExpiresString field which contains the unparsed value from the service
	// response.
	Expires time.Time

	// The unparsed value of the Expires field from the service response. Prefer use
	// of this value over the normal Expires response field where possible.
	ExpiresString string

	// Date and time when the object was last modified.
	//
	// General purpose buckets - When you specify a versionId of the object in your
	// request, if the specified version in the request is a delete marker, the
	// response returns a 405 Method Not Allowed error and the Last-Modified: timestamp
	// response header.
	LastModified time.Time

	// A map of metadata to store with the object in S3.
	//
	// Map keys will be normalized to lower-case.
	Metadata map[string]string

	// This is set to the number of metadata entries not returned in the headers that
	// are prefixed with x-amz-meta- . This can happen if you create metadata using an
	// API like SOAP that supports more flexible metadata than the REST API. For
	// example, using SOAP, you can create metadata whose values are not legal HTTP
	// headers.
	//
	// This functionality is not supported for directory buckets.
	MissingMeta int32

	// Indicates whether this object has an active legal hold. This field is only
	// returned if you have permission to view an object's legal hold status.
	//
	// This functionality is not supported for directory buckets.
	ObjectLockLegalHoldStatus types.ObjectLockLegalHoldStatus

	// The Object Lock mode that's currently in place for this object.
	//
	// This functionality is not supported for directory buckets.
	ObjectLockMode types.ObjectLockMode

	// The date and time when this object's Object Lock will expire.
	//
	// This functionality is not supported for directory buckets.
	ObjectLockRetainUntilDate time.Time

	// The count of parts this object has. This value is only returned if you specify
	// partNumber in your request and the object was uploaded as a multipart upload.
	PartsCount int32

	// Amazon S3 can return this if your request involves a bucket that is either a
	// source or destination in a replication rule.
	//
	// This functionality is not supported for directory buckets.
	ReplicationStatus types.ReplicationStatus

	// If present, indicates that the requester was successfully charged for the
	// request.
	//
	// This functionality is not supported for directory buckets.
	RequestCharged types.RequestCharged

	// Provides information about object restoration action and expiration time of the
	// restored object copy.
	//
	// This functionality is not supported for directory buckets. Only the S3 Express
	// One Zone storage class is supported by directory buckets to store objects.
	Restore string

	// If server-side encryption with a customer-provided encryption key was
	// requested, the response will include this header to confirm the encryption
	// algorithm that's used.
	//
	// This functionality is not supported for directory buckets.
	SSECustomerAlgorithm string

	// If server-side encryption with a customer-provided encryption key was
	// requested, the response will include this header to provide the round-trip
	// message integrity verification of the customer-provided encryption key.
	//
	// This functionality is not supported for directory buckets.
	SSECustomerKeyMD5 string

	// If present, indicates the ID of the KMS key that was used for object encryption.
	SSEKMSKeyID string

	// The server-side encryption algorithm used when you store this object in Amazon
	// S3.
	ServerSideEncryption types.ServerSideEncryption

	// Provides storage class information of the object. Amazon S3 returns this header
	// for all objects except for S3 Standard storage class objects.
	//
	// Directory buckets - Only the S3 Express One Zone storage class is supported by
	// directory buckets to store objects.
	StorageClass types.StorageClass

	// The number of tags, if any, on the object, when you have the relevant
	// permission to read object tags.
	//
	// You can use [GetObjectTagging] to retrieve the tag set associated with an object.
	//
	// This functionality is not supported for directory buckets.
	//
	// [GetObjectTagging]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectTagging.html
	TagCount int32

	// Version ID of the object.
	//
	// This functionality is not supported for directory buckets.
	VersionID string

	// If the bucket is configured as a website, redirects requests for this object to
	// another object in the same bucket or to an external URL. Amazon S3 stores the
	// value of this header in the object metadata.
	//
	// This functionality is not supported for directory buckets.
	WebsiteRedirectLocation string

	// Metadata pertaining to the operation's result.
	ResultMetadata smithymiddleware.Metadata
}

func (o *GetObjectOutput) mapFromGetObjectOutput(out *s3.GetObjectOutput, checksumMode s3types.ChecksumMode) {
	o.AcceptRanges = aws.ToString(out.AcceptRanges)
	o.CacheControl = aws.ToString(out.CacheControl)
	o.ChecksumMode = types.ChecksumMode(checksumMode)
	o.ChecksumCRC32 = aws.ToString(out.ChecksumCRC32)
	o.ChecksumCRC32C = aws.ToString(out.ChecksumCRC32C)
	o.ChecksumSHA1 = aws.ToString(out.ChecksumSHA1)
	o.ChecksumSHA256 = aws.ToString(out.ChecksumSHA256)
	o.ContentDisposition = aws.ToString(out.ContentDisposition)
	o.ContentEncoding = aws.ToString(out.ContentEncoding)
	o.ContentLanguage = aws.ToString(out.ContentLanguage)
	o.ContentRange = aws.ToString(out.ContentRange)
	o.ContentType = aws.ToString(out.ContentType)
	o.ETag = aws.ToString(out.ETag)
	o.Expiration = aws.ToString(out.Expiration)
	o.ExpiresString = aws.ToString(out.ExpiresString)
	o.Restore = aws.ToString(out.Restore)
	o.SSECustomerAlgorithm = aws.ToString(out.SSECustomerAlgorithm)
	o.SSECustomerKeyMD5 = aws.ToString(out.SSECustomerKeyMD5)
	o.SSEKMSKeyID = aws.ToString(out.SSEKMSKeyId)
	o.VersionID = aws.ToString(out.VersionId)
	o.WebsiteRedirectLocation = aws.ToString(out.WebsiteRedirectLocation)
	o.BucketKeyEnabled = aws.ToBool(out.BucketKeyEnabled)
	o.DeleteMarker = aws.ToBool(out.DeleteMarker)
	o.MissingMeta = aws.ToInt32(out.MissingMeta)
	o.PartsCount = aws.ToInt32(out.PartsCount)
	o.TagCount = aws.ToInt32(out.TagCount)
	o.ContentLength = aws.ToInt64(out.ContentLength)
	o.Body = out.Body
	o.Expires = aws.ToTime(out.Expires)
	o.LastModified = aws.ToTime(out.LastModified)
	o.ObjectLockRetainUntilDate = aws.ToTime(out.ObjectLockRetainUntilDate)
	o.Metadata = out.Metadata
	o.ObjectLockLegalHoldStatus = types.ObjectLockLegalHoldStatus(out.ObjectLockLegalHoldStatus)
	o.ObjectLockMode = types.ObjectLockMode(out.ObjectLockMode)
	o.ReplicationStatus = types.ReplicationStatus(out.ReplicationStatus)
	o.RequestCharged = types.RequestCharged(out.RequestCharged)
	o.ServerSideEncryption = types.ServerSideEncryption(out.ServerSideEncryption)
	o.StorageClass = types.StorageClass(out.StorageClass)
	o.ResultMetadata = out.ResultMetadata.Clone()
}

func (o *GetObjectOutput) mapFromHeadObjectOutput(out *s3.HeadObjectOutput, checksumMode types.ChecksumMode, enableChecksumValidation bool, body *concurrentReader) {
	o.AcceptRanges = aws.ToString(out.AcceptRanges)
	o.CacheControl = aws.ToString(out.CacheControl)
	if checksumMode != "" {
		o.ChecksumMode = checksumMode
	} else if enableChecksumValidation {
		o.ChecksumMode = types.ChecksumModeEnabled
	}
	o.ChecksumCRC32 = aws.ToString(out.ChecksumCRC32)
	o.ChecksumCRC32C = aws.ToString(out.ChecksumCRC32C)
	o.ChecksumSHA1 = aws.ToString(out.ChecksumSHA1)
	o.ChecksumSHA256 = aws.ToString(out.ChecksumSHA256)
	o.ContentDisposition = aws.ToString(out.ContentDisposition)
	o.ContentEncoding = aws.ToString(out.ContentEncoding)
	o.ContentLanguage = aws.ToString(out.ContentLanguage)
	o.ContentType = aws.ToString(out.ContentType)
	o.ETag = aws.ToString(out.ETag)
	o.Expiration = aws.ToString(out.Expiration)
	o.ExpiresString = aws.ToString(out.ExpiresString)
	o.Restore = aws.ToString(out.Restore)
	o.SSECustomerAlgorithm = aws.ToString(out.SSECustomerAlgorithm)
	o.SSECustomerKeyMD5 = aws.ToString(out.SSECustomerKeyMD5)
	o.SSEKMSKeyID = aws.ToString(out.SSEKMSKeyId)
	o.VersionID = aws.ToString(out.VersionId)
	o.WebsiteRedirectLocation = aws.ToString(out.WebsiteRedirectLocation)
	o.BucketKeyEnabled = aws.ToBool(out.BucketKeyEnabled)
	o.DeleteMarker = aws.ToBool(out.DeleteMarker)
	o.MissingMeta = aws.ToInt32(out.MissingMeta)
	o.PartsCount = aws.ToInt32(out.PartsCount)
	o.ContentLength = getTotalBytes(out)
	o.Body = body
	o.Expires = aws.ToTime(out.Expires)
	o.LastModified = aws.ToTime(out.LastModified)
	o.ObjectLockRetainUntilDate = aws.ToTime(out.ObjectLockRetainUntilDate)
	o.Metadata = out.Metadata
	o.ObjectLockLegalHoldStatus = types.ObjectLockLegalHoldStatus(out.ObjectLockLegalHoldStatus)
	o.ObjectLockMode = types.ObjectLockMode(out.ObjectLockMode)
	o.ReplicationStatus = types.ReplicationStatus(out.ReplicationStatus)
	o.RequestCharged = types.RequestCharged(out.RequestCharged)
	o.ServerSideEncryption = types.ServerSideEncryption(out.ServerSideEncryption)
	o.StorageClass = types.StorageClass(out.StorageClass)
	o.ResultMetadata = out.ResultMetadata.Clone()
}

// GetObject downloads an object from S3, intelligently splitting large
// files into smaller parts/ranges according to config and getting them in parallel across
// multiple goroutines. You can configure the download type, chunk size and concurrency
// through the Options parameters.
//
// Additional functional options can be provided to configure the individual
// download. These options are copies of the original Options instance, the client of which GetObject is called from.
// Modifying the options will not impact the original Client and Options instance.
func (c *Client) GetObject(ctx context.Context, input *GetObjectInput, opts ...func(*Options)) (*GetObjectOutput, error) {
	i := getter{in: input, options: c.options.Copy()}
	for _, opt := range opts {
		opt(&i.options)
	}

	return i.get(ctx)
}

type getter struct {
	options Options
	in      *GetObjectInput
}

func (g *getter) get(ctx context.Context) (out *GetObjectOutput, err error) {
	if err := g.init(); err != nil {
		return nil, fmt.Errorf("unable to initialize download: %w", err)
	}

	clientOptions := []func(*s3.Options){
		func(o *s3.Options) {
			o.APIOptions = append(o.APIOptions,
				middleware.AddSDKAgentKey(middleware.FeatureMetadata, userAgentKey),
				addFeatureUserAgent,
			)
		}}

	if g.in.PartNumber > 0 {
		return g.singleDownload(ctx, clientOptions...)
	}

	r := &concurrentReader{
		ctx:      ctx,
		buf:      make(map[int32]*outChunk),
		partSize: 1,
		options:  g.options.Copy(),
		in:       g.in,
		ch:       make(chan outChunk, g.options.Concurrency),
	}

	output := &GetObjectOutput{}
	if g.options.GetObjectType == types.GetObjectParts {
		if g.in.Range != "" {
			return g.singleDownload(ctx, clientOptions...)
		}
		// must know the part size before creating stream reader
		out, err := g.options.S3.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket:     aws.String(g.in.Bucket),
			Key:        aws.String(g.in.Key),
			PartNumber: aws.Int32(1),
		}, clientOptions...)
		if err != nil {
			return nil, err
		}

		output.mapFromHeadObjectOutput(out, g.in.ChecksumMode, !g.options.DisableChecksumValidation, r)
		contentLength := getTotalBytes(out)
		output.ContentRange = fmt.Sprintf("bytes=0-%d/%d", contentLength-1, contentLength)

		partsCount := max(aws.ToInt32(out.PartsCount), 1)
		partSize := max(aws.ToInt64(out.ContentLength), 1)
		sectionParts := int32(max(1, g.options.GetBufferSize/partSize))
		capacity := sectionParts
		r.sectionParts = sectionParts
		r.partSize = partSize
		r.setCapacity(min(capacity, partsCount))
		r.partsCount = partsCount
	} else {
		out, err := g.options.S3.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(g.in.Bucket),
			Key:    aws.String(g.in.Key),
		}, clientOptions...)
		if err != nil {
			return nil, err
		}
		if aws.ToInt64(out.ContentLength) == 0 {
			return g.singleDownload(ctx, clientOptions...)
		}
		total := aws.ToInt64(out.ContentLength)
		var pos int64
		if g.in.Range != "" {
			start, totalBytes, err := g.getDownloadRange()
			if err != nil || start < 0 || start >= total || totalBytes > total || start >= totalBytes {
				return nil, &errInvalidRange{
					max: total - 1,
				}
			}
			pos = start
			total = totalBytes
		}
		contentLength := total - pos

		output.mapFromHeadObjectOutput(out, g.in.ChecksumMode, !g.options.DisableChecksumValidation, r)
		output.ContentLength = contentLength
		output.ContentRange = fmt.Sprintf("bytes=%d-%d/%d", pos, total-1, aws.ToInt64(out.ContentLength))

		partsCount := int32((contentLength-1)/g.options.PartSizeBytes + 1)
		sectionParts := int32(max(1, g.options.GetBufferSize/g.options.PartSizeBytes))
		capacity := min(sectionParts, partsCount)
		r.partSize = g.options.PartSizeBytes
		r.setCapacity(capacity)
		r.partsCount = partsCount
		r.sectionParts = sectionParts
		r.totalBytes = total
		r.pos = pos
	}

	r.etag = output.ETag
	output.Body = r
	return output, nil
}

func (g *getter) init() error {
	if g.options.PartSizeBytes < minPartSizeBytes {
		return fmt.Errorf("part size must be at least %d bytes", minPartSizeBytes)
	}

	return nil
}

func (g *getter) singleDownload(ctx context.Context, clientOptions ...func(*s3.Options)) (*GetObjectOutput, error) {
	params := g.in.mapGetObjectInput(!g.options.DisableChecksumValidation)
	out, err := g.options.S3.GetObject(ctx, params, clientOptions...)
	if err != nil {
		return nil, err
	}

	output := &GetObjectOutput{}
	output.mapFromGetObjectOutput(out, params.ChecksumMode)
	return output, nil
}

func getTotalBytes(resp *s3.HeadObjectOutput) int64 {
	if resp.ContentRange == nil {
		// ContentRange is nil when the full file contents is provided, and
		// is not chunked. Use ContentLength instead.
		return aws.ToInt64(resp.ContentLength)
	}
	parts := strings.Split(*resp.ContentRange, "/")
	totalStr := parts[len(parts)-1]
	total, err := strconv.ParseInt(totalStr, 10, 64)
	if err != nil {
		return -1
	}
	return total
}

func (g *getter) getDownloadRange() (int64, int64, error) {
	parts := strings.Split(strings.Split(g.in.Range, "=")[1], "-")

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	end, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return start, end + 1, nil
}
