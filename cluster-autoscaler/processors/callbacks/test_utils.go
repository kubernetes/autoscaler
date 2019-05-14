/*
Copyright 2019 The Kubernetes Authors.

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

package callbacks

// TestProcessorCallbacks is test implementation of ProcessorCallbacks
type TestProcessorCallbacks struct {
	// ScaleDownDisabledForLoop marks if scaledown should be disabled for loop
	ScaleDownDisabledForLoop bool
	// ExtraValues stores values set by GetExtraValue
	ExtraValues map[string]interface{}
}

// NewTestProcessorCallbacks creates new instance of TestProcessorCallbacks
func NewTestProcessorCallbacks() *TestProcessorCallbacks {
	callbacks := &TestProcessorCallbacks{}
	callbacks.Reset()
	return callbacks
}

// Reset resets TestProcessorCallbacks
func (callbacks *TestProcessorCallbacks) Reset() {
	callbacks.ScaleDownDisabledForLoop = false
	callbacks.ExtraValues = make(map[string]interface{})
}

// DisableScaleDownForLoop is implementation of ProcessorCallbacks.DisableScaleDownForLoop
func (callbacks *TestProcessorCallbacks) DisableScaleDownForLoop() {
	callbacks.ScaleDownDisabledForLoop = true
}

// SetExtraValue is implementation of ProcessorCallbacks.SetExtraValue
func (callbacks *TestProcessorCallbacks) SetExtraValue(key string, value interface{}) {
	callbacks.ExtraValues[key] = value
}

// GetExtraValue is implementation of ProcessorCallbacks.GetExtraValue
func (callbacks *TestProcessorCallbacks) GetExtraValue(key string) (value interface{}, found bool) {
	value, found = callbacks.ExtraValues[key]
	return
}
