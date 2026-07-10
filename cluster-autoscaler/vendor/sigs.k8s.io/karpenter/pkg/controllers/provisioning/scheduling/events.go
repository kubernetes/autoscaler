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
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/flowcontrol"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/events"
)

// PodNominationRateLimiter is a pointer so it rate-limits across events
var PodNominationRateLimiter = flowcontrol.NewTokenBucketRateLimiter(5, 10)

func NominatePodEvent(pod *corev1.Pod, node *corev1.Node, nodeClaim *v1.NodeClaim) events.Event {
	var info []string
	if nodeClaim != nil {
		info = append(info, fmt.Sprintf("nodeclaim/%s", nodeClaim.GetName()))
	}
	if node != nil {
		info = append(info, fmt.Sprintf("node/%s", node.Name))
	}
	return events.Event{
		InvolvedObject: pod,
		Type:           corev1.EventTypeNormal,
		Reason:         events.Nominated,
		Message:        fmt.Sprintf("Pod should schedule on: %s", strings.Join(info, ", ")),
		DedupeValues:   []string{string(pod.UID)},
		RateLimiter:    PodNominationRateLimiter,
	}
}

func NoCompatibleInstanceTypes(np *v1.NodePool, minValuesIncompatibleError bool) events.Event {
	return events.Event{
		InvolvedObject: np,
		Type:           corev1.EventTypeWarning,
		Reason:         events.NoCompatibleInstanceTypes,
		Message:        lo.Ternary(minValuesIncompatibleError, "NodePool requirements filtered out all compatible available instance types due to minValues incompatibility", "NodePool requirements filtered out all compatible available instance types"),
		DedupeValues:   []string{string(np.UID)},
		DedupeTimeout:  1 * time.Minute,
	}
}

func PodFailedToScheduleEvent(pod *corev1.Pod, err error) events.Event {
	return events.Event{
		InvolvedObject: pod,
		Type:           corev1.EventTypeWarning,
		Reason:         events.FailedScheduling,
		Message:        fmt.Sprintf("Failed to schedule pod, %s", err),
		DedupeValues:   []string{string(pod.UID)},
		DedupeTimeout:  5 * time.Minute,
	}
}
