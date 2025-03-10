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
	"k8s.io/component-base/featuregate"
)

const (
	// Every feature gate should add method here following this template:
	//
	// // alpha: v1.X
	// MyFeature featuregate.Feature = "MyFeature".

	// InPlaceVerticalScaling is a feature gate for enabling the InPlaceOrRecreate update mode to be used.
	// Requires InPlacePodVerticalScaling feature-gate to be enabled on the Kubernetes cluster itself.
	//
	// TODO(maxcao13): fill this in
	// alpha: v...
	// beta: v...
	// GA: v...
	InPlaceVerticalScaling featuregate.Feature = "InPlaceVerticalScaling"
)

var MutableFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()

// defaultVPAFeatureGates consists of all known vertical-pod-autoscaler-specific feature keys.
var defaultVPAFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	InPlaceVerticalScaling: {Default: false, PreRelease: featuregate.Alpha},
}

func Enabled(f featuregate.Feature) bool {
	return MutableFeatureGate.Enabled(f)
}

func init() {
	runtime.Must(MutableFeatureGate.Add(defaultVPAFeatureGates))
}
