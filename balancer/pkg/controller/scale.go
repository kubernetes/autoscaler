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
	"context"
	"fmt"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	hpa "k8s.io/api/autoscaling/v2"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/klog/v2"
)

// ScaleClient implements ScaleClientInterface and issues real queries to K8S
// apiserver.
type ScaleClient struct {
	context         context.Context
	scaleNamespacer scaleclient.ScalesGetter
	mapper          apimeta.RESTMapper
}

// ScaleClientInterface is an interface to interact with Scale subresources.
type ScaleClientInterface interface {
	// GetScale gets scale subresource for the given reference.
	GetScale(namespace string, scaleRef hpa.CrossVersionObjectReference) (*autoscalingv1.Scale, *schema.GroupResource, error)

	// UpdateScale updates the given scale resource.
	UpdateScale(scale *autoscalingv1.Scale, resource *schema.GroupResource) error
}

// NewScaleClient builds scale client.
func NewScaleClient(context context.Context, scale scaleclient.ScalesGetter, mapper apimeta.RESTMapper) *ScaleClient {
	return &ScaleClient{
		context:         context,
		scaleNamespacer: scale,
		mapper:          mapper,
	}
}

// GetScale gets scale subresource for the given reference. Copied from HPA controller.
// TODO(mwielgus): Add cache if frequent scale resource lookups become a problem.
func (s *ScaleClient) GetScale(namespace string, scaleRef hpa.CrossVersionObjectReference) (*autoscalingv1.Scale, *schema.GroupResource, error) {

	reference := fmt.Sprintf("%s/%s/%s", scaleRef.Kind, namespace, scaleRef.Name)
	targetGV, err := schema.ParseGroupVersion(scaleRef.APIVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGK := schema.GroupKind{
		Group: targetGV.Group,
		Kind:  scaleRef.Kind,
	}

	mappings, err := s.mapper.RESTMappings(targetGK)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	scale, gr, err := s.scaleForResourceMappings(namespace, scaleRef.Name, mappings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query scale subresource for %s: %v", reference, err)
	}
	return scale, &gr, nil
}

// scaleForResourceMappings attempts to fetch the scale for the
// resource with the given name and namespace, trying each RESTMapping
// in turn until a working one is found.  If none work, the first error
// is returned.  It returns both the scale, as well as the group-resource from
// the working mapping.
func (s *ScaleClient) scaleForResourceMappings(namespace, name string,
	mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := s.scaleNamespacer.Scales(namespace).Get(s.context,
			targetGR, name, metav1.GetOptions{})

		if err == nil {
			return scale, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}

// UpdateScale updates the given scale resource.
func (s *ScaleClient) UpdateScale(scale *autoscalingv1.Scale, resource *schema.GroupResource) error {
	klog.V(4).Infof("Scaling %s/%s/%s to %d", resource.Resource, scale.Namespace, scale.Name,
		scale.Spec.Replicas)
	_, err := s.scaleNamespacer.Scales(scale.Namespace).Update(s.context, *resource, scale, metav1.UpdateOptions{})
	return err
}
