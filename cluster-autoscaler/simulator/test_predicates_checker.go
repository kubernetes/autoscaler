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

import scheduler_predicates "k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"

// PredicateInfo assigns a name to a predicate
type PredicateInfo struct {
	Name      string
	Predicate scheduler_predicates.FitPredicate
}

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() PredicateChecker {
	// TODO(scheduler_framework)
	return nil
}

// NewCustomTestPredicateChecker builds test version of PredicateChecker with additional predicates.
// Helps with benchmarking different ordering of predicates.
func NewCustomTestPredicateChecker(predicateInfos []PredicateInfo) PredicateChecker {
	// TODO(scheduler_framework)
	return nil
}
