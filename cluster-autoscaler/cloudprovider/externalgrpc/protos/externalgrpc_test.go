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
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoundTripPricingPodPriceRequest(t *testing.T) {
	t1 := time.Unix(1000, 100)
	t1meta := metav1.NewTime(t1)
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

	r := &PricingPodPriceRequest{
		// These three fields are expected to stop being serializable in 1.35,
		// and to need to be removed from the .proto file
		StartTime: &t1meta,
		EndTime:   &t1meta,
		Pod:       pod,

		// These fields should remain serializable
		StartTimestamp: timestamppb.New(t1),
		EndTimestamp:   timestamppb.New(t1),
		PodBytes:       podBytes,
	}
	data, err := proto.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	r2 := &PricingPodPriceRequest{}
	if err := proto.Unmarshal(data, r2); err != nil {
		t.Fatal(err)
	}

	pod2 := &corev1.Pod{}
	if err := pod2.Unmarshal(r2.PodBytes); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(r, r2) {
		t.Fatalf("message did not round-trip: %s", cmp.Diff(r, r2))
	}
	// The Pod field is expected to be removed in 1.35
	if !apiequality.Semantic.DeepEqual(pod, r2.Pod) {
		t.Fatalf("pod did not round-trip: %s", cmp.Diff(r, r2))
	}
	// Pod bytes must remain round-trippable
	if !apiequality.Semantic.DeepEqual(pod, pod2) {
		t.Fatalf("pod bytes did not round-trip: %s", cmp.Diff(r, r2))
	}
}

func TestRoundTripPricingNodePriceRequest(t *testing.T) {
	t1 := time.Unix(1000, 100)
	t1meta := metav1.NewTime(t1)

	r := &PricingNodePriceRequest{
		// These three fields are expected to stop being serializable in 1.35,
		// and to need to be removed from the .proto file
		StartTime: &t1meta,
		EndTime:   &t1meta,

		// These fields should remain serializable
		StartTimestamp: timestamppb.New(t1),
		EndTimestamp:   timestamppb.New(t1),
	}
	data, err := proto.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	r2 := &PricingNodePriceRequest{}
	if err := proto.Unmarshal(data, r2); err != nil {
		t.Fatal(err)
	}

	if !proto.Equal(r, r2) {
		t.Fatalf("message did not round-trip: %s", cmp.Diff(r, r2))
	}
}
