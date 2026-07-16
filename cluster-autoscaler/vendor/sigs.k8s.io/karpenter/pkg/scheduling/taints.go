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

package scheduling

import (
	"fmt"
	"strings"

	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	cloudproviderapi "k8s.io/cloud-provider/api"

	"sigs.k8s.io/karpenter/pkg/operator/logging"
	"sigs.k8s.io/karpenter/pkg/utils/pretty"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// KnownEphemeralTaints are taints that are expected to be added to a node while it's initializing
// If the node is a Karpenter-managed node, we don't consider these taints while the node is uninitialized
// since we expect these taints to eventually be removed
var KnownEphemeralTaints = []corev1.Taint{
	{Key: corev1.TaintNodeNotReady, Effect: corev1.TaintEffectNoSchedule},
	{Key: corev1.TaintNodeNotReady, Effect: corev1.TaintEffectNoExecute},
	{Key: corev1.TaintNodeUnreachable, Effect: corev1.TaintEffectNoSchedule},
	{Key: cloudproviderapi.TaintExternalCloudProvider, Effect: corev1.TaintEffectNoSchedule, Value: "true"},
	v1.UnregisteredNoExecuteTaint,
}

// KnownEphemeralTaintKeyPrefixes are taint key prefixes that are expected to be ephemeral.
// Used for taint families whose key has a controller-managed suffix (e.g. Node Readiness
// Controller adds taints with keys like "readiness.k8s.io/<rule-name>"). See #2934.
var KnownEphemeralTaintKeyPrefixes = []string{
	// https://kubernetes.io/blog/2026/02/03/introducing-node-readiness-controller/
	"readiness.k8s.io/",
}

// IsKnownEphemeralTaint reports whether the given taint is in KnownEphemeralTaints
// (exact match on key/value/effect) or has a key matching any prefix in
// KnownEphemeralTaintKeyPrefixes.
func IsKnownEphemeralTaint(taint *corev1.Taint) bool {
	if taint == nil {
		return false
	}
	for i := range KnownEphemeralTaints {
		if KnownEphemeralTaints[i].MatchTaint(taint) {
			return true
		}
	}
	for _, prefix := range KnownEphemeralTaintKeyPrefixes {
		if strings.HasPrefix(taint.Key, prefix) {
			return true
		}
	}
	return false
}

// Taints is a decorated alias type for []corev1.Taint
type Taints []corev1.Taint

// ToleratesPod returns true if the pod tolerates all taints.
func (ts Taints) ToleratesPod(pod *corev1.Pod) error {
	return ts.Tolerates(pod.Spec.Tolerations)
}

// Tolerates returns true if the toleration slice tolerate all taints.
func (ts Taints) Tolerates(tolerations []corev1.Toleration) (errs error) {
	for i := range ts {
		taint := ts[i]
		tolerates := false
		for _, t := range tolerations {
			tolerates = tolerates || t.ToleratesTaint(logging.NopLogger, &taint, true)
		}
		if !tolerates {
			errs = multierr.Append(errs, serrors.Wrap(fmt.Errorf("did not tolerate taint"), "taint", pretty.Taint(taint)))
		}
	}
	return errs
}

// Merge merges in taints with the passed in taints.
func (ts Taints) Merge(with Taints) Taints {
	res := lo.Map(ts, func(t corev1.Taint, _ int) corev1.Taint {
		return t
	})
	for _, taint := range with {
		if _, ok := lo.Find(res, func(t corev1.Taint) bool {
			return taint.MatchTaint(&t)
		}); !ok {
			res = append(res, taint)
		}
	}
	return res
}
