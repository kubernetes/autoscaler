/*
Copyright 2026 The Kubernetes Authors.

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

package framework

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
)

func TestNewKarpenterDisabledPluginsSchedulerConfig(t *testing.T) {
	cfg, err := NewKarpenterDisabledPluginsSchedulerConfig(nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Profiles)

	profile := cfg.Profiles[0]
	assert.NotNil(t, profile.Plugins)

	preFilterDisabled := profile.Plugins.PreFilter.Disabled
	disabledNames := make(map[string]bool)
	for _, p := range preFilterDisabled {
		disabledNames[p.Name] = true
	}
	assert.True(t, disabledNames[interpodaffinity.Name], "InterPodAffinity should be disabled in PreFilter")
	assert.True(t, disabledNames[podtopologyspread.Name], "PodTopologySpread should be disabled in PreFilter")

	filterDisabled := profile.Plugins.Filter.Disabled
	disabledNamesFilter := make(map[string]bool)
	for _, p := range filterDisabled {
		disabledNamesFilter[p.Name] = true
	}
	assert.True(t, disabledNamesFilter[interpodaffinity.Name], "InterPodAffinity should be disabled in Filter")
	assert.True(t, disabledNamesFilter[podtopologyspread.Name], "PodTopologySpread should be disabled in Filter")
}
