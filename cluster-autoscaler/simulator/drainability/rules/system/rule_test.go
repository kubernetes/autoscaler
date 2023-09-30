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

package system

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
)

func TestDrain(t *testing.T) {
	var (
		testTime = time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
		replicas = int32(5)

		rc = apiv1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rc",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc",
			},
			Spec: apiv1.ReplicationControllerSpec{
				Replicas: &replicas,
			},
		}

		rcPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "default",
				OwnerReferences: GenerateOwnerReferences(rc.Name, "ReplicationController", "core/v1", ""),
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		kubeSystemRc = apiv1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rc",
				Namespace: "kube-system",
				SelfLink:  "api/v1/namespaces/kube-system/replicationcontrollers/rc",
			},
			Spec: apiv1.ReplicationControllerSpec{
				Replicas: &replicas,
			},
		}

		kubeSystemRcPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "kube-system",
				OwnerReferences: GenerateOwnerReferences(kubeSystemRc.Name, "ReplicationController", "core/v1", ""),
				Labels: map[string]string{
					"k8s-app": "bar",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		emptyPDB = &policyv1.PodDisruptionBudget{}

		kubeSystemPDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "bar",
					},
				},
			},
		}

		kubeSystemFakePDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "foo",
					},
				},
			},
		}

		defaultNamespacePDB = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "PDB-managed pod",
					},
				},
			},
		}

		kubeSystemFailedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyNever,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		kubeSystemTerminalPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyOnFailure,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodSucceeded,
			},
		}

		kubeSystemEvictedPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
			},
			Spec: apiv1.PodSpec{
				NodeName:      "node",
				RestartPolicy: apiv1.RestartPolicyAlways,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodFailed,
			},
		}

		kubeSystemSafePod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "kube-system",
				Annotations: map[string]string{
					drain.PodSafeToEvictKey: "true",
				},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node",
			},
		}

		zeroGracePeriod              = int64(0)
		kubeSystemLongTerminatingPod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "kube-system",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &zeroGracePeriod,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
			},
		}

		extendedGracePeriod                                 = int64(6 * 60) // 6 minutes
		kubeSystemLongTerminatingPodWithExtendedGracePeriod = &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "kube-system",
				DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * time.Duration(extendedGracePeriod) * time.Second)},
			},
			Spec: apiv1.PodSpec{
				NodeName:                      "node",
				RestartPolicy:                 apiv1.RestartPolicyOnFailure,
				TerminationGracePeriodSeconds: &extendedGracePeriod,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodUnknown,
			},
		}
	)

	for _, test := range []struct {
		desc string
		pod  *apiv1.Pod
		rcs  []*apiv1.ReplicationController
		rss  []*appsv1.ReplicaSet
		pdbs []*policyv1.PodDisruptionBudget

		wantReason drain.BlockingPodReason
		wantError  bool
	}{
		{
			desc: "kube-system pod with PodSafeToEvict annotation",
			pod:  kubeSystemSafePod,
		},
		{
			desc: "empty PDB with RC-managed pod",
			pod:  rcPod,
			rcs:  []*apiv1.ReplicationController{&rc},
			pdbs: []*policyv1.PodDisruptionBudget{emptyPDB},
		},
		{
			desc: "kube-system PDB with matching kube-system pod",
			pod:  kubeSystemRcPod,
			rcs:  []*apiv1.ReplicationController{&kubeSystemRc},
			pdbs: []*policyv1.PodDisruptionBudget{kubeSystemPDB},
		},
		{
			desc:       "kube-system PDB with non-matching kube-system pod",
			pod:        kubeSystemRcPod,
			rcs:        []*apiv1.ReplicationController{&kubeSystemRc},
			pdbs:       []*policyv1.PodDisruptionBudget{kubeSystemFakePDB},
			wantReason: drain.UnmovableKubeSystemPod,
			wantError:  true,
		},
		{
			desc: "kube-system PDB with default namespace pod",
			pod:  rcPod,
			rcs:  []*apiv1.ReplicationController{&rc},
			pdbs: []*policyv1.PodDisruptionBudget{kubeSystemPDB},
		},
		{
			desc:       "default namespace PDB with matching labels kube-system pod",
			pod:        kubeSystemRcPod,
			rcs:        []*apiv1.ReplicationController{&kubeSystemRc},
			pdbs:       []*policyv1.PodDisruptionBudget{defaultNamespacePDB},
			wantReason: drain.UnmovableKubeSystemPod,
			wantError:  true,
		},
		{
			desc: "kube-system failed pod",
			pod:  kubeSystemFailedPod,
		},
		{
			desc: "kube-system terminal pod",
			pod:  kubeSystemTerminalPod,
		},
		{
			desc: "kube-system evicted pod",
			pod:  kubeSystemEvictedPod,
		},
		{
			desc: "kube-system pod with PodSafeToEvict annotation",
			pod:  kubeSystemSafePod,
		},
		{
			desc: "kube-system long terminating pod with 0 grace period",
			pod:  kubeSystemLongTerminatingPod,
		},
		{
			desc: "kube-system long terminating pod with extended grace period",
			pod:  kubeSystemLongTerminatingPodWithExtendedGracePeriod,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			tracker := pdb.NewBasicRemainingPdbTracker()
			tracker.SetPdbs(test.pdbs)

			drainCtx := &drainability.DrainContext{
				RemainingPdbTracker: tracker,
				DeleteOptions: options.NodeDeleteOptions{
					SkipNodesWithSystemPods: true,
				},
				Timestamp: testTime,
			}
			status := New().Drainable(drainCtx, test.pod)
			assert.Equal(t, test.wantReason, status.BlockingReason)
			assert.Equal(t, test.wantError, status.Error != nil)
		})
	}
}
