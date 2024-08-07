/*
Copyright 2024 The Kubernetes Authors.

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

package podinjection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
)

func TestDesiredReplicasFromJob(t *testing.T) {
	one := int32(1)
	five := int32(5)
	ten := int32(10)

	testCases := []struct {
		name         string
		job          *batchv1.Job
		wantReplicas int
	}{
		{
			name:         "No parallel jobs - Parallelism and completion not defined",
			job:          &batchv1.Job{},
			wantReplicas: 1,
		},
		{
			name: "No parallel jobs - Parallelism and completion set to 1",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Parallelism: &one,
					Completions: &one,
				},
			},
			wantReplicas: 1,
		},
		{
			name: "Parallel Jobs with a fixed completion count: incomplete pods less than parallelism",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Completions: &ten,
					Parallelism: &five,
				},
				Status: batchv1.JobStatus{
					Succeeded: 6,
				},
			},
			wantReplicas: 4,
		},
		{
			name: "Parallel Jobs with a fixed completion count: incomplete pods more than parallelism",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Completions: &ten,
					Parallelism: &five,
				},
				Status: batchv1.JobStatus{
					Succeeded: 2,
				},
			},
			wantReplicas: 5,
		},
		{
			name: "Work queue with succeeded pods",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Parallelism: &five,
				},
				Status: batchv1.JobStatus{
					Succeeded: 2,
				},
			},
			wantReplicas: 0,
		},
		{
			name: "Work queue without succeeded pods",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Parallelism: &five,
				},
				Status: batchv1.JobStatus{
					Succeeded: 0,
				},
			},
			wantReplicas: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			replicas := desiredReplicasFromJob(tc.job)
			assert.Equal(t, tc.wantReplicas, replicas)
		})
	}
}
