/*
Copyright 2020 The Kubernetes Authors.

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

package retry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"

	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// RateLimited error string
const RateLimited = "rate limited"

var (
	// The function to get current time.
	now = time.Now

	// StatusCodesForRetry are a defined group of status code for which the client will retry.
	StatusCodesForRetry = []int{
		http.StatusRequestTimeout,      // 408
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
	}
)

// Error indicates an error returned by Azure APIs.
type Error struct {
	// Retriable indicates whether the request is retriable.
	Retriable bool
	// HTTPStatusCode indicates the response HTTP status code.
	HTTPStatusCode int
	// RetryAfter indicates the time when the request should retry after throttling.
	// A throttled request is retriable.
	RetryAfter time.Time
	// RetryAfter indicates the raw error from API.
	RawError error
}

// Error returns the error.
// Note that Error doesn't implement error interface because (nil *Error) != (nil error).
func (err *Error) Error() error {
	if err == nil {
		return nil
	}

	// Convert time to seconds for better logging.
	retryAfterSeconds := 0
	curTime := now()
	if err.RetryAfter.After(curTime) {
		retryAfterSeconds = int(err.RetryAfter.Sub(curTime) / time.Second)
	}

	return fmt.Errorf("Retriable: %v, RetryAfter: %ds, HTTPStatusCode: %d, RawError: %w",
		err.Retriable, retryAfterSeconds, err.HTTPStatusCode, err.RawError)
}

// IsThrottled returns true the if the request is being throttled.
func (err *Error) IsThrottled() bool {
	if err == nil {
		return false
	}

	return err.HTTPStatusCode == http.StatusTooManyRequests || err.RetryAfter.After(now())
}

// IsNotFound returns true the if the requested object wasn't found
func (err *Error) IsNotFound() bool {
	if err == nil {
		return false
	}

	return err.HTTPStatusCode == http.StatusNotFound
}

// NewError creates a new Error.
func NewError(retriable bool, err error) *Error {
	return &Error{
		Retriable: retriable,
		RawError:  err,
	}
}

// NewError creates a new Error. Returns nil if err is nil
func NewErrorOrNil(retriable bool, err error) *Error {
	if err == nil {
		return nil
	}
	return NewError(retriable, err)
}

// GetRetriableError gets new retriable Error.
func GetRetriableError(err error) *Error {
	return &Error{
		Retriable: true,
		RawError:  err,
	}
}

// GetRateLimitError creates a new error for rate limiting.
func GetRateLimitError(isWrite bool, opName string) *Error {
	opType := "read"
	if isWrite {
		opType = "write"
	}
	return GetRetriableError(fmt.Errorf("azure cloud provider %s(%s) for operation %q", RateLimited, opType, opName))
}

// GetThrottlingError creates a new error for throttling.
func GetThrottlingError(operation, reason string, retryAfter time.Time) *Error {
	rawError := fmt.Errorf("azure cloud provider throttled for operation %s with reason %q", operation, reason)
	return &Error{
		Retriable:  true,
		RawError:   rawError,
		RetryAfter: retryAfter,
	}
}

// GetError gets a new Error based on resp and error.
func GetError(resp *http.Response, err error) *Error {
	if err == nil && resp == nil {
		return nil
	}

	if err == nil && resp != nil && IsSuccessHTTPResponse(resp) {
		// HTTP 2xx suggests a successful response
		return nil
	}

	retryAfter := time.Time{}
	if retryAfterDuration := getRetryAfter(resp); retryAfterDuration != 0 {
		retryAfter = now().Add(retryAfterDuration)
	}
	return &Error{
		RawError:       getRawError(resp, err),
		RetryAfter:     retryAfter,
		Retriable:      shouldRetryHTTPRequest(resp, err),
		HTTPStatusCode: getHTTPStatusCode(resp),
	}
}

// IsSuccessHTTPResponse determines if the response from an HTTP request suggests success
func IsSuccessHTTPResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	// HTTP 2xx suggests a successful response
	if 199 < resp.StatusCode && resp.StatusCode < 300 {
		return true
	}

	return false
}

func getRawError(resp *http.Response, err error) error {
	if err != nil {
		return err
	}

	if resp == nil || resp.Body == nil {
		return fmt.Errorf("empty HTTP response")
	}

	// return the http status if it is unable to get response body.
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	if len(respBody) == 0 {
		return fmt.Errorf("HTTP status code (%d)", resp.StatusCode)
	}

	// return the raw response body.
	return fmt.Errorf("%s", string(respBody))
}

func getHTTPStatusCode(resp *http.Response) int {
	if resp == nil {
		return -1
	}

	return resp.StatusCode
}

// shouldRetryHTTPRequest determines if the request is retriable.
func shouldRetryHTTPRequest(resp *http.Response, err error) bool {
	if resp != nil {
		for _, code := range StatusCodesForRetry {
			if resp.StatusCode == code {
				return true
			}
		}

		// should retry on <200, error>.
		if IsSuccessHTTPResponse(resp) && err != nil {
			return true
		}

		return false
	}

	// should retry when error is not nil and no http.Response.
	if err != nil {
		return true
	}

	return false
}

// getRetryAfter gets the retryAfter from http response.
// The value of Retry-After can be either the number of seconds or a date in RFC1123 format.
func getRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}

	ra := resp.Header.Get(consts.RetryAfterHeaderKey)
	if ra == "" {
		return 0
	}

	var dur time.Duration
	if retryAfter, _ := strconv.Atoi(ra); retryAfter > 0 {
		dur = time.Duration(retryAfter) * time.Second
	} else if t, err := time.Parse(time.RFC1123, ra); err == nil {
		dur = t.Sub(now())
	}
	return dur
}

// IsInHTTPStatusCodeSet return true when status code falls in the status code list
// It is used with doBackoffRetry to retry on some HTTPStatusCodes.
func IsInHTTPStatusCodeSet(rerr *Error, httpStatusCodes []int) bool {
	if rerr == nil {
		return false
	}
	for _, code := range httpStatusCodes {
		if rerr.HTTPStatusCode == code {
			return true
		}
	}

	return false
}

// isInErrorsSet return true when error message falls in the error message set
// It is used with doBackoffRetry to retry on some errors.
func isInErrorsSet(rerr *Error, errorMsgs []string) bool {

	if rerr == nil {
		return false
	}

	for _, err := range errorMsgs {
		if strings.Contains(rerr.RawError.Error(), err) {
			return true
		}
	}
	return false
}

// GetStatusNotFoundAndForbiddenIgnoredError gets an error with StatusNotFound and StatusForbidden ignored.
// It is only used in DELETE operations.
func GetStatusNotFoundAndForbiddenIgnoredError(resp *http.Response, err error) *Error {
	rerr := GetError(resp, err)
	if rerr == nil {
		return nil
	}

	// Returns nil when it is StatusNotFound error.
	if rerr.HTTPStatusCode == http.StatusNotFound {
		klog.V(3).Infof("Ignoring StatusNotFound error: %+v", rerr)
		return nil
	}

	// Returns nil if the status code is StatusForbidden.
	// This happens when AuthorizationFailed is reported from Azure API.
	if rerr.HTTPStatusCode == http.StatusForbidden {
		klog.V(3).Infof("Ignoring StatusForbidden error: %+v", rerr)
		return nil
	}

	return rerr
}

// IsErrorRetriable returns true if the error is retriable.
func IsErrorRetriable(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Retriable: true")
}

// HasStatusForbiddenOrIgnoredError return true if the given error code is part of the error message
// This should only be used when trying to delete resources
func HasStatusForbiddenOrIgnoredError(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), fmt.Sprintf("HTTPStatusCode: %d", http.StatusNotFound)) {
		return true
	}

	if strings.Contains(err.Error(), fmt.Sprintf("HTTPStatusCode: %d", http.StatusForbidden)) {
		return true
	}
	return false
}

// GetVMSSMetadataByRawError gets the vmss name by parsing the error message
func GetVMSSMetadataByRawError(err *Error) (string, string, error) {
	if err == nil || !isErrorLoadBalancerInUseByVirtualMachineScaleSet(err.RawError.Error()) {
		return "", "", nil
	}

	reg := regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.*)/providers/Microsoft.Compute/virtualMachineScaleSets/(.+).`)
	matches := reg.FindStringSubmatch(err.ServiceErrorMessage())
	if len(matches) != 3 {
		return "", "", fmt.Errorf("GetVMSSMetadataByRawError: couldn't find a VMSS resource Id from error message %w", err.RawError)
	}

	return matches[1], matches[2], nil
}

// isErrorLoadBalancerInUseByVirtualMachineScaleSet determines if the Error is
// LoadBalancerInUseByVirtualMachineScaleSet
func isErrorLoadBalancerInUseByVirtualMachineScaleSet(rawError string) bool {
	return strings.Contains(rawError, "LoadBalancerInUseByVirtualMachineScaleSet")
}

const (
	// OperationNotAllowed is an umbrella errrfor a lot of errors
	OperationNotAllowed string = "OperationNotAllowed"
	// QuotaExceeded falls under OperationNotAllowed error code but we make it more specific here
	QuotaExceeded string = "QuotaExceeded"
)

// ServiceRawError wraps the RawError field satisfying autorest.ServiceError
type ServiceRawError struct {
	ServiceError *azure.ServiceError `json:"error,omitempty"`
}

// ServiceErrorMessage returns the message associated with the autorest.ServiceError body
func (err *Error) ServiceErrorMessage() string {
	if err == nil || err.RawError == nil {
		return ""
	}

	sre := ServiceRawError{}
	marshalErr := json.Unmarshal([]byte(err.RawError.Error()), &sre)
	if marshalErr != nil {
		return ""
	}
	if sre.ServiceError == nil {
		return ""
	}
	return sre.ServiceError.Message
}

// ServiceErrorCode returns the code associated with the autorest.ServiceError body
func (err *Error) ServiceErrorCode() string {
	if err == nil || err.RawError == nil {
		return ""
	}

	sre := ServiceRawError{}
	marshalErr := json.Unmarshal([]byte(err.RawError.Error()), &sre)
	if marshalErr != nil {
		return ""
	}
	if sre.ServiceError == nil {
		return ""
	}
	return classifyErrorCode(*sre.ServiceError)
}

func classifyErrorCode(sre azure.ServiceError) string {
	if sre.Code == OperationNotAllowed {
		return getOperationNotAllowedReason(sre.Message)
	}
	return sre.Code
}

// getOperationNotAllowedReason attempts to better classify OperationNotAllowed errors
// by looking at the message
func getOperationNotAllowedReason(msg string) string {
	if strings.Contains(strings.ToLower(msg), strings.ToLower("Quota increase")) {
		return QuotaExceeded
	}
	return OperationNotAllowed
}

// PartialUpdateError implements error interface. It is meant to be returned for errors with http status code of 2xx
type PartialUpdateError struct {
	message string
}

func NewPartialUpdateError(msg string) *PartialUpdateError {
	return &PartialUpdateError{message: msg}
}

func (e *PartialUpdateError) Error() string {
	return e.message
}
