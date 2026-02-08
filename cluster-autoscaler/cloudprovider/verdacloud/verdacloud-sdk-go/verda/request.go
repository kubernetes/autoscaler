/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Response wraps an HTTP response.
type Response struct {
	*http.Response
}

func getRequest[T any](ctx context.Context, client *Client, url string) (T, *Response, error) {
	var respBody T

	req, err := client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func requestWithBody[T any](ctx context.Context, client *Client, method, url string, reqBody any) (T, *Response, error) {
	var respBody T

	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return respBody, nil, err
		}

		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := client.NewRequest(ctx, method, url, reqBodyReader)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func postRequest[T any](ctx context.Context, client *Client, url string, reqBody any) (T, *Response, error) {
	return requestWithBody[T](ctx, client, http.MethodPost, url, reqBody)
}

func postRequestNoResult(ctx context.Context, client *Client, url string, reqBody any) (*Response, error) {
	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := client.NewRequest(ctx, http.MethodPost, url, reqBodyReader)
	if err != nil {
		return nil, err
	}

	return client.Do(req, nil)
}

//nolint:unused // Reserved for PUT endpoints
func putRequest[T any](ctx context.Context, client *Client, url string, reqBody any) (T, *Response, error) {
	return requestWithBody[T](ctx, client, http.MethodPut, url, reqBody)
}

func patchRequest[T any](ctx context.Context, client *Client, url string, reqBody any) (T, *Response, error) {
	return requestWithBody[T](ctx, client, http.MethodPatch, url, reqBody)
}

//nolint:unused // Reserved for DELETE endpoints with response bodies
func deleteRequest[T any](ctx context.Context, client *Client, url string) (T, *Response, error) {
	var respBody T

	req, err := client.NewRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func deleteRequestNoResult(ctx context.Context, client *Client, url string) (*Response, error) {
	req, err := client.NewRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req, nil)
}

func deleteRequestWithBody(ctx context.Context, client *Client, url string, reqBody any) (*Response, error) {
	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := client.NewRequest(ctx, http.MethodDelete, url, reqBodyReader)
	if err != nil {
		return nil, err
	}

	return client.Do(req, nil)
}
