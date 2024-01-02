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

package metrics

import (
	tencent_errors "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/errors"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	caNamespace = "cluster_autoscaler"
)

var (
	cloudAPIInvokedCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "invoked_cloudapi_total",
			Help:      "Number of cloudapi invoked by Node Autoprovisioning.",
		}, []string{"service", "ops"},
	)

	cloudAPIInvokedErrorCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "invoked_cloudapi_error_total",
			Help:      "Number of errors that cloudapi invoked by Node Autoprovisioning.",
		}, []string{"service", "ops", "code"},
	)
)

func init() {
	legacyregistry.MustRegister(cloudAPIInvokedCount)
	legacyregistry.MustRegister(cloudAPIInvokedErrorCount)
}

// RegisterCloudAPIInvoked registers cloudapi invoked
func RegisterCloudAPIInvoked(service string, ops string, err error) {
	cloudAPIInvokedCount.WithLabelValues(service, ops).Inc()

	if err != nil {
		if e, ok := err.(*tencent_errors.TencentCloudSDKError); ok {
			RegisterCloudAPIInvokedError("as", "DescribeAutoScalingGroups", e.Code)
		}
	}
}

// RegisterCloudAPIInvokedError registers error in cloudapi invoked
func RegisterCloudAPIInvokedError(service string, ops string, code string) {
	cloudAPIInvokedErrorCount.WithLabelValues(service, ops, code).Inc()
}
