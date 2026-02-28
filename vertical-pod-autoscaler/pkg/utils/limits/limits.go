/*
Copyright The Kubernetes Authors.

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

package limits

import apiv1 "k8s.io/api/core/v1"

// GlobalMaxAllowed holds global maximum caps for recommendations, i.e. the constraints defined by the recommender flags.
// The VPA-level maxAllowed takes precedence.
// The flags that set minimum constraints are currently defined in
// another part of the code.
type GlobalMaxAllowed struct {
	// Container-level maximums apply to per-Container recommendations.
	Container apiv1.ResourceList
	// Pod-level maximums apply to the Pod-level recommendations
	Pod apiv1.ResourceList
}
