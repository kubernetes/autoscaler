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

package universal

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/client"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/client/metadata"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/corehandlers"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/session"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/signer/volc"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcenginequery"
)

func New(session *session.Session) *Universal {
	return &Universal{
		Session: session,
	}
}

func (u *Universal) newClient(info RequestUniversal) *client.Client {
	config := u.Session.ClientConfig(info.ServiceName)
	c := client.New(
		*config.Config,
		metadata.ClientInfo{
			SigningName:   config.SigningName,
			SigningRegion: config.SigningRegion,
			Endpoint:      config.Endpoint,
			APIVersion:    info.Version,
			ServiceName:   info.ServiceName,
			ServiceID:     info.ServiceName,
		},
		config.Handlers,
	)
	c.Handlers.Build.PushBackNamed(corehandlers.SDKVersionUserAgentHandler)
	c.Handlers.Sign.PushBackNamed(volc.SignRequestHandler)
	c.Handlers.Build.PushBackNamed(volcenginequery.BuildHandler)
	c.Handlers.Unmarshal.PushBackNamed(volcenginequery.UnmarshalHandler)
	c.Handlers.UnmarshalMeta.PushBackNamed(volcenginequery.UnmarshalMetaHandler)
	c.Handlers.UnmarshalError.PushBackNamed(volcenginequery.UnmarshalErrorHandler)

	return c
}

func (u *Universal) getMethod(m HttpMethod) string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case HEAD:
		return "HEAD"
	default:
		return "GET"
	}
}

func getContentType(m ContentType) string {
	switch m {
	case ApplicationJSON:
		return "application/json"
	case FormUrlencoded:
		return "x-www-form-urlencoded"
	default:
		return ""
	}
}

func (u *Universal) DoCall(info RequestUniversal, input *map[string]interface{}) (output *map[string]interface{}, err error) {
	c := u.newClient(info)
	op := &request.Operation{
		HTTPMethod: u.getMethod(info.HttpMethod),
		HTTPPath:   "/",
		Name:       info.Action,
	}
	if input == nil {
		input = &map[string]interface{}{}
	}
	output = &map[string]interface{}{}
	req := c.NewRequest(op, input, output)

	if getContentType(info.ContentType) != "" {
		req.HTTPRequest.Header.Set("Content-Type", getContentType(info.ContentType))
	}
	err = req.Send()
	return output, err
}
