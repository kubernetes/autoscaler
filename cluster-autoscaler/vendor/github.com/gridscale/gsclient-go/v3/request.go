package gsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// gsRequest gridscale's custom gsRequest struct.
type gsRequest struct {
	uri                 string
	method              string
	body                interface{}
	queryParameters     map[string]string
	skipCheckingRequest bool
}

// CreateResponse represents a common response for creation.
type CreateResponse struct {
	// UUID of the object being created.
	ObjectUUID string `json:"object_uuid"`

	// UUID of the request.
	RequestUUID string `json:"request_uuid"`
}

// RequestStatus represents status of a request.
type RequestStatus map[string]RequestStatusProperties

// RequestStatusProperties holds  properties of a request's status.
type RequestStatusProperties struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	CreateTime GSTime `json:"create_time"`
}

// RequestError represents an error of a request.
type RequestError struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StatusCode  int
	RequestUUID string
}

const (
	authUserIDHeaderKey = "X-Auth-Userid"
	authTokenHeaderKey  = "X-Auth-Token"
	maskedValue         = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
)

// Error just returns error as string.
func (r RequestError) Error() string {
	message := r.Description
	if message == "" {
		message = "no error message received from server"
	}
	errorMessageFormat := "Status code: %v. Error: %s. Request UUID: %s. "
	if r.StatusCode >= 500 {
		errorMessageFormat += "Please report this error along with the request UUID."
	}
	return fmt.Sprintf(errorMessageFormat, r.StatusCode, message, r.RequestUUID)
}

const (
	requestUUIDHeader           = "X-Request-Id"
	requestRateLimitResetHeader = "Ratelimit-Reset"
	retryAfterHeader            = "Retry-After"
)

// This function takes the client and a struct and then adds the result to the given struct if possible.
func (r *gsRequest) execute(ctx context.Context, c Client, output interface{}) error {
	startTime := time.Now()
	var err error
	var requestUUID string
	// Get caller name.
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	callerName := details.Name()
	// No need to trace `waitForRequestCompleted` method.
	if !strings.Contains(callerName, "waitForRequestCompleted") {
		defer func() {
			interval := time.Now().Sub(startTime).Milliseconds()
			if err != nil {
				logger.WithFields(logrus.Fields{
					"method":      callerName,
					"timeMs":      interval,
					"requestUUID": requestUUID,
				}).Tracef("Failed with error %s", err.Error())
				return
			}
			logger.WithFields(logrus.Fields{
				"method":      callerName,
				"timeMs":      interval,
				"requestUUID": requestUUID,
			}).Tracef("Successful")
		}()
	}

	// Execute the request (including retrying when needed).
	requestUUID, responseBodyBytes, err := r.retryHTTPRequest(ctx, c.cfg)
	if err != nil {
		return err
	}

	// if output is set.
	if output != nil {
		// Unmarshal body bytes to the given struct.
		err = json.Unmarshal(responseBodyBytes, output)
		if err != nil {
			logger.Errorf("Error while marshaling JSON: %v", err)
			return err
		}
	}

	// If the client is synchronous, and the request does not skip
	// checking a request, wait until the request completes.
	if c.Synchronous() && !r.skipCheckingRequest {
		return c.waitForRequestCompleted(ctx, requestUUID)
	}
	return nil
}

