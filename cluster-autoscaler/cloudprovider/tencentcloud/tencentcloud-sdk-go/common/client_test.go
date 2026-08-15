/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/profile"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/regions"
)

type requestWithClientToken struct {
	tchttp.CommonRequest
	ClientToken *string `json:"ClientToken,omitempty" name:"ClientToken"`
}

func newTestRequest() *requestWithClientToken {
	var testRequest = &requestWithClientToken{
		CommonRequest: *tchttp.NewCommonRequest("cvm", "2017-03-12", "RunInstances"),
		ClientToken:   nil,
	}
	return testRequest
}

type retryErr struct{}

func (r retryErr) Error() string   { return "retry error" }
func (r retryErr) Timeout() bool   { return true }
func (r retryErr) Temporary() bool { return true }

var (
	successResp   = `{"Response": {"RequestId": ""}}`
	rateLimitResp = `{"Response": {"RequestId": "", "Error": {"Code": "RequestLimitExceeded"}}}`
)

type mockRT struct {
	NetworkFailures   int
	RateLimitFailures int

	NetworkTries   int
	RateLimitTries int
}

func (s *mockRT) RoundTrip(request *http.Request) (*http.Response, error) {
	if s.NetworkTries < s.NetworkFailures {
		s.NetworkTries++
		return nil, retryErr{}
	}

	if s.RateLimitTries < s.RateLimitFailures {
		s.RateLimitTries++
		return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(rateLimitResp))}, nil
	}

	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(successResp))}, nil
}

type testCase struct {
	msg      string
	prof     *profile.ClientProfile
	request  tchttp.Request
	response tchttp.Response
	specific mockRT
	expected mockRT
	success  bool
}

func TestNormalSucceedRequest(t *testing.T) {
	test(t, testCase{
		prof:     profile.NewClientProfile(),
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{},
		expected: mockRT{},
		success:  true,
	})
}

func TestNetworkFailedButSucceedAfterRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.NetworkFailureMaxRetries = 1
	prof.NetworkFailureRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{NetworkFailures: 1},
		expected: mockRT{NetworkTries: 1},
		success:  true,
	})
}

func TestNetworkFailedAndShouldNotRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.NetworkFailureMaxRetries = 1
	prof.NetworkFailureRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  tchttp.NewCommonRequest("cvm", "2017-03-12", "DescribeInstances"),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{NetworkFailures: 2},
		expected: mockRT{NetworkTries: 1},
		success:  false,
	})
}

func TestNetworkFailedAfterRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.NetworkFailureMaxRetries = 1
	prof.NetworkFailureRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{NetworkFailures: 2},
		expected: mockRT{NetworkTries: 2},
		success:  false,
	})
}

func TestRateLimitButSucceedAfterRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.RateLimitExceededMaxRetries = 1
	prof.RateLimitExceededRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{RateLimitFailures: 1},
		expected: mockRT{RateLimitTries: 1},
		success:  true,
	})
}

func TestRateLimitExceededAfterRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.RateLimitExceededMaxRetries = 1
	prof.RateLimitExceededRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{RateLimitFailures: 3},
		expected: mockRT{RateLimitTries: 2},
		success:  false,
	})
}

func TestBothFailuresOccurredButSucceedAfterRetry(t *testing.T) {
	prof := profile.NewClientProfile()
	prof.NetworkFailureMaxRetries = 1
	prof.NetworkFailureRetryDuration = profile.ConstantDurationFunc(0)
	prof.RateLimitExceededMaxRetries = 1
	prof.RateLimitExceededRetryDuration = profile.ConstantDurationFunc(0)

	test(t, testCase{
		prof:     prof,
		request:  newTestRequest(),
		response: tchttp.NewCommonResponse(),
		specific: mockRT{RateLimitFailures: 1, NetworkFailures: 1},
		expected: mockRT{RateLimitTries: 1, NetworkTries: 1},
		success:  true,
	})
}

func test(t *testing.T, tc testCase) {
	credential := NewCredential("", "")
	client := NewCommonClient(credential, regions.Guangzhou, tc.prof)

	client.WithHttpTransport(&tc.specific)

	err := client.Send(tc.request, tc.response)
	if tc.success != (err == nil) {
		t.Fatalf("unexpected failed on request: %+v", err)
	}

	if tc.expected.RateLimitTries != tc.specific.RateLimitTries {
		t.Fatalf("unexpected rate limit retry, expected %d, got %d",
			tc.expected.RateLimitTries, tc.specific.RateLimitTries)
	}

	if tc.expected.NetworkTries != tc.specific.NetworkTries {
		t.Fatalf("unexpected network failure retry, expected %d, got %d",
			tc.expected.NetworkTries, tc.specific.NetworkTries)
	}
}

func TestClient_withRegionBreaker(t *testing.T) {
	cpf := profile.NewClientProfile()
	//cpf.Debug =true
	cpf.DisableRegionBreaker = false
	cpf.BackupEndpoint = ""
	c := (&Client{}).Init("")
	c.WithProfile(cpf)
	if c.rb.backupEndpoint != "ap-guangzhou.tencentcloudapi.com" {
		t.Errorf("want %s ,got %s", "ap-beijing", c.rb.backupEndpoint)
	}
	if c.rb.maxFailNum != defaultMaxFailNum {
		t.Errorf("want %d ,got %d", defaultMaxFailNum, c.rb.maxFailNum)
	}
}
