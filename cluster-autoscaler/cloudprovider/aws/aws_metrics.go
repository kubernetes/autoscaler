/*
Copyright 2021 The Kubernetes Authors.

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

package aws

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/awserr"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	caNamespace = "cluster_autoscaler"
)

var (
	/**** Metrics related to AWS API usage ****/
	requestSummary = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Namespace: caNamespace,
			Name:      "aws_request_duration_seconds",
			Help:      "Time taken by AWS requests, by method and status code, in seconds",
			Buckets:   []float64{0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 60.0},
		}, []string{"endpoint", "status"},
	)
)

// RegisterMetrics registers all AWS metrics.
func RegisterMetrics() {
	legacyregistry.MustRegister(requestSummary)
}

// observeAWSRequest records AWS API calls counts and durations
func observeAWSRequest(endpoint string, err error, start time.Time) {
	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
		if awsErr, ok := err.(awserr.Error); ok {
			status = awsErr.Code()
		}
	}
	requestSummary.WithLabelValues(endpoint, status).Observe(duration)
}
