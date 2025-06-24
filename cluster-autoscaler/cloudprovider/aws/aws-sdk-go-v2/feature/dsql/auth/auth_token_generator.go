package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	v4 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/signer/v4"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/internal/sdk"
)

const (
	vendorCode       = "dsql"
	emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	userAction       = "DbConnect"
	adminUserAction  = "DbConnectAdmin"
)

// TokenOptions is the optional set of configuration properties for AuthToken
type TokenOptions struct {
	ExpiresIn time.Duration
}

// GenerateDbConnectAuthToken generates an authentication token for IAM authentication to a DSQL database
//
// This is the regular user variant, see [GenerateDBConnectAdminAuthToken] for the admin variant
//
// * endpoint - Endpoint is the hostname to connect to the database
// * region - Region is where the database is located
// * creds - Credentials to use when signing the token
func GenerateDbConnectAuthToken(ctx context.Context, endpoint, region string, creds aws.CredentialsProvider, optFns ...func(options *TokenOptions)) (string, error) {
	values := url.Values{
		"Action": []string{userAction},
	}
	return generateAuthToken(ctx, endpoint, region, values, vendorCode, creds, optFns...)
}

// GenerateDBConnectAdminAuthToken Generates an admin authentication token for IAM authentication to a DSQL database.
//
// This is the admin user variant, see [GenerateDbConnectAuthToken] for the regular user variant
//
// * endpoint - Endpoint is the hostname to connect to the database
// * region - Region is where the database is located
// * creds - Credentials to use when signing the token
func GenerateDBConnectAdminAuthToken(ctx context.Context, endpoint, region string, creds aws.CredentialsProvider, optFns ...func(options *TokenOptions)) (string, error) {
	values := url.Values{
		"Action": []string{adminUserAction},
	}
	return generateAuthToken(ctx, endpoint, region, values, vendorCode, creds, optFns...)
}

// All generate token functions are presigned URLs behind the scenes with the scheme stripped.
// This function abstracts generating this for all use cases
func generateAuthToken(ctx context.Context, endpoint, region string, values url.Values, signingID string, creds aws.CredentialsProvider, optFns ...func(options *TokenOptions)) (string, error) {
	if len(region) == 0 {
		return "", fmt.Errorf("region is required")
	}
	if len(endpoint) == 0 {
		return "", fmt.Errorf("endpoint is required")
	}

	o := TokenOptions{}

	for _, fn := range optFns {
		fn(&o)
	}

	if o.ExpiresIn == 0 {
		o.ExpiresIn = 15 * time.Minute
	}

	if creds == nil {
		return "", fmt.Errorf("credetials provider must not ne nil")
	}

	// the scheme is arbitrary and is only needed because validation of the URL requires one.
	if !(strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://")) {
		endpoint = "https://" + endpoint
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = values.Encode()
	signer := v4.NewSigner()

	credentials, err := creds.Retrieve(ctx)
	if err != nil {
		return "", err
	}

	expires := o.ExpiresIn
	// if credentials expire before expiresIn, set that as the expiration time
	if credentials.CanExpire && !credentials.Expires.IsZero() {
		credsExpireIn := credentials.Expires.Sub(sdk.NowTime())
		expires = min(o.ExpiresIn, credsExpireIn)
	}
	query := req.URL.Query()
	query.Set("X-Amz-Expires", strconv.Itoa(int(expires.Seconds())))
	req.URL.RawQuery = query.Encode()

	signedURI, _, err := signer.PresignHTTP(ctx, credentials, req, emptyPayloadHash, signingID, region, sdk.NowTime().UTC())
	if err != nil {
		return "", err
	}

	url := signedURI
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	} else if strings.HasPrefix(url, "https://") {
		url = url[len("https://"):]
	}

	return url, nil
}
