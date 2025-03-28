/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	gorecorder "gopkg.in/dnaeon/go-vcr.v3/recorder"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

var requestHeadersToRemove = []string{
	// remove all Authorization headers from stored requests
	"Authorization",

	// Not needed, adds to diff churn:
	"User-Agent",
}

var responseHeadersToRemove = []string{
	// Request IDs
	"X-Ms-Arm-Service-Request-Id",
	"X-Ms-Correlation-Request-Id",
	"X-Ms-Request-Id",
	"X-Ms-Ests-Server",
	"X-Ms-Routing-Request-Id",
	"X-Ms-Client-Request-Id",
	"Client-Request-Id",

	// Quota limits
	"X-Ms-Ratelimit-Remaining-Subscription-Deletes",
	"X-Ms-Ratelimit-Remaining-Subscription-Reads",
	"X-Ms-Ratelimit-Remaining-Subscription-Writes",

	// Not needed, adds to diff churn
	"Date",
	"Set-Cookie",
	// Causes client to delay long time
	"Retry-After",

	"Content-Security-Policy-Report-Only",
	"X-Msedge-Ref",
}

var (
	dateMatcher   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|((\+|\-)\d{2}(:?\d{2})?(:?\d{2})?))?`)
	sshKeyMatcher = regexp.MustCompile("ssh-rsa [0-9a-zA-Z+/=]+")

	// This is pretty involved, here's the breakdown of what each bit means:
	// [p|P]assword":\s*" - find any JSON field that ends in the string password, followed by any number of spaces and another quote.
	// ((?:[^\\"]*?(?:(\\\\)|(\\"))*?)*?) - The outer group is a capturing group, which selects the actual password
	// [^\\"]*? - this matches any characters that aren't \ or " (need to handle them specially because of escaped quotes)
	// (?:(?:\\\\)|(?:\\")|(?:\\))*? - lazily match any number of escaped backslahes or escaped quotes.
	// The above two sections are repeated until the first unescaped "s
	passwordMatcher = regexp.MustCompile(`[p|P]assword":\s*"((?:[^\\"]*?(?:(?:\\\\)|(?:\\")|(?:\\))*?)*?)"`)

	// keyMatcher matches any valid base64 value with at least 10 sets of 4 bytes of data that ends in = or ==.
	// Both storage account keys and Redis account keys are longer than that and end in = or ==. Note that technically
	// base64 values need not end in == or =, but allowing for that in the match will flag tons of false positives as
	// any text (including long URLs) have strings of characters that meet this requirement. There are other base64 values
	// in the payloads (such as operationResults URLs for polling async operations for some services) that seem to use
	// very long base64 strings as well.
	keyMatcher = regexp.MustCompile("(?:[A-Za-z0-9+/]{4}){10,}(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)")

	// uuidMatcher matches any valid UUID
	uuidMatcher = regexp.MustCompile("[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}")
)

// hideDates replaces all ISO8601 datetimes with a fixed value
// this lets us match requests that may contain time-sensitive information (timestamps, etc)
func hideDates(s string) string {
	return dateMatcher.ReplaceAllLiteralString(s, "2001-02-03T04:05:06Z") // this should be recognizable/parseable as a fake date
}

// hideSSHKeys hides anything that looks like SSH keys
func hideSSHKeys(s string) string {
	return sshKeyMatcher.ReplaceAllLiteralString(s, "ssh-rsa {KEY}")
}

// hideuuid hides uuid
func hideUUID(s string) string {
	return uuidMatcher.ReplaceAllLiteralString(s, uuid.Nil.String())
}

// hidePasswords hides anything that looks like a generated password
func hidePasswords(s string) string {
	matches := passwordMatcher.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		for n, submatch := range match {
			if n%2 == 0 {
				continue
			}
			s = strings.ReplaceAll(s, submatch, "{PASSWORD}")
		}
	}
	return s
}

func hideKeys(s string) string {
	return keyMatcher.ReplaceAllLiteralString(s, "{KEY}")
}

func hideRecordingData(s string) string {
	result := hideDates(s)
	result = hideSSHKeys(result)
	result = hidePasswords(result)
	result = hideKeys(result)
	result = hideUUID(result)
	return result
}

