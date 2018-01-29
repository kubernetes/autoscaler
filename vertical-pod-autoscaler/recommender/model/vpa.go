/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"k8s.io/apimachinery/pkg/labels"
)

// Vpa (Vertical Pod Autoscaler) object is responsible for vertical scaling of
// Pods matching a given label selector.
type Vpa struct {
	ID VpaID
	// Labels selector that determines which Pods are controlled by this VPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Original (string) representation of the PodSelector.
	PodSelectorStr string
	// Pods controlled by this VPA object.
	Pods map[PodID]*PodState
}

// NewVpa returns a new Vpa with a given ID and pod selector. Doesn't set the
// links to the matched pods.
func NewVpa(id VpaID, podSelectorStr string) (*Vpa, error) {
	vpa := &Vpa{
		id, nil, "",
		make(map[PodID]*PodState), // Empty pods map.
	}
	if err := vpa.SetPodSelectorStr(podSelectorStr); err != nil {
		return nil, err
	}
	return vpa, nil
}

// SetPodSelectorStr sets the pod selector of the VPA to the given value, passed
// in the text format (see apimachinery/pkg/labels for the syntax). It returns
// an error if the string cannot be parsed. Doesn't update the links to the
// matched pods.
func (vpa *Vpa) SetPodSelectorStr(podSelectorStr string) error {
	podSelector, err := labels.Parse(podSelectorStr)
	if err != nil {
		return err
	}
	vpa.PodSelectorStr = podSelectorStr
	vpa.PodSelector = podSelector
	return nil
}

// MatchesPod returns true iff a given pod is matched by the Vpa pod selector.
func (vpa *Vpa) MatchesPod(pod *PodState) bool {
	if vpa.ID.Namespace != pod.ID.Namespace {
		return false
	}
	return pod.Labels != nil && vpa.PodSelector.Matches(pod.Labels)
}

// UpdatePodLink marks the Pod as controlled or not-controlled by the VPA
// depending on whether the pod labels match the Vpa pod selector.
// If multiple VPAs match the same Pod, only one of them will effectively
// control the Pod.
func (vpa *Vpa) UpdatePodLink(pod *PodState) bool {
	_, previouslyMatched := pod.MatchingVpas[vpa.ID]
	currentlyMatching := vpa.MatchesPod(pod)

	if previouslyMatched == currentlyMatching {
		return false
	}
	if currentlyMatching {
		// Create links between VPA and pod.
		vpa.Pods[pod.ID] = pod
		pod.MatchingVpas[vpa.ID] = vpa
		if pod.Vpa == nil {
			pod.Vpa = vpa
		}
	} else {
		// Delete the links between VPA and pod.
		delete(vpa.Pods, pod.ID)
		delete(pod.MatchingVpas, vpa.ID)
		if pod.Vpa == vpa {
			pod.Vpa = getAnyVpa(pod.MatchingVpas)
		}
	}
	return true
}

// Returns any VPA from the map or nil if the map is empty.
func getAnyVpa(vpas map[VpaID]*Vpa) *Vpa {
	for _, vpa := range vpas {
		return vpa
	}
	return nil
}
