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

package controller

import (
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	hpa "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type scaleClientMock struct {
	scales map[string]*autoscalingv1.Scale
}

func scalesMockKey(namespace string, scaleRef hpa.CrossVersionObjectReference) string {
	return fmt.Sprintf("%s/%s/%s/%s", namespace, scaleRef.APIVersion, scaleRef.Kind, scaleRef.Name)
}

func (s *scaleClientMock) PutForTest(namespace string, scaleRef hpa.CrossVersionObjectReference, scale *autoscalingv1.Scale) {
	key := scalesMockKey(namespace, scaleRef)
	s.scales[key] = scale
}

func (s *scaleClientMock) GetForTest(namespace string, scaleRef hpa.CrossVersionObjectReference) *autoscalingv1.Scale {
	key := scalesMockKey(namespace, scaleRef)
	return s.scales[key]
}

func (s *scaleClientMock) GetScale(namespace string, scaleRef hpa.CrossVersionObjectReference) (*autoscalingv1.Scale, *schema.GroupResource, error) {
	key := scalesMockKey(namespace, scaleRef)
	if scale, found := s.scales[key]; found {
		return scale, &schema.GroupResource{
			Group:    scaleRef.APIVersion,
			Resource: scaleRef.Kind,
		}, nil
	}
	return nil, nil, fmt.Errorf("Not found: %s", key)
}

func (s *scaleClientMock) UpdateScale(scale *autoscalingv1.Scale, resource *schema.GroupResource) error {
	key := scalesMockKey(scale.Namespace, hpa.CrossVersionObjectReference{
		Name:       scale.Name,
		APIVersion: resource.Group,
		Kind:       resource.Resource,
	})
	if _, found := s.scales[key]; found {
		s.scales[key] = scale
		return nil
	}
	return fmt.Errorf("Not found: %s", key)
}
