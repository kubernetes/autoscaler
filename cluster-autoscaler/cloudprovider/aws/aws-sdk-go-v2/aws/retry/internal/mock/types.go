package mock

// NoSuchBucketException is a mock error
type NoSuchBucketException struct{}

// Error is a mock error message
func (*NoSuchBucketException) Error() string {
	return "mock error message"
}

// ErrorCode is a mock error code
func (*NoSuchBucketException) ErrorCode() string {
	return "NoSuchBucketException"
}
