/*
Copyright 2022 The Kubernetes Authors.

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

package scheduling

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
)

// HintKey uniquely identifies a pod for the sake of scheduling hints.
type HintKey string

// HintKeyFromPod generates a HintKey for a given pod.
func HintKeyFromPod(pod *apiv1.Pod) HintKey {
	if pod.UID != "" {
		return HintKey(pod.UID)
	}
	return HintKey(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
}

// Hints can be used for tracking past scheduling decisions. It is
// essentially equivalent to a map, with the ability to replace whole
// generations of keys. See DropOld() for more information.
type Hints struct {
	current map[HintKey]string
	old     map[HintKey]string
}

// NewHints returns a new Hints object.
func NewHints() *Hints {
	return &Hints{
		current: make(map[HintKey]string),
		old:     make(map[HintKey]string),
	}
}

// Get retrieves a hinted node name for a given key.
func (h *Hints) Get(hk HintKey) (string, bool) {
	if v, ok := h.current[hk]; ok {
		return v, ok
	}
	v, ok := h.old[hk]
	return v, ok
}

// Set updates a hinted node name for a given key.
func (h *Hints) Set(hk HintKey, nodeName string) {
	h.current[hk] = nodeName
}

// DropOld cleans up old keys. All keys are considered old if they were added
// before the previous call to DropOld().
func (h *Hints) DropOld() {
	oldHintsCount := len(h.old)
	h.old = h.current
	h.current = make(map[HintKey]string, oldHintsCount)
}
