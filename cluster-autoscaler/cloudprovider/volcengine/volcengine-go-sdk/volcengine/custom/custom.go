/*
Copyright 2023 The Kubernetes Authors.

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

package custom

import (
	"context"
	"net/http"
	"net/url"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/response"
)

type RequestMetadata struct {
	ServiceName string
	Version     string
	Action      string
	HttpMethod  string
	Region      string
	Request     *http.Request
	RawQuery    *url.Values
}

type ExtendHttpRequest func(ctx context.Context, request *http.Request)

type ExtendHttpRequestWithMeta func(ctx context.Context, request *http.Request, meta RequestMetadata)

type ExtraHttpParameters func(ctx context.Context) map[string]string

type ExtraHttpParametersWithMeta func(ctx context.Context, meta RequestMetadata) map[string]string

type ExtraHttpJsonBody func(ctx context.Context, input *map[string]interface{}, meta RequestMetadata)

type LogAccount func(ctx context.Context) *string

type DynamicCredentials func(ctx context.Context) (*credentials.Credentials, *string)

// DynamicCredentialsIncludeError func return Credentials info and error info when error appear
type DynamicCredentialsIncludeError func(ctx context.Context) (*credentials.Credentials, *string, error)

type CustomerUnmarshalError func(ctx context.Context, meta RequestMetadata, resp response.VolcengineResponse) error

type CustomerUnmarshalData func(ctx context.Context, info RequestInfo, resp response.VolcengineResponse) interface{}

type ForceJsonNumberDecode func(ctx context.Context, info RequestInfo) bool
