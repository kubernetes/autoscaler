/*
Copyright 2025 The Kubernetes Authors.

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

package protos

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoundTripBestOptionsRequest(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("1Gi"),
					},
				},
			}},
		},
	}
	podBytes, err := pod.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
		Spec:       corev1.NodeSpec{},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("1Gi"),
			},
		},
	}
	nodeBytes, err := node.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	r := &BestOptionsRequest{
		Options: []*Option{{
			// This field should remain serializable
			PodBytes: [][]byte{podBytes},
		}},
		// This field is expected to stop being serializable in 1.35, and to need to be removed from the .proto file
		// This field should remain serializable
		NodeBytesMap: map[string][]byte{"node": nodeBytes},
	}
	data, err := proto.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	r2 := &BestOptionsRequest{}
	if err := proto.Unmarshal(data, r2); err != nil {
		t.Fatal(err)
	}

	pod2 := &corev1.Pod{}
	if err := pod2.Unmarshal(r2.Options[0].PodBytes[0]); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(r, r2) {
		t.Fatalf("message did not round-trip: %s", cmp.Diff(r, r2))
	}
	// Pod bytes must remain round-trippable
	if !apiequality.Semantic.DeepEqual(pod, pod2) {
		t.Fatalf("pod bytes did not round-trip: %s", cmp.Diff(r, r2))
	}

	node2 := &corev1.Node{}
	if err := node2.Unmarshal(r2.NodeBytesMap["node"]); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(r, r2) {
		t.Fatalf("message did not round-trip: %s", cmp.Diff(r, r2))
	}
	// Node bytes must remain round-trippable
	if !apiequality.Semantic.DeepEqual(node, node2) {
		t.Fatalf("pod bytes did not round-trip: %s", cmp.Diff(r, r2))
	}
}

func TestRoundTripBestOptionsResponse(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("1Gi"),
					},
				},
			}},
		},
	}
	podBytes, err := pod.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	r := &BestOptionsResponse{
		Options: []*Option{{
			// This field should remain serializable
			PodBytes: [][]byte{podBytes},
		}},
	}
	data, err := proto.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	r2 := &BestOptionsResponse{}
	if err := proto.Unmarshal(data, r2); err != nil {
		t.Fatal(err)
	}

	if !proto.Equal(r, r2) {
		t.Fatalf("message did not round-trip: %s", cmp.Diff(r, r2))
	}
}