// prepareHTTPRequest prepares a http request.
func (r *gsRequest) prepareHTTPRequest(ctx context.Context, cfg *Config) (*http.Request, error) {
	url := cfg.apiURL + r.uri
	logger.Debugf("Preparing %v request sent to URL: %v", r.method, url)

	// Convert the body of the request to json.
	jsonBody := new(bytes.Buffer)
	if r.body != nil {
		err := json.NewEncoder(jsonBody).Encode(r.body)
		if err != nil {
			return nil, err
		}
	}

	// Add authentication headers and content type.
	request, err := http.NewRequest(r.method, url, jsonBody)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	request.Header.Set("User-Agent", cfg.userAgent)
	request.Header.Set("Content-Type", bodyType)

	// Omit X-Auth-UserID when cfg.userUUID is empty.
	if cfg.userUUID != "" {
		request.Header.Set(authUserIDHeaderKey, cfg.userUUID)
	}
	// Omit X-Auth-Token when cfg.apiToken is empty.
	if cfg.apiToken != "" {
		request.Header.Set(authTokenHeaderKey, cfg.apiToken)
	}

	// Set headers based on a given list of custom headers.
	// Use Header.Set() instead of Header.Add() because we want to
	// override the headers' values if they are already set.
	for k, v := range cfg.httpHeaders {
		request.Header.Set(k, v)
	}

	// Set query parameters if there are any of them.
	query := request.URL.Query()
	for k, v := range r.queryParameters {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()
	logger.Debugf("Finished Preparing %v request sent to URL: %v://%v%v", request.Method, request.URL.Scheme, request.URL.Host, request.URL.RequestURI())
	return request, nil
}

// retryHTTPRequest prepares and sends a HTTP request.
// If 429 error code is returned from the server, retry after the rate-limit is reset.
// If 503 error code is returned and Retry-After response header is defined (x seconds),
// retry after x seconds.
// If 503 error code is returned and Retry-After response header is NOT defined,
// the next retry depends on the config of gsclient-go (delayIntervalMilliSecs and maxNumberOfRetries).
// Returns UUID (string), response body ([]byte), error
func (r *gsRequest) retryHTTPRequest(ctx context.Context, cfg *Config) (string, []byte, error) {
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	default:
	}
	var requestUUID string
	var responseBodyBytes []byte
	err := retryNTimes(func() (bool, error) {
		httpReq, err := r.prepareHTTPRequest(ctx, cfg)
		logger.Debugf("Request body: %v", httpReq.Body)
		logger.Debugf("Request headers: %v", maskHeaderCred(httpReq.Header))
		resp, err := cfg.httpClient.Do(httpReq)
		if err != nil {
			// If the error is caused by expired context, return context error and no need to retry.
			if ctx.Err() != nil {
				return false, ctx.Err()
			}

			if err, ok := err.(net.Error); ok {
				// exclude retry request with none GET method (write operations) in case of a request timeout or a context error.
				if err.Timeout() && r.method != http.MethodGet {
					return false, err
				}
				logger.Debugf("Retrying request due to network error %v", err)
				return true, err
			}
			logger.Errorf("Error while executing the request: %v", err)
			return false, err
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		requestUUID = resp.Header.Get(requestUUIDHeader)
		responseBodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf("Error while reading the response's body: %v", err)
			return false, err
		}
		logger.Debugf("Status code: %v. Request UUID: %v. Headers: %v", statusCode, requestUUID, resp.Header)
		// If the status code is an error code.
		if statusCode >= 300 {
			var errorMessage RequestError
			errorMessage.StatusCode = statusCode
			errorMessage.RequestUUID = requestUUID
			json.Unmarshal(responseBodyBytes, &errorMessage)

			switch statusCode {
			case http.StatusServiceUnavailable, http.StatusFailedDependency, http.StatusInternalServerError, http.StatusConflict:
				// Get the delay (in second) for the next retry
				delayDurationStr := resp.Header.Get(retryAfterHeader)
				delayDuration, err := strconv.Atoi(delayDurationStr)
				if err != nil { // If there is no valid "Retry-After", do normal retry.
					return true, errorMessage
				}
				select {
				case <-ctx.Done(): // If context expires first, return context.Err()
					return false, ctx.Err()
				case <-time.After(time.Duration(delayDuration) * time.Second): // If the delay finishes first, continue.
				}
				logger.Debugf("Retrying request: %v method sent to url %v with body %v", r.method, httpReq.URL.RequestURI(), r.body)
				return true, errorMessage

			case http.StatusTooManyRequests: // If status code is 429, that means we reach the rate limit.
				// Get the time that the rate limit will be reset.
				rateLimitResetTimestamp := resp.Header.Get(requestRateLimitResetHeader)
				delayMs, err := getDelayTimeInMsFromTimestampStr(rateLimitResetTimestamp)
				if err != nil {
					return false, err
				}
				// Delay the retry until the rate limit is reset.
				logger.Debugf("Delay request for %d ms: %v method sent to url %v with body %v", delayMs, r.method, httpReq.URL.RequestURI(), r.body)
				select {
				case <-ctx.Done(): // If context expires first, return context.Err()
					return false, ctx.Err()
				case <-time.After(time.Duration(delayMs) * time.Millisecond): // If the delay finishes first, continue.
				}
				logger.Debugf("Retrying request - recursive retryHTTPRequest (due to rate limit): %v method sent to url %v with body %v", r.method, httpReq.URL.RequestURI(), r.body)
				// Recursive retryHTTPRequest.
				requestUUID, responseBodyBytes, err = r.retryHTTPRequest(ctx, cfg)
				// Because of the recursive retryHTTPRequest, no need to retry here.
				return false, err
			}
			logger.Errorf(
				"Error message: %v. Title: %v. Code: %v. Request UUID: %v.",
				errorMessage.Description,
				errorMessage.Title,
				errorMessage.StatusCode,
				errorMessage.RequestUUID,
			)
			return false, errorMessage
		}
		logger.Debugf("Response body: %v", string(responseBodyBytes))
		return false, nil
	}, cfg.maxNumberOfRetries, cfg.delayInterval)
	return requestUUID, responseBodyBytes, err
}

// getDelayTimeInMsFromTimestampStr takes a unix timestamp (ms) string
// and returns the amount of ms to delay the next HTTP request retry.
// Return error if the input string is not a valid unix timestamp(ms).
func getDelayTimeInMsFromTimestampStr(timestamp string) (int64, error) {
	if timestamp == "" {
		return 0, errors.New("timestamp is empty")
	}
	// convert timestamp from string to int
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, err
	}
	currentTimestampMs := time.Now().UnixNano() / 1000000
	return timestampInt - currentTimestampMs, nil
}

// maskHeaderCred returns new HTTP header with masked credentials.
// Used when debugging.
func maskHeaderCred(header http.Header) http.Header {
	newHeaders := make(http.Header)
	for k, v := range header {
		if k == authUserIDHeaderKey || k == authTokenHeaderKey {
			if len(v[0]) > 5 {
				newHeaders[k] = []string{v[0][:5] + maskedValue}
				continue
			}
			newHeaders[k] = []string{maskedValue}
			continue
		}
		newHeaders[k] = v
	}
	return newHeaders
}
