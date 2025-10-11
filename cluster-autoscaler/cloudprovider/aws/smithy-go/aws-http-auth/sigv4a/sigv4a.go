// Package sigv4a implements request signing for AWS Signature Version 4a
// (asymmetric).
//
// The algorithm for Signature Version 4a is identical to that of plain v4
// apart from the following:
//   - A request can be signed for multiple regions. This is represented in the
//     signature using the X-Amz-Region-Set header. The credential scope string
//     used in the calculation correspondingly lacks the region component from
//     that of plain v4.
//   - The string-to-sign component of the calculation is instead signed with
//     an ECDSA private key. This private key is typically derived from your
//     regular AWS credentials.
package sigv4a

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/credentials"
	v4internal "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/internal/v4"
	v4 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/v4"
)

const algorithm = "AWS4-ECDSA-P256-SHA256"

// Signer signs requests with AWS Signature Version 4a.
//
// Unlike Sigv4, AWS SigV4a signs requests with an ECDSA private key. This is
// derived automatically from the AWS credential identity passed to
// SignRequest. This derivation result is cached on the Signer and is uniquely
// identified by the access key ID (AKID) of the credentials that were
// provided.
//
// Because of this, the caller is encouraged to create multiple instances of
// Signer for different underlying identities (e.g. IAM roles).
type Signer struct {
	options v4.SignerOptions

	// derived asymmetric credentials
	privCache *ecdsaCache
}

// New returns an instance of Signer with applied options.
func New(opts ...v4.SignerOption) *Signer {
	options := v4.SignerOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	return &Signer{
		options:   options,
		privCache: &ecdsaCache{},
	}
}

// SignRequestInput is the set of inputs for the Sigv4a signing process.
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

	// The service for which this request is to be signed.
	//
	// The appropriate value for this field is determined by the service
	// vendor.
	Service string

	// The set of regions for which this request is to be signed.
	//
	// The sentinel {"*"} is used to indicate that the signed request is valid
	// in all regions. Callers MUST set a value for this field - the API will
	// not fill in a default and the resulting signature will ultimately be
	// invalid.
	//
	// The acceptable values for list entries of this field are determined by
	// the service vendor.
	RegionSet []string

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
//   - X-Amz-Date: required for v4a-signed requests
//   - X-Amz-Region-Set: used to convey the regions for which the request is
//     signed in v4a
//   - X-Amz-Security-Token: required for v4a-signed requests IF present on
//     credentials used to sign, otherwise this header will not be set
//   - Authorization: contains the v4a signature string
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
// the header values set as part of the signature will simply be re-calculated.
// Note that v4a signatures are non-deterministic due to the random component
// of ECDSA signing, callers should not expect two calls to SignRequest() to
// produce an identical signature.
func (s *Signer) SignRequest(in *SignRequestInput, opts ...v4.SignerOption) error {
	options := s.options
	for _, fn := range opts {
		fn(&options)
	}

	priv, err := s.privCache.Derive(in.Credentials)
	if err != nil {
		return err
	}

	in.Request.Header.Set("X-Amz-Region-Set", strings.Join(in.RegionSet, ","))

	tm := v4internal.ResolveTime(in.Time)
	signer := &v4internal.Signer{
		Request:     in.Request,
		PayloadHash: in.PayloadHash,
		Time:        tm,
		Credentials: in.Credentials,
		Options:     options,

		Algorithm:       algorithm,
		CredentialScope: scope(tm, in.Service),
		Finalizer:       &finalizer{priv},
	}
	if err := signer.Do(); err != nil {
		return err
	}

	return nil
}

func scope(tm time.Time, service string) string {
	return strings.Join([]string{
		tm.Format(v4internal.ShortTimeFormat),
		service,
		"aws4_request",
	}, "/")
}

type finalizer struct {
	Secret *ecdsa.PrivateKey
}

func (f *finalizer) SignString(strToSign string) (string, error) {
	sig, err := f.Secret.Sign(rand.Reader, v4internal.Stosha(strToSign), crypto.SHA256)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}
