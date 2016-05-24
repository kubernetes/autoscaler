/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/client/unversioned/fake"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/runtime"
)

type MyReq struct {
	Request *http.Request
}

func (m *MyReq) isFor(method string, path string) bool {
	req := m.Request
	return method == req.Method && (req.URL.Path == path || req.URL.Path == strings.Join([]string{"/api/v1", path}, "") || req.URL.Path == strings.Join([]string{"/apis/extensions/v1beta1", path}, ""))
}

func TestRequiredPodsForNode(t *testing.T) {
	pod1 := kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
			SelfLink:  "pod1",
		},
	}
	// Manifest pod.
	pod2 := kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Name:      "pod2",
			Namespace: "kube-system",
			SelfLink:  "pod2",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
	}

	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	codec := testapi.Default.Codec()
	fakeClient := &fake.RESTClient{
		Codec: codec,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			m := &MyReq{req}
			switch {
			case m.isFor("GET", "/pods"):
				return &http.Response{StatusCode: 200, Header: header,
					Body: objBody(codec, &kube_api.PodList{Items: []kube_api.Pod{pod1, pod2}})}, nil
			default:
				t.Fatalf("unexpected request: %v %#v\n%#v", req.Method, req.URL, req)
				return nil, nil
			}
		}),
	}

	clientconfig := &restclient.Config{
		ContentConfig: restclient.ContentConfig{
			ContentType:  runtime.ContentTypeJSON,
			GroupVersion: testapi.Default.GroupVersion(),
		},
	}

	client := client.NewOrDie(clientconfig)
	client.Client = fakeClient.Client
	client.ExtensionsClient.Client = fakeClient.Client

	pods, err := GetRequiredPodsForNode("node1", client)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
	assert.Equal(t, "pod2", pods[0].Name)
}

func objBody(codec runtime.Codec, obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(codec, obj))))
}
