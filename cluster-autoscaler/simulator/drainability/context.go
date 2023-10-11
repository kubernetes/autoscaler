/*
Copyright 2023 The Kubernetes Authors.

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

package drainability

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// DrainContext contains parameters for drainability rules.
type DrainContext struct {
	RemainingPdbTracker pdb.RemainingPdbTracker
	Listers             kube_util.ListerRegistry
	Timestamp           time.Time
}
