// Package sigv4 implements request signing for the basic form AWS Signature
// Version 4.
package sigv4

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/credentials"
	v4internal "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/internal/v4"
	v4 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/v4"
)

const algorithm = "AWS4-HMAC-SHA256"

// Signer signs requests with AWS Signature version 4.
type Signer struct {
	options v4.SignerOptions
}

// New returns an instance of Signer with applied options.
func New(opts ...v4.SignerOption) *Signer {
	options := v4.SignerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return &Signer{options}
}

// SignRequestInput is the set of inputs for Sigv4 signing.
type SignRequestInput struct {
	// The input request, which will modified in-place during signing.
	Request *http.Request

	// The SHA256 hash of the input request body.
	//
	// This value is NOT required to sign the request, but it is recommended to
	// provide it (or provide a Body on the HTTP request that implements
	// io.Seeker such that the signer can calculate it for you). Many services
	// do not accept requests with unsigned payloads.
	//
	// If a value is not provided, and DisableImplicitPayloadHashing has not
	// been set on SignerOptions, the signer will attempt to derive the payload
	// hash itself. The request's Body MUST implement io.Seeker in order to do
	// this, if it does not, the magic value for unsigned payload is used. If
	// the body does implement io.Seeker, but a call to Seek returns an error,
	// the signer will forward that error.
	PayloadHash []byte

	// The identity used to sign the request.
	Credentials credentials.Credentials

	// The service and region for which this request is to be signed.
	//
	// The appropriate values for these fields are determined by the service
	// vendor.
	Service, Region string

	// Wall-clock time used for calculating the signature.
	//
	// If the zero-value is given (generally by the caller not setting it), the
	// signer will instead use the current system clock time for the signature.
	Time time.Time
}

// SignRequest signs an HTTP request with AWS Signature Version 4, modifying
// the request in-place by adding the headers that constitute the signature.
//
// SignRequest will modify the request by setting the following headers:
//   - Host: required in general for HTTP/1.1 as well as for v4-signed requests
//   - X-Amz-Date: required for v4-signed requests
//   - X-Amz-Security-Token: required for v4-signed requests IF present on
//     credentials used to sign, otherwise this header will not be set
//   - Authorization: contains the v4 signature string
//
// The request MUST have a Host value set at the time that this API is called,
// such that it can be included in the signature calculation. Standard library
// HTTP clients set this as a request header by default, meaning that a request
// signed without a Host value will end up transmitting with the Host header
// anyway, which will cause the request to be rejected by the service due to
// signature mismatch (the Host header is required to be signed with Sigv4).
//
// Generally speaking, using http.NewRequest will ensure that request instances
// are sufficiently initialized to be used with this API, though it is not
// strictly required.
//
// SignRequest may be called any number of times on an http.Request instance,
// the header values set as part of the signature will simply be overwritten
// with newer or re-calculated values (such as a new set of credentials with a
// new session token, which would in turn result in a different signature).
func (s *Signer) SignRequest(in *SignRequestInput, opts ...v4.SignerOption) error {
	options := s.options
	for _, opt := range opts {
		opt(&options)
	}

	tm := v4internal.ResolveTime(in.Time)
	signer := v4internal.Signer{
		Request:     in.Request,
		PayloadHash: in.PayloadHash,
		Time:        tm,
		Credentials: in.Credentials,
		Options:     options,

		Algorithm:       algorithm,
		CredentialScope: scope(tm, in.Region, in.Service),
		Finalizer: &finalizer{
			Secret:  in.Credentials.SecretAccessKey,
			Service: in.Service,
			Region:  in.Region,
			Time:    tm,
		},
	}
	if err := signer.Do(); err != nil {
		return err
	}

	return nil
}

func scope(signingTime time.Time, region, service string) string {
	return strings.Join([]string{
		signingTime.Format(v4internal.ShortTimeFormat),
		region,
		service,
		"aws4_request",
	}, "/")
}

type finalizer struct {
	Secret          string
	Service, Region string
	Time            time.Time
}

func (f *finalizer) SignString(toSign string) (string, error) {
	key := hmacSHA256([]byte("AWS4"+f.Secret), []byte(f.Time.Format(v4internal.ShortTimeFormat)))
	key = hmacSHA256(key, []byte(f.Region))
	key = hmacSHA256(key, []byte(f.Service))
	key = hmacSHA256(key, []byte("aws4_request"))

	return hex.EncodeToString(hmacSHA256(key, []byte(toSign))), nil
}

func hmacSHA256(key, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
