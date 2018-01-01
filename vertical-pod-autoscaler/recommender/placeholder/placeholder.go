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

package placeholder

import (
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/apis/metrics/v1alpha1"
	"k8s.io/metrics/pkg/client/clientset_generated/clientset/fake"
	resourceclient "k8s.io/metrics/pkg/client/clientset_generated/clientset/typed/metrics/v1alpha1"
)

var (
	_ = metrics.NodeMetrics{}
	_ = &fake.Clientset{}
)

func Nothing(config *rest.Config) {
	_ = resourceclient.NewForConfigOrDie(config)
}
