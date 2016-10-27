/*
Copyright 2016 The Kubernetes Authors.

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

package drain

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/fake"
	"k8s.io/kubernetes/pkg/controller"

	"k8s.io/kubernetes/pkg/runtime"
)

func TestDrain(t *testing.T) {
	labels := make(map[string]string)
	labels["my_key"] = "my_value"

	rc := api.ReplicationController{
		ObjectMeta: api.ObjectMeta{
			Name:              "rc",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			SelfLink:          testapi.Default.SelfLink("replicationcontrollers", "rc"),
		},
		Spec: api.ReplicationControllerSpec{
			Selector: labels,
		},
	}

	rcAnno := make(map[string]string)
	rcAnno[controller.CreatedByAnnotation] = refJSON(t, &rc)

	rcPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			Annotations:       rcAnno,
		},
		Spec: api.PodSpec{
			NodeName: "node",
		},
	}

	ds := extensions.DaemonSet{
		ObjectMeta: api.ObjectMeta{
			Name:              "ds",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			SelfLink:          "/apis/extensions/v1beta1/namespaces/default/daemonsets/ds",
		},
		Spec: extensions.DaemonSetSpec{
			Selector: &unversioned.LabelSelector{MatchLabels: labels},
		},
	}

	dsAnno := make(map[string]string)
	dsAnno[controller.CreatedByAnnotation] = refJSON(t, &ds)

	dsPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			Annotations:       dsAnno,
		},
		Spec: api.PodSpec{
			NodeName: "node",
		},
	}

	job := batch.Job{
		ObjectMeta: api.ObjectMeta{
			Name:              "job",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			SelfLink:          "/apis/extensions/v1beta1/namespaces/default/jobs/job",
		},
		Spec: batch.JobSpec{
			Selector: &unversioned.LabelSelector{MatchLabels: labels},
		},
	}

	jobPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			Annotations:       map[string]string{controller.CreatedByAnnotation: refJSON(t, &job)},
		},
	}

	rs := extensions.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name:              "rs",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			SelfLink:          testapi.Default.SelfLink("replicasets", "rs"),
		},
		Spec: extensions.ReplicaSetSpec{
			Selector: &unversioned.LabelSelector{MatchLabels: labels},
		},
	}

	rsAnno := make(map[string]string)
	rsAnno[controller.CreatedByAnnotation] = refJSON(t, &rs)

	rsPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
			Annotations:       rsAnno,
		},
		Spec: api.PodSpec{
			NodeName: "node",
		},
	}

	nakedPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
		},
		Spec: api.PodSpec{
			NodeName: "node",
		},
	}

	emptydirPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:              "bar",
			Namespace:         "default",
			CreationTimestamp: unversioned.Time{Time: time.Now()},
			Labels:            labels,
		},
		Spec: api.PodSpec{
			NodeName: "node",
			Volumes: []api.Volume{
				{
					Name:         "scratch",
					VolumeSource: api.VolumeSource{EmptyDir: &api.EmptyDirVolumeSource{Medium: ""}},
				},
			},
		},
	}

	tests := []struct {
		description string
		pods        []*api.Pod
		rcs         []api.ReplicationController
		replicaSets []extensions.ReplicaSet
		expectFatal bool
		expectPods  []*api.Pod
	}{
		{
			description: "RC-managed pod",
			pods:        []*api.Pod{rcPod},
			rcs:         []api.ReplicationController{rc},
			expectFatal: false,
			expectPods:  []*api.Pod{rcPod},
		},
		{
			description: "DS-managed pod",
			pods:        []*api.Pod{dsPod},
			expectFatal: false,
			expectPods:  []*api.Pod{},
		},
		{
			description: "Job-managed pod",
			pods:        []*api.Pod{jobPod},
			rcs:         []api.ReplicationController{rc},
			expectFatal: false,
			expectPods:  []*api.Pod{jobPod},
		},
		{
			description: "RS-managed pod",
			pods:        []*api.Pod{rsPod},
			replicaSets: []extensions.ReplicaSet{rs},
			expectFatal: false,
			expectPods:  []*api.Pod{rsPod},
		},
		{
			description: "naked pod",
			pods:        []*api.Pod{nakedPod},
			expectFatal: true,
			expectPods:  []*api.Pod{},
		},
		{
			description: "pod with EmptyDir",
			pods:        []*api.Pod{emptydirPod},
			expectFatal: true,
			expectPods:  []*api.Pod{},
		},
	}

	for _, test := range tests {

		codec := testapi.Default.Codec()
		extcodec := testapi.Extensions.Codec()

		fakeClient := &fake.RESTClient{
			Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				m := &MyReq{req}
				switch {
				case m.isFor("GET", "/namespaces/default/replicationcontrollers/rc"):
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &test.rcs[0])}, nil
				case m.isFor("GET", "/namespaces/default/daemonsets/ds"):
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(extcodec, &ds)}, nil
				case m.isFor("GET", "/namespaces/default/jobs/job"):
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(extcodec, &job)}, nil
				case m.isFor("GET", "/namespaces/default/replicasets/rs"):
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(extcodec, &test.replicaSets[0])}, nil
				case m.isFor("GET", "/namespaces/default/pods/bar"):
					return &http.Response{StatusCode: 404, Header: defaultHeader(), Body: objBody(codec, nil)}, nil
				case m.isFor("GET", "/replicationcontrollers"):
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &api.ReplicationControllerList{Items: test.rcs})}, nil
				default:
					t.Fatalf("%s: unexpected request: %v %#v\n%#v", test.description, req.Method, req.URL, req)
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

		pods, err := GetPodsForDeletionOnNodeDrain(test.pods, api.Codecs.UniversalDecoder(),
			true, true, true, client, 0)

		if test.expectFatal {
			if err == nil {
				t.Fatalf("%s: unexpected non-error", test.description)
			}
		}

		if !test.expectFatal {
			if err != nil {
				t.Fatalf("%s: error occurred: %v", test.description, err)
			}
		}

		if len(pods) != len(test.expectPods) {
			t.Fatalf("Wrong pod list content: %v", test.description)
		}
	}
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}

type MyReq struct {
	Request *http.Request
}

func (m *MyReq) isFor(method string, path string) bool {
	req := m.Request

	return method == req.Method && (req.URL.Path == path ||
		req.URL.Path == strings.Join([]string{"/api/v1", path}, "") ||
		req.URL.Path == strings.Join([]string{"/apis/extensions/v1beta1", path}, "") ||
		req.URL.Path == strings.Join([]string{"/apis/batch/v1", path}, ""))
}

func refJSON(t *testing.T, o runtime.Object) string {
	ref, err := api.GetReference(o)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	codec := testapi.Default.Codec()
	json := runtime.EncodeOrDie(codec, &api.SerializedReference{Reference: *ref})
	return string(json)
}

func objBody(codec runtime.Codec, obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(codec, obj))))
}
