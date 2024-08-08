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

package podinjection

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/klog/v2"
)

func createJobControllers(ctx *context.AutoscalingContext) []controller {
	var controllers []controller
	jobs, err := ctx.ListerRegistry.JobLister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list jobs: %v", err)
	}
	for _, job := range jobs {
		controllers = append(controllers, controller{uid: job.UID, desiredReplicas: desiredReplicasFromJob(job)})
	}
	return controllers
}

func desiredReplicasFromJob(job *batchv1.Job) int {
	parallelism := 1
	if job.Spec.Parallelism != nil {
		parallelism = int(*(job.Spec.Parallelism))
	}
	if workQueue(job) {
		if job.Status.Succeeded > 0 {
			return 0
		} else {
			return parallelism
		}
	}

	completion := 1
	if job.Spec.Completions != nil {
		completion = int(*(job.Spec.Completions))
	}
	incomplete := completion - int(job.Status.Succeeded)
	desiredReplicas := min(incomplete, parallelism)
	return desiredReplicas
}

func workQueue(job *batchv1.Job) bool {
	return (job.Spec.Completions == nil || *(job.Spec.Completions) == 1) && job.Spec.Parallelism != nil && *(job.Spec.Parallelism) >= 0
}
