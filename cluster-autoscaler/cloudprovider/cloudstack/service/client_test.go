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

package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const APIKey = "APIKey"
const SecretKey = "SecretKey"

type TestResponse struct {
	InnerResponse *InnerResponse `json:"response"`
}

type InnerResponse struct {
	Count     int    `json:"count"`
	ErrorCode int    `json:"errorcode,omitempty"`
	ErrorText string `json:"errortext,omitempty"`
}

func createResponse(success bool) (TestResponse, []byte) {
	response := TestResponse{
		InnerResponse: &InnerResponse{
			Count: 1,
		},
	}
	if !success {
		response.InnerResponse.ErrorCode = 500
		response.InnerResponse.ErrorText = "ERROR"
	}
	js, _ := json.Marshal(response)
	return response, js
}

func createQueryJobResponse() (QueryAsyncJobResponse, []byte) {
	response := QueryAsyncJobResponse{
		JobResponse: &JobResponse{
			JobStatus: 0,
			JobID:     "abcd",
		},
	}
	js, _ := json.Marshal(response)
	return response, js
}
func TestQueryString(t *testing.T) {
	params := map[string]string{
		"abc": "def",
		"ghi": "jkl",
	}
	command := "hello"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, APIKey, query.Get("apiKey"))
		assert.Equal(t, command, query.Get("command"))
		for k, v := range params {
			assert.Equal(t, query.Get(k), v)
		}
	}))

	config := &Config{
		APIKey:    APIKey,
		SecretKey: SecretKey,
		Endpoint:  server.URL,
	}
	client := NewAPIClient(config)
	client.NewRequest(command, params, nil)
}

func testNewRequestError(t *testing.T) {
	_, js := createResponse(false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(js)
	}))

	config := &Config{
		APIKey:    APIKey,
		SecretKey: SecretKey,
		Endpoint:  server.URL,
	}
	client := NewAPIClient(config)

	_, err := client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, nil)
	assert.NotEqual(t, nil, err)
}

func testNewRequestSuccess(t *testing.T) {
	response, js := createResponse(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(js)
	}))

	config := &Config{
		APIKey:    APIKey,
		SecretKey: SecretKey,
		Endpoint:  server.URL,
	}
	client := NewAPIClient(config)

	var out TestResponse
	client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, &out)
	assert.Equal(t, response, out)
}

func TestNewRequest(t *testing.T) {
	t.Run("testNewRequestError", testNewRequestError)
	t.Run("testNewRequestSuccess", testNewRequestSuccess)
}

func testNewRequestWithPollError(t *testing.T) {
	response, js1 := createQueryJobResponse()

	jobResult, _ := createResponse(false)
	response.JobResponse.JobStatus = 2
	response.JobResponse.JobResult = jobResult
	js2, _ := json.Marshal(response)

	command := "hello"
	queryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("command")
		switch cmd {
		case command:
			w.Write(js1)
		case "queryAsyncJobResult":
			queryCount++
			w.Write(js2)
		default:
			t.Errorf("Unknown command called : %s", cmd)
		}
	}))

	config := &Config{
		APIKey:       APIKey,
		SecretKey:    SecretKey,
		Endpoint:     server.URL,
		PollInterval: 1,
	}
	client := NewAPIClient(config)

	_, err := client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, nil)

	assert.Equal(t, 1, queryCount)
	assert.NotEqual(t, nil, err)
}

func testNewRequestWithPollTimeout(t *testing.T) {
	_, js := createQueryJobResponse()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(js)
	}))

	config := &Config{
		APIKey:       APIKey,
		SecretKey:    SecretKey,
		Endpoint:     server.URL,
		PollInterval: 2,
		Timeout:      1,
	}
	client := NewAPIClient(config)

	var out TestResponse
	_, err := client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, &out)

	assert.NotEqual(t, nil, err)
}

func testNewRequestWithPollSuccess(t *testing.T) {
	response, js1 := createQueryJobResponse()

	jobResult, _ := createResponse(true)
	response.JobResponse.JobStatus = 1
	response.JobResponse.JobResult = jobResult
	js2, _ := json.Marshal(response)

	command := "hello"
	queryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("command")
		switch cmd {
		case command:
			w.Write(js1)
		case "queryAsyncJobResult":
			queryCount++
			if (queryCount) < 2 {
				w.Write(js1)
			} else {
				w.Write(js2)
			}
		default:
			t.Errorf("Unknown command called : %s", cmd)
		}
	}))

	config := &Config{
		APIKey:       APIKey,
		SecretKey:    SecretKey,
		Endpoint:     server.URL,
		PollInterval: 1,
	}
	client := NewAPIClient(config)

	var out TestResponse
	client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, &out)

	assert.Equal(t, 2, queryCount)
	assert.Equal(t, jobResult.InnerResponse, out.InnerResponse)
}

func TestNewRequestWithPoll(t *testing.T) {
	t.Run("testNewRequestWithPollError", testNewRequestWithPollError)
	t.Run("testNewRequestWithPollTimeout", testNewRequestWithPollTimeout)
	t.Run("testNewRequestWithPollSuccess", testNewRequestWithPollSuccess)
}

func TestClose(t *testing.T) {
	_, js := createQueryJobResponse()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(js)
	}))

	config := &Config{
		APIKey:    APIKey,
		SecretKey: SecretKey,
		Endpoint:  server.URL,
	}
	client := NewAPIClient(config)

	time.AfterFunc(2*time.Second, func() {
		client.Close()
	})

	var out TestResponse
	_, err := client.NewRequest("hello", map[string]string{
		"param": "abcd",
	}, &out)
	assert.NotEqual(t, nil, err)
}
