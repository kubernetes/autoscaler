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

package common

// Field manager names identifying which VPA component owns a field when writing
// to the API server. Setting these explicitly means the API server records
// ownership against a stable, recognizable name instead of deriving one from the
// User-Agent, and they are a prerequisite for Server-Side Apply.
const (
	// FieldManagerRecommender is used by the VPA recommender.
	FieldManagerRecommender = "vpa-recommender"
	// FieldManagerUpdater is used by the VPA updater.
	FieldManagerUpdater = "vpa-updater"
	// FieldManagerAdmissionController is used by the VPA admission controller.
	FieldManagerAdmissionController = "vpa-admission-controller"
)
