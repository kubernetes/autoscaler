//go:build e2e
// +build e2e

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

package e2e

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/e2e-framework/klient"
)

func withKwokToleration(pod *corev1.Pod) {
	for _, t := range pod.Spec.Tolerations {
		if t.Key == "kwok-provider" {
			return
		}
	}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations, corev1.Toleration{
		Key:      "kwok-provider",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	})
}

func allPodsScheduled(client klient.Client, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		pods := &corev1.PodList{}
		err := client.Resources(namespace).List(ctx, pods)
		if err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			if pod.Spec.NodeName == "" {
				return false, nil
			}
		}
		return true, nil
	}
}
