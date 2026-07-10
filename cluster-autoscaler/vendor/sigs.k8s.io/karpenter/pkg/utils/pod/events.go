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

package pod

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/karpenter/pkg/events"
)

func InvalidDoNotDisruptAnnotationEvent(pod *corev1.Pod, message string) events.Event {
	return events.Event{
		InvolvedObject: pod,
		Type:           corev1.EventTypeWarning,
		Reason:         "InvalidDoNotDisruptAnnotation",
		Message:        fmt.Sprintf("Invalid karpenter.sh/do-not-disrupt annotation: %s, ignoring annotation", message),
		DedupeValues:   []string{string(pod.UID), message},
	}
}

func DoNotDisruptUntilEvent(pod *corev1.Pod, disruptableAt time.Time) events.Event {
	return events.Event{
		InvolvedObject: pod,
		Type:           corev1.EventTypeNormal,
		Reason:         "DoNotDisruptUntil",
		Message:        fmt.Sprintf("The karpenter.sh/do-not-disrupt grace period will elapse at %s", disruptableAt.Format(time.RFC3339)),
		DedupeValues:   []string{string(pod.UID), disruptableAt.Format(time.RFC3339)},
	}
}

func DoNotDisruptGracePeriodElapsedEvent(pod *corev1.Pod) events.Event {
	return events.Event{
		InvolvedObject: pod,
		Type:           corev1.EventTypeNormal,
		Reason:         "DoNotDisruptGracePeriodElapsed",
		Message:        "The karpenter.sh/do-not-disrupt grace period has elapsed, pod is now disruptable",
		DedupeValues:   []string{string(pod.UID)},
	}
}
