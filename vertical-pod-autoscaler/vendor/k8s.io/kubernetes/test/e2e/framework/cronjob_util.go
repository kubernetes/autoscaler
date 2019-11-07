/*
Copyright 2017 The Kubernetes Authors.

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

package framework

import (
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imageutils "k8s.io/kubernetes/test/utils/image"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/util/wait"
)


// NewTestCronJob returns a CronJob which does one of several testing behaviors.
func NewTestCronJob(name, schedule string, concurrencyPolicy batchv1beta1.ConcurrencyPolicy,
	command []string, parallelism, completions, backofflimit int32, successfulJobsHistoryLimit *int32,
	failedJobsHistoryLimit *int32) *batchv1beta1.CronJob {
	sj := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "CronJob",
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: concurrencyPolicy,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism:  &parallelism,
					Completions:  &completions,
					BackoffLimit: &backofflimit,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{JobSelectorKey: name},
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyOnFailure,
							Volumes: []v1.Volume{
								{
									Name: "data",
									VolumeSource: v1.VolumeSource{
										EmptyDir: &v1.EmptyDirVolumeSource{},
									},
								},
							},
							Containers: []v1.Container{
								{
									Name:  "c",
									Image: imageutils.GetE2EImage(imageutils.BusyBox),
									VolumeMounts: []v1.VolumeMount{
										{
											MountPath: "/data",
											Name:      "data",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	sj.Spec.SuccessfulJobsHistoryLimit = successfulJobsHistoryLimit
	sj.Spec.FailedJobsHistoryLimit = failedJobsHistoryLimit
	if command != nil {
		sj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command = command
	}
	return sj
}

func WaitForActiveJobs(c clientset.Interface, ns, cronJobName string, active int) error {
	return wait.Poll(Poll, CronJobsWaitTimeout, func() (bool, error) {
		curr, err := getCronJob(c, ns, cronJobName)
		if err != nil {
			return false, err
		}
		return len(curr.Status.Active) >= active, nil
	})
}

func CreateCronJob(c clientset.Interface, ns string, cronJob *batchv1beta1.CronJob) (*batchv1beta1.CronJob, error) {
	return c.BatchV1beta1().CronJobs(ns).Create(cronJob)
}

func getCronJob(c clientset.Interface, ns, name string) (*batchv1beta1.CronJob, error) {
	return c.BatchV1beta1().CronJobs(ns).Get(name, metav1.GetOptions{})
}