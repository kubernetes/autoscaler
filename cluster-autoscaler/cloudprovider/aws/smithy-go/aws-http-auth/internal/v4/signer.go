package v4

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/credentials"
	v4 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/v4"
)

const (
	// TimeFormat is the full-width form to be used in the X-Amz-Date header.
	TimeFormat = "20060102T150405Z"

	// ShortTimeFormat is the shortened form used in credential scope.
	ShortTimeFormat = "20060102"
)

// Signer is the implementation structure for all variants of v4 signing.
type Signer struct {
	Request     *http.Request
	PayloadHash []byte
	Time        time.Time
	Credentials credentials.Credentials
	Options     v4.SignerOptions

	// variant-specific inputs
	Algorithm       string
	CredentialScope string
	Finalizer       Finalizer
}

// Finalizer performs the final step in v4 signing, deriving a signature for
// the string-to-sign with algorithm-specific key material.
type Finalizer interface {
	SignString(string) (string, error)
}

// Do performs v4 signing, modifying the request in-place with the
// signature.
//
// Do should be called exactly once for a configured Signer. The behavior of
// doing otherwise is undefined.
func (s *Signer) Do() error {
	if err := s.init(); err != nil {
		return err
	}

	s.setRequiredHeaders()

	canonicalRequest, signedHeaders := s.buildCanonicalRequest()
	stringToSign := s.buildStringToSign(canonicalRequest)
	signature, err := s.Finalizer.SignString(stringToSign)
	if err != nil {
		return err
	}

	s.Request.Header.Set("Authorization",
		s.buildAuthorizationHeader(signature, signedHeaders))

	return nil
}

func (s *Signer) init() error {
	// it might seem like time should also get defaulted/normalized here, but
	// in practice sigv4 and sigv4a both need to do that beforehand to
	// calculate scope, so there's no point

	if s.Options.HeaderRules == nil {
		s.Options.HeaderRules = defaultHeaderRules{}
	}

	if err := s.resolvePayloadHash(); err != nil {
		return err
	}

	return nil
}

// ensure we have a value for payload hash, whether that be explicit, implicit,
// or the unsigned sentinel
func (s *Signer) resolvePayloadHash() error {
	if len(s.PayloadHash) > 0 {
		return nil
	}

	rs, ok := s.Request.Body.(io.ReadSeeker)
	if !ok || s.Options.DisableImplicitPayloadHashing {
		s.PayloadHash = v4.UnsignedPayload()
		return nil
	}

	p, err := rtosha(rs)
	if err != nil {
		return err
	}

	s.PayloadHash = p
	return nil
}

func (s *Signer) setRequiredHeaders() {
	headers := s.Request.Header

	s.Request.Header.Set("Host", s.Request.Host)
	s.Request.Header.Set("X-Amz-Date", s.Time.Format(TimeFormat))

	if len(s.Credentials.SessionToken) > 0 {
		s.Request.Header.Set("X-Amz-Security-Token", s.Credentials.SessionToken)
	}
	if len(s.PayloadHash) > 0 && s.Options.AddPayloadHashHeader {
		headers.Set("X-Amz-Content-Sha256", payloadHashString(s.PayloadHash))
	}
}

func (s *Signer) buildCanonicalRequest() (string, string) {
	canonPath := s.Request.URL.EscapedPath()
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html:
	// if input has no path, "/" is used
	if len(canonPath) == 0 {
		canonPath = "/"
	}
	if !s.Options.DisableDoublePathEscape {
		canonPath = uriEncode(canonPath)
	}

	query := s.Request.URL.Query()
	for key := range query {
		sort.Strings(query[key])
	}
	canonQuery := strings.Replace(query.Encode(), "+", "%20", -1)

	canonHeaders, signedHeaders := s.buildCanonicalHeaders()

	req := strings.Join([]string{
		s.Request.Method,
		canonPath,
		canonQuery,
		canonHeaders,
		signedHeaders,
		payloadHashString(s.PayloadHash),
	}, "\n")

	return req, signedHeaders
}

func (s *Signer) buildCanonicalHeaders() (canon, signed string) {
	var canonHeaders []string
	signedHeaders := map[string][]string{}

	// step 1: find what we're signing
	for header, values := range s.Request.Header {
		lowercase := strings.ToLower(header)
		if !s.Options.HeaderRules.IsSigned(lowercase) {
			continue
		}

		canonHeaders = append(canonHeaders, lowercase)
		signedHeaders[lowercase] = values
	}
	sort.Strings(canonHeaders)

	// step 2: indexing off of the list we built previously (which guarantees
	// alphabetical order), build the canonical list
	var ch strings.Builder
	for i := range canonHeaders {
		ch.WriteString(canonHeaders[i])
		ch.WriteRune(':')

		// headers can have multiple values
		values := signedHeaders[canonHeaders[i]]
		for j, value := range values {
			ch.WriteString(strings.TrimSpace(value))
			if j < len(values)-1 {
				ch.WriteRune(',')
			}
		}
		ch.WriteRune('\n')
	}

	return ch.String(), strings.Join(canonHeaders, ";")
}

func (s *Signer) buildStringToSign(canonicalRequest string) string {
	return strings.Join([]string{
		s.Algorithm,
		s.Time.Format(TimeFormat),
		s.CredentialScope,
		hex.EncodeToString(Stosha(canonicalRequest)),
	}, "\n")
}

func (s *Signer) buildAuthorizationHeader(signature, headers string) string {
	return fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s",
		s.Algorithm,
		s.Credentials.AccessKeyID+"/"+s.CredentialScope,
		headers,
		signature)
}

func payloadHashString(p []byte) string {
	if string(p) == "UNSIGNED-PAYLOAD" {
		return string(p) // sentinel, do not hex-encode
	}
	return hex.EncodeToString(p)
}

// ResolveTime initializes a time value for signing.
func ResolveTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}

type defaultHeaderRules struct{}

func (defaultHeaderRules) IsSigned(h string) bool {
	return h == "host" || strings.HasPrefix(h, "x-amz-")
}
