/*
Copyright 2018 The Kubernetes Authors.

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

package signals

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	correctResponse = `{
		"status":"success",
		"data":{
			"resultType": "matrix",
		        "result": []}}`
)

type mockHTTPGetter struct {
	mock.Mock
}

func (m mockHTTPGetter) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	var returnArg http.Response
	if args.Get(0) != nil {
		returnArg = args.Get(0).(http.Response)
	}
	return &returnArg, args.Error(1)
}

// This type is here to implement io.ReadCloser used by http.Response.Body.
type readerPseudoCloser struct {
	*strings.Reader
}

func (r readerPseudoCloser) Close() error {
	return fmt.Errorf("readerPseudoCloser cannot really close anything")
}

func newReaderPseudoCloser(s string) readerPseudoCloser {
	return readerPseudoCloser{strings.NewReader(s)}
}

func TestUrl(t *testing.T) {
	retryDelay = time.Hour
	mockGetter := mockHTTPGetter{}
	client := NewPrometheusClient(&mockGetter, "https://1.1.1.1")
	mockGetter.On("Get", "https://1.1.1.1/api/v1/query?query=up%7Ba%3Db%7D%5B2d%5D").Times(1).Return(
		http.Response{
			StatusCode: http.StatusOK,
			Body:       newReaderPseudoCloser(correctResponse)}, nil)
	tss, err := client.GetTimeseries("up{a=b}[2d]")
	assert.Nil(t, err)
	assert.NotNil(t, tss)
	assert.Empty(t, tss)
}

func TestSuccessfulRetry(t *testing.T) {
	retryDelay = 100 * time.Millisecond
	mockGetter := mockHTTPGetter{}
	client := NewPrometheusClient(&mockGetter, "http://bla.com")
	mockGetter.On("Get", mock.AnythingOfType("string")).Times(1).Return(
		http.Response{StatusCode: http.StatusInternalServerError}, nil)
	mockGetter.On("Get", mock.AnythingOfType("string")).Times(1).Return(
		http.Response{
			StatusCode: http.StatusOK,
			Body:       newReaderPseudoCloser(correctResponse)}, nil)
	tss, err := client.GetTimeseries("up")
	assert.Nil(t, err)
	assert.NotNil(t, tss)
	assert.Empty(t, tss)
}

func TestUnsuccessfulRetries(t *testing.T) {
	retryDelay = 10 * time.Millisecond
	mockGetter := mockHTTPGetter{}
	client := NewPrometheusClient(&mockGetter, "http://bla.com")
	mockGetter.On("Get", mock.AnythingOfType("string")).Times(numRetries).Return(
		http.Response{StatusCode: http.StatusInternalServerError}, nil)
	_, err := client.GetTimeseries("up")
	assert.NotNil(t, err)
}
