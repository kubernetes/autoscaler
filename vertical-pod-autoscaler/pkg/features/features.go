/*
Copyright 2025 The Kubernetes Authors.

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

package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/component-base/featuregate"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
)

const (
	// Every feature gate should add method here following this template:
	//
	// // alpha: v1.X
	// // components: admission-controller, recommender, updater
	//
	// MyFeature featuregate.Feature = "MyFeature".
	//
	// Feature gates should be listed in alphabetical, case-sensitive
	// (upper before any lower case character) order. This reduces the risk
	// of code conflicts because changes are more likely to be scattered
	// across the file.
	//
	// In each feature gate description, you must specify "components".
	// The feature must be enabled by the --feature-gates argument on each listed component.

	// alpha: v1.4.0
	// components: admission-controller, updater

	// InPlaceOrRecreate enables the InPlaceOrRecreate update mode to be used.
	// Requires KEP-1287 InPlacePodVerticalScaling feature-gate to be enabled on the cluster.
	InPlaceOrRecreate featuregate.Feature = "InPlaceOrRecreate"
)

// MutableFeatureGate is a mutable, versioned, global FeatureGate.
var MutableFeatureGate featuregate.MutableVersionedFeatureGate = featuregate.NewFeatureGate()

// Enabled is a helper function for MutableFeatureGate.Enabled(f)
func Enabled(f featuregate.Feature) bool {
	return MutableFeatureGate.Enabled(f)
}

func init() {
	// set the emulation version to align with VPA versioning system
	runtime.Must(MutableFeatureGate.SetEmulationVersion(version.MustParse(common.VerticalPodAutoscalerVersion())))
	runtime.Must(MutableFeatureGate.AddVersioned(defaultVersionedFeatureGates))
}
