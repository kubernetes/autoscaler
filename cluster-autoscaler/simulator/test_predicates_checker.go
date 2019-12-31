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

package simulator

import (
	clientsetfake "k8s.io/client-go/kubernetes/fake"
)

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() (PredicateChecker, error) {
	// just call out to NewSchedulerBasedPredicateChecker but use fake kubeClient
	return NewSchedulerBasedPredicateChecker(clientsetfake.NewSimpleClientset(), make(chan struct{}))
}
