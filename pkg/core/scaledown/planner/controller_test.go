/*
Copyright 2019 The Kubernetes Authors.

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

package planner

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var podLabels = map[string]string{
	"app": "test",
}

func TestReplicasCounter(t *testing.T) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			UID:       types.UID("batch/v1/namespaces/default/jobs/job"),
		},
		Spec: batchv1.JobSpec{
			Parallelism: proto.Int32(3),
			Selector:    metav1.SetAsLabelSelector(podLabels),
		},
		Status: batchv1.JobStatus{Active: 1},
	}
	unsetJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unset_job",
			Namespace: "default",
			UID:       types.UID("batch/v1/namespaces/default/jobs/unset_job"),
		},
	}
	jobWithSucceededReplicas := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "succeeded_job",
			Namespace: "default",
			UID:       types.UID("batch/v1/namespaces/default/jobs/succeeded_job"),
		},
		Spec: batchv1.JobSpec{
			Parallelism: proto.Int32(3),
			Completions: proto.Int32(3),
			Selector:    metav1.SetAsLabelSelector(podLabels),
		},
		Status: batchv1.JobStatus{
			Active:    1,
			Succeeded: 2,
		},
	}
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs",
			Namespace: "default",
			UID:       types.UID("apps/v1/namespaces/default/replicasets/rs"),
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: proto.Int32(1),
			Selector: metav1.SetAsLabelSelector(podLabels),
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas: 1,
		},
	}
	unsetRs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unset_rs",
			Namespace: "default",
			UID:       types.UID("apps/v1/namespaces/default/replicasets/unset_rs"),
		},
	}
	rC := &apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			UID:       types.UID("core/v1/namespaces/default/replicationcontrollers/rc"),
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: proto.Int32(1),
			Selector: podLabels,
		},
		Status: apiv1.ReplicationControllerStatus{
			Replicas: 0,
		},
	}
	sS := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sset",
			Namespace: "default",
			UID:       types.UID("apps/v1/namespaces/default/statefulsets/sset"),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: proto.Int32(3),
			Selector: metav1.SetAsLabelSelector(podLabels),
		},
		Status: appsv1.StatefulSetStatus{
			Replicas: 1,
		},
	}
	rcLister, _ := kube_util.NewTestReplicationControllerLister([]*apiv1.ReplicationController{rC})
	jobLister, _ := kube_util.NewTestJobLister([]*batchv1.Job{job, unsetJob, jobWithSucceededReplicas})
	rsLister, _ := kube_util.NewTestReplicaSetLister([]*appsv1.ReplicaSet{rs, unsetRs})
	ssLister, _ := kube_util.NewTestStatefulSetLister([]*appsv1.StatefulSet{sS})
	listers := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, rcLister, jobLister, rsLister, ssLister)
	testCases := []struct {
		name         string
		ownerRef     metav1.OwnerReference
		wantReplicas replicasInfo
		expectErr    bool
	}{
		{
			name:     "job owner reference",
			ownerRef: ownerRef("Job", job.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 1,
				targetReplicas:  3,
			},
		},
		{
			name:     "job without parallelism owner reference",
			ownerRef: ownerRef("Job", unsetJob.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 0,
				targetReplicas:  1,
			},
		},
		{
			name:     "job with succeeded replicas owner reference",
			ownerRef: ownerRef("Job", jobWithSucceededReplicas.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 1,
				targetReplicas:  1,
			},
		},
		{
			name:     "replica set owner reference",
			ownerRef: ownerRef("ReplicaSet", rs.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 1,
				targetReplicas:  1,
			},
		},
		{
			name:     "replica set without replicas spec specified owner reference",
			ownerRef: ownerRef("ReplicaSet", unsetRs.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 0,
				targetReplicas:  1,
			},
		},
		{
			name:     "replica controller owner reference",
			ownerRef: ownerRef("ReplicationController", rC.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 0,
				targetReplicas:  1,
			},
		},
		{
			name:     "stateful set owner reference",
			ownerRef: ownerRef("StatefulSet", sS.Name),
			wantReplicas: replicasInfo{
				currentReplicas: 0,
				targetReplicas:  3,
			},
		},
		{
			name:      "not existing job owner ref",
			ownerRef:  ownerRef("Job", "j"),
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := newControllerReplicasCalculator(listers)
			res, err := c.getReplicas(tc.ownerRef, "default")
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				if diff := cmp.Diff(tc.wantReplicas, *res, cmp.AllowUnexported(replicasInfo{})); diff != "" {
					t.Errorf("getReplicas() diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func ownerRef(ownerType, ownerName string) metav1.OwnerReference {
	api := ""
	strType := ""
	switch ownerType {
	case "ReplicaSet":
		api = "apps/v1"
		strType = "replicasets"
	case "StatefulSet":
		api = "apps/v1"
		strType = "statefulsets"
	case "ReplicationController":
		api = "core/v1"
		strType = "replicationcontrollers"
	case "Job":
		api = "batch/v1"
		strType = "jobs"
	}
	return test.GenerateOwnerReferences(ownerName, ownerType, api, types.UID(fmt.Sprintf("%s/namespaces/default/%s/%s", api, strType, ownerName)))[0]
}
