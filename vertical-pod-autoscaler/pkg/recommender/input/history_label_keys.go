/*
Copyright The Kubernetes Authors.

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

package input

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	invalidLabelNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
	lowerThenUpperRE       = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

// sanitizeLabelKey returns the name under which kube-state-metrics exports a
// pod label key, without its "label_" prefix: every rune outside [a-zA-Z0-9_]
// becomes "_", then the result is converted to snake case. The prefix keeps a
// leading digit valid, so it needs no special handling.
func sanitizeLabelKey(key string) string {
	sanitized := invalidLabelNameCharRE.ReplaceAllString(key, "_")
	return strings.ToLower(lowerThenUpperRE.ReplaceAllString(sanitized, "${1}_${2}"))
}

// selectorLabelKeysByNamespace returns, per namespace, the label keys used in
// the pod selectors of the VPAs in that namespace, indexed by their sanitized
// form. When several keys sanitize to the same name, a key that is its own
// sanitized form is kept, so it is never renamed; otherwise the first in sorted
// order is kept.
func selectorLabelKeysByNamespace(vpas map[model.VpaID]*model.Vpa) map[string]map[string]string {
	keysByNamespace := make(map[string]map[string]bool)
	for vpaID, vpa := range vpas {
		if vpa.PodSelector == nil {
			continue
		}
		requirements, _ := vpa.PodSelector.Requirements()
		for _, requirement := range requirements {
			if keysByNamespace[vpaID.Namespace] == nil {
				keysByNamespace[vpaID.Namespace] = make(map[string]bool)
			}
			keysByNamespace[vpaID.Namespace][requirement.Key()] = true
		}
	}

	result := make(map[string]map[string]string, len(keysByNamespace))
	for namespace, keys := range keysByNamespace {
		bySanitized := make(map[string]string, len(keys))
		for key := range keys {
			if sanitizeLabelKey(key) == key {
				bySanitized[key] = key
			}
		}
		for _, key := range slices.Sorted(maps.Keys(keys)) {
			sanitized := sanitizeLabelKey(key)
			if sanitized == key {
				continue
			}
			if kept, found := bySanitized[sanitized]; found {
				klog.V(4).InfoS("Selector label keys sanitize to the same name, keeping the first", "namespace", namespace, "sanitized", sanitized, "kept", kept, "ignored", key)
				continue
			}
			bySanitized[sanitized] = key
		}
		result[namespace] = bySanitized
	}
	return result
}

// restoreLabelKeys returns the labels of a pod read from history with each key
// that is the sanitized form of a selector label key renamed to that selector
// key, so the pod matches the same VPA selectors as the live pod would.
// Other keys, and keys whose selector key the pod already has, are kept
// unchanged.
func restoreLabelKeys(podLabels map[string]string, selectorKeys map[string]string) map[string]string {
	if len(selectorKeys) == 0 {
		return podLabels
	}
	restored := make(map[string]string, len(podLabels))
	for key, value := range podLabels {
		if original, found := selectorKeys[key]; found {
			if _, exact := podLabels[original]; !exact {
				key = original
			}
		}
		restored[key] = value
	}
	return restored
}