type Recorder struct {
	credential       azcore.TokenCredential
	rec              *gorecorder.Recorder
	subscriptionID   string
	tenantID         string
	clientID         string
	clientSecret     string
	clientCertPath   string
	clientCertPasswd string
}

type DummyTokenCredential func(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error)

func (d DummyTokenCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return d(ctx, opts)
}

func NewRecorder(cassetteName string) (*Recorder, error) {

	rec, err := gorecorder.NewWithOptions(&gorecorder.Options{
		CassetteName:       cassetteName,
		Mode:               gorecorder.ModeRecordOnce,
		SkipRequestLatency: true,
		RealTransport:      utils.DefaultTransport,
	})
	if err != nil {
		return nil, err
	}
	rec.SetReplayableInteractions(false)
	var tokenCredential azcore.TokenCredential
	var subscriptionID string
	var tenantID string
	var clientID string
	var clientSecret string
	var clientCertPath string
	var clientCertPasswd string
	if rec.IsNewCassette() {
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
		if subscriptionID == "" {
			return nil, errors.New("required environment variable AZURE_SUBSCRIPTION_ID was not supplied")
		}
		tokenCredential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		tenantID = os.Getenv(utils.AzureTenantID)
		if tenantID == "" {
			return nil, errors.New("required environment variable AZURE_TENANT_ID was not supplied")
		}

		clientID = os.Getenv(utils.AzureClientID)
		if clientID == "" {
			return nil, errors.New("required environment variable AZURE_CLIENT_ID was not supplied")
		}
		clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
		clientCertPath = os.Getenv("AZURE_CLIENT_CERT_PATH")
		clientCertPasswd = os.Getenv("AZURE_CLIENT_CERT_PASSWD")
		if clientSecret == "" && clientCertPath == "" {
			return nil, errors.New("either AZURE_CLIENT_SECRET or AZURE_CLIENT_CERT_PATH must be supplied")
		}
	} else {
		// if we are replaying, we won't need auth
		// and we use a dummy subscription ID
		subscriptionID = uuid.Nil.String()
		tokenCredential = DummyTokenCredential(func(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
			return azcore.AccessToken{}, nil
		})

		tenantID = "tenantid"
		clientID = "clientid"
		clientSecret = "clientsecret"
	}

	rec.AddHook(func(i *cassette.Interaction) error {
		//ignore inprogress requests
		if strings.Contains(i.Response.Body, "\"status\": \"InProgress\"") {
			i.DiscardOnSave = true
			return nil
		}
		if !strings.EqualFold(tenantID, "tenantid") {
			i.Request.URL = strings.Replace(i.Request.URL, tenantID, "tenantid", -1)
			i.Request.Body = strings.Replace(i.Request.Body, tenantID, "tenantid", -1)
			i.Response.Body = strings.Replace(i.Response.Body, tenantID, "tenantid", -1)
			if i.Request.Form.Has("tenant_id") {
				i.Request.Form.Set("tenant_id", tenantID)
			}
		}
		if !strings.EqualFold(clientID, "clientid") {
			i.Request.URL = strings.Replace(i.Request.URL, clientID, "clientid", -1)
			i.Request.Body = strings.Replace(i.Request.Body, clientID, "clientid", -1)
			i.Response.Body = strings.Replace(i.Response.Body, clientID, "clientid", -1)
			if i.Request.Form.Has("client_id") {
				i.Request.Form.Set("client_id", "clientid")
			}
		}

		if len(clientSecret) > 0 && !strings.EqualFold(clientSecret, "clientsecret") {
			i.Request.URL = strings.Replace(i.Request.URL, clientSecret, "clientsecret", -1)
			i.Request.Body = strings.Replace(i.Request.Body, clientSecret, "clientsecret", -1)
			i.Response.Body = strings.Replace(i.Response.Body, clientSecret, "clientsecret", -1)
			if i.Request.Form.Has("client_secret") {
				i.Request.Form.Set("client_secret", "clientsecret")
			}
		}
		if i.Request.Form.Has("client_assertion") {
			i.Request.Form.Set("client_assertion", "clientassertion")
			i.Request.Body = "client_assertion=clientassertion&client_assertion_type=urn%3Aietf%3Aparams%3Aoauth%3Aclient-assertion-type%3Ajwt-bearer&client_id=clientid&client_info=1&grant_type=client_credentials&scope=https%3A%2F%2Fmanagement.azure.com%2F.default+openid+offline_access+profile"
		}
		if strings.Contains(i.Response.Body, "access_token") {
			i.Response.Body = `{"token_type":"Bearer","expires_in":86399,"ext_expires_in":86399,"access_token":"faketoken"}`
		}
		if strings.Contains(i.Response.Body, "-----BEGIN RSA PRIVATE KEY-----") {
			i.Response.Body = "{\r\n  \"privateKey\": \"-----BEGIN RSA PRIVATE KEY-----\\r\\n\\r\\n-----END RSA PRIVATE KEY-----\\r\\n\",\r\n  \"publicKey\": \"ssh-rsa {KEY} generated-by-azure\",\r\n  \"id\": \"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/AKS-CIT-SSHPUBLICKEYRESOURCE/providers/Microsoft.Compute/sshPublicKeys/testResource\"\r\n}"
		}
		if strings.Contains(i.Response.Body, "skiptoken") {
			re := regexp.MustCompile(`skiptoken=(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?`)
			i.Response.Body = string(re.ReplaceAll([]byte(i.Response.Body), []byte("skiptoken=skiptoken")))
		}
		if strings.Contains(i.Request.URL, "skiptoken") {
			re := regexp.MustCompile(`skiptoken=(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}%3D%3D|[A-Za-z0-9+/]{3}%3D)?`)
			i.Request.URL = string(re.ReplaceAll([]byte(i.Request.URL), []byte("skiptoken=skiptoken")))
		}
		if strings.Contains(i.Request.RequestURI, "skiptoken") {
			re := regexp.MustCompile(`skiptoken=(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}%3D%3D|[A-Za-z0-9+/]{3}%3D)?`)
			i.Response.Body = string(re.ReplaceAll([]byte(i.Response.Body), []byte("skiptoken=skiptoken")))
		}

		for _, header := range requestHeadersToRemove {
			delete(i.Request.Headers, header)
		}

		for _, header := range responseHeadersToRemove {
			delete(i.Response.Headers, header)
		}

		i.Request.Body = hideRecordingData(i.Request.Body)
		i.Response.Body = hideRecordingData(i.Response.Body)
		i.Request.URL = hideUUID(i.Request.URL)

		for _, values := range i.Request.Headers {
			for i := range values {
				values[i] = hideUUID(values[i])
			}
		}

		for _, values := range i.Response.Headers {
			for i := range values {
				values[i] = hideUUID(values[i])
			}
		}

		return nil
	}, gorecorder.BeforeSaveHook)

	return &Recorder{
		credential:       tokenCredential,
		rec:              rec,
		subscriptionID:   subscriptionID,
		tenantID:         tenantID,
		clientID:         clientID,
		clientSecret:     clientSecret,
		clientCertPath:   clientCertPath,
		clientCertPasswd: clientCertPasswd,
	}, nil
}

func (r *Recorder) HTTPClient() *http.Client {
	return r.rec.GetDefaultClient()
}

func (r *Recorder) TokenCredential() azcore.TokenCredential {
	return r.credential
}

func (r *Recorder) SubscriptionID() string {
	return r.subscriptionID
}

func (r *Recorder) TenantID() string {
	return r.tenantID
}

func (r *Recorder) ClientID() string {
	return r.clientID
}

func (r *Recorder) ClientSecret() string {
	return r.clientSecret
}

func (r *Recorder) ClientCertPath() string {
	return r.clientCertPath
}

func (r *Recorder) ClientCertPasswd() string {
	return r.clientCertPasswd
}

func (r *Recorder) Stop() error {
	return r.rec.Stop()
}

func (r *Recorder) IsNewCassette() bool {
	return r.rec.IsNewCassette()
}