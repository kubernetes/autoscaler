/*
Copyright 2020 The Kubernetes Authors.

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

package predicatechecker

import (
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_config_latest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
)

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() (PredicateChecker, error) {
	defaultConfig, err := scheduler_config_latest.Default()
	if err != nil {
		return nil, err
	}
	return NewTestPredicateCheckerWithCustomConfig(defaultConfig)
}

// NewTestPredicateCheckerWithCustomConfig builds test version of PredicateChecker with custom scheduler config.
func NewTestPredicateCheckerWithCustomConfig(schedConfig *config.KubeSchedulerConfiguration) (PredicateChecker, error) {
	// just call out to NewSchedulerBasedPredicateChecker but use fake kubeClient
	fwHandle, err := framework.NewHandle(informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0), schedConfig)
	if err != nil {
		return nil, err
	}
	return NewSchedulerBasedPredicateChecker(fwHandle), nil
}
