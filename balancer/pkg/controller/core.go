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
	v1 "k8s.io/client-go/informers/core/v1"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	balancerapi "k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/balancer/pkg/pods"
	"k8s.io/autoscaler/balancer/pkg/policy"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// CoreInterface of the balancer controller. Handles individual Balancer reconciliation.
type CoreInterface interface {
	ProcessBalancer(balancer *balancerapi.Balancer, now time.Time) (*BalancerStatusInfo, *BalancerError)
	IsSynced() bool
}

// BalancerError adds phase information to error returned by balancer core.
type BalancerError struct {
	phase BalancerPhase
	err   error
}

// BalancerStatusInfo summarizes the balancing operation.
type BalancerStatusInfo struct {
	replicasObserved int32
	updated          bool
}

func newBalancerStatusInfo(replicas int32, updated bool) BalancerStatusInfo {
	return BalancerStatusInfo{replicasObserved: replicas, updated: updated}
}

// core is CoreInferface implementation.
type core struct {
	scaleClient ScaleClientInterface
	podLister   corelisters.PodLister
	podSynced   func() bool
}

func newCoreForTests(client ScaleClientInterface, lister corelisters.PodLister) CoreInterface {
	return &core{
		scaleClient: client,
		podLister:   lister,
		podSynced: func() bool {
			return true
		},
	}
}

// NewCore returns an implementation of the CoreInterface.
func NewCore(client ScaleClientInterface, informer v1.PodInformer) CoreInterface {
	return &core{
		scaleClient: client,
		podLister:   informer.Lister(),
		podSynced:   informer.Informer().HasSynced,
	}
}

type scaleInfo struct {
	scale         *autoscalingv1.Scale
	groupResource *schema.GroupResource
}

// BalancerPhase indicates the phase of the balancer reconciliation.
type BalancerPhase string

const (
	// 50 years
	infDeadline = time.Hour * 24 * 365 * 50

	// ScaleSubresourcePolling - phase where the scale subresources af a balancer are get.
	ScaleSubresourcePolling BalancerPhase = "ScaleSubresourcePolling"
	// PodListing - phase where pods under balancers target are listed.
	PodListing BalancerPhase = "PodListing"
	// PodLabelsChecking - phase where pods labels are checked.
	PodLabelsChecking BalancerPhase = "PodLabelsChecking"
	// ApplyingPolicyListing - phase where the balancer policy is applied.
	ApplyingPolicyListing BalancerPhase = "ApplyingBalancerPolicy"
	// ReplicaCountSetting - phase where balancer targets are resized.
	ReplicaCountSetting BalancerPhase = "ReplicaCountSetting"
)

func (b *BalancerError) Error() string {
	return fmt.Sprintf("%s: %v", b.phase, b.err.Error())
}

func newBalancerError(phase BalancerPhase, err error) *BalancerError {
	return &BalancerError{
		phase: phase,
		err:   err,
	}
}

// ProcessBalancer process Balancer and returns status information and/or error
// depending on how far the process progressed before encountering a problem.
func (c *core) ProcessBalancer(balancer *balancerapi.Balancer, now time.Time) (*BalancerStatusInfo, *BalancerError) {
	scaleInfos := make(map[string]scaleInfo)
	summaries := make(map[string]pods.Summary)

	for _, target := range balancer.Spec.Targets {
		scale, gr, err := c.scaleClient.GetScale(balancer.Namespace, target.ScaleTargetRef)
		if err != nil {
			return nil, newBalancerError(ScaleSubresourcePolling, err)
		}
		scaleInfos[target.Name] = scaleInfo{
			scale:         scale,
			groupResource: gr,
		}
	}

	balancerSelector, err := metav1.LabelSelectorAsSelector(&balancer.Spec.Selector)
	if err != nil {
		return nil, newBalancerError(PodLabelsChecking, fmt.Errorf("incorrect selector"))
	}

	statusInfo := BalancerStatusInfo{}

	for name, si := range scaleInfos {
		selector, err := labels.Parse(si.scale.Status.Selector)
		if err != nil {
			return nil, newBalancerError(PodListing, err)
		}
		podList, err := c.podLister.Pods(balancer.Namespace).List(selector)
		if err != nil {
			return nil, newBalancerError(PodListing, err)
		}

		for _, p := range podList {
			if !balancerSelector.Matches(labels.Set(p.Labels)) {
				return nil, newBalancerError(PodLabelsChecking,
					fmt.Errorf("incorrect labeling for pods in target %s", name))
			}
		}
		deadline := infDeadline
		if balancer.Spec.Policy.Fallback != nil {
			deadline = time.Duration(balancer.Spec.Policy.Fallback.StartupTimeoutSeconds) * time.Second
		}
		summary := pods.CalculateSummary(podList, now, deadline)
		summaries[name] = summary

		statusInfo.replicasObserved += summary.Total
	}
	placement, _, err := policy.GetPlacement(balancer, summaries)
	if err != nil {
		return &statusInfo, newBalancerError(ApplyingPolicyListing, err)
	}

	for name, scaleInfo := range scaleInfos {
		replicas, found := placement[name]
		if !found {
			return &statusInfo, newBalancerError(ApplyingPolicyListing, fmt.Errorf("placement for %s not found", name))
		}
		if scaleInfo.scale.Spec.Replicas != replicas {
			statusInfo.updated = true
			scaleInfo.scale.Spec.Replicas = replicas
			err := c.scaleClient.UpdateScale(scaleInfo.scale, scaleInfo.groupResource)
			if err != nil {
				return &statusInfo, newBalancerError(ReplicaCountSetting, err)
			}
		}
	}

	return &statusInfo, nil
}

func (c *core) IsSynced() bool {
	return c.podSynced()
}
