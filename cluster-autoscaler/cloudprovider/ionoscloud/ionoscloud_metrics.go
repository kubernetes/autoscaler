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

package ionoscloud

import (
	"strconv"

	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	caNamespace = "cluster_autoscaler"
)

var requestTotal = k8smetrics.NewCounterVec(
	&k8smetrics.CounterOpts{
		Namespace: caNamespace,
		Name:      "ionoscloud_api_request_total",
		Help:      "Counter of IonosCloud API requests for each action and response status.",
	}, []string{"action", "status"},
)

// RegisterMetrics registers all IonosCloud metrics.
func RegisterMetrics() {
	legacyregistry.MustRegister(requestTotal)
}

func registerRequest(action string, resp *ionos.APIResponse, err error) {
	status := "success"
	if err != nil {
		status = "error"
		if resp.Response != nil {
			status = strconv.Itoa(resp.Response.StatusCode)
		}
	}
	requestTotal.WithLabelValues(action, status).Inc()
}
