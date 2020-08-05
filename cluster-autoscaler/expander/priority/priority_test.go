/*
Copyright 2016 The Kubernetes Authors.

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

package priority

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

const (
	testNamespace                  = "default"
	configWarnGroupNotFoundMessage = "Warning PriorityConfigMapNotMatchedGroup Priority expander: node group " +
		"%s not found in priority expander configuration. The group won't be used."
	configWarnConfigMapEmpty = "Warning PriorityConfigMapInvalid Wrong configuration for priority expander: " +
		"priority configuration in cluster-autoscaler-priority-expander configmap is empty; please provide " +
		"valid configuration. Ignoring update."
	configWarnEmptyMsg = "priority configuration in cluster-autoscaler-priority-expander configmap is empty; please provide valid configuration"
	configWarnParseMsg = "Can't parse YAML with priorities in the configmap"
)

var (
	config = `
5:
  - ".*t2\\.micro.*"
10: 
  - ".*t2\\.large.*"
  - ".*t3\\.large.*"
50: 
  - ".*m4\\.4xlarge.*"
`
	oneEntryConfig = `
10: 
  - ".*t2\\.large.*"
`
	notMatchingConfig = `
5:
  - ".*t\\.micro.*"
10: 
  - ".*t\\.large.*"
`
	wildcardMatchConfig = `
5:
  - ".*"
10:
  - ".t2\\.large.*"
`

	eoT2Micro = expander.Option{
		Debug:     "t2.micro",
		NodeGroup: test.NewTestNodeGroup("my-asg.t2.micro", 10, 1, 1, true, false, "t2.micro", nil, nil),
	}
	eoT2Large = expander.Option{
		Debug:     "t2.large",
		NodeGroup: test.NewTestNodeGroup("my-asg.t2.large", 10, 1, 1, true, false, "t2.large", nil, nil),
	}
	eoT3Large = expander.Option{
		Debug:     "t3.large",
		NodeGroup: test.NewTestNodeGroup("my-asg.t3.large", 10, 1, 1, true, false, "t3.large", nil, nil),
	}
	eoM44XLarge = expander.Option{
		Debug:     "m4.4xlarge",
		NodeGroup: test.NewTestNodeGroup("my-asg.m4.4xlarge", 10, 1, 1, true, false, "m4.4xlarge", nil, nil),
	}
)

func getStrategyInstance(t *testing.T, config string) (expander.Strategy, *record.FakeRecorder, *apiv1.ConfigMap, error) {
	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      PriorityConfigMapName,
		},
		Data: map[string]string{
			ConfigMapKey: config,
		},
	}
	lister, err := kubernetes.NewTestConfigMapLister([]*apiv1.ConfigMap{cm})
	assert.Nil(t, err)
	r := record.NewFakeRecorder(100)
	s, err := NewStrategy(lister.ConfigMaps(testNamespace), r)
	return s, r, cm, err
}

func TestPriorityExpanderCorrecltySelectsSingleMatchingOptionOutOfOne(t *testing.T) {
	s, _, _, _ := getStrategyInstance(t, config)
	ret := s.BestOption([]expander.Option{eoT2Large}, nil)
	assert.Equal(t, *ret, eoT2Large)
}

func TestPriorityExpanderCorrecltySelectsSingleMatchingOptionOutOfMany(t *testing.T) {
	s, _, _, _ := getStrategyInstance(t, config)
	ret := s.BestOption([]expander.Option{eoT2Large, eoM44XLarge}, nil)
	assert.Equal(t, *ret, eoM44XLarge)
}

func TestPriorityExpanderDoesNotFallBackToRandomWhenHigherPriorityMatches(t *testing.T) {
	s, _, _, _ := getStrategyInstance(t, wildcardMatchConfig)
	for i := 0; i < 10; i++ {
		ret := s.BestOption([]expander.Option{eoT2Large, eoT2Micro}, nil)
		assert.Equal(t, *ret, eoT2Large)
	}
}

func TestPriorityExpanderCorrecltySelectsOneOfTwoMatchingOptionsOutOfMany(t *testing.T) {
	s, _, _, _ := getStrategyInstance(t, config)
	for i := 0; i < 10; i++ {
		ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoT2Micro}, nil)
		assert.True(t, ret.NodeGroup.Id() == eoT2Large.NodeGroup.Id() || ret.NodeGroup.Id() == eoT3Large.NodeGroup.Id())
	}
}

func TestPriorityExpanderCorrecltyFallsBackToRandomWhenNoMatches(t *testing.T) {
	s, _, _, _ := getStrategyInstance(t, config)
	for i := 0; i < 10; i++ {
		ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large}, nil)
		assert.True(t, ret.NodeGroup.Id() == eoT2Large.NodeGroup.Id() || ret.NodeGroup.Id() == eoT3Large.NodeGroup.Id())
	}
}

func TestPriorityExpanderCorrecltyHandlesConfigUpdate(t *testing.T) {
	s, r, cm, _ := getStrategyInstance(t, oneEntryConfig)
	ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)
	assert.Equal(t, *ret, eoT2Large)

	var event string
	for _, group := range []string{eoT3Large.NodeGroup.Id(), eoM44XLarge.NodeGroup.Id()} {
		event = <-r.Events
		assert.EqualValues(t, fmt.Sprintf(configWarnGroupNotFoundMessage, group), event)
	}

	cm.Data[ConfigMapKey] = config
	ret = s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)

	priority := s.(*priority)
	assert.Equal(t, 2, priority.okConfigUpdates)
	assert.Equal(t, *ret, eoM44XLarge)
}

func TestPriorityExpanderCorrecltySkipsBadChangeConfig(t *testing.T) {
	s, r, cm, _ := getStrategyInstance(t, oneEntryConfig)
	priority := s.(*priority)
	assert.Equal(t, 0, priority.okConfigUpdates)

	cm.Data[ConfigMapKey] = ""
	ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)

	assert.Equal(t, 1, priority.badConfigUpdates)

	event := <-r.Events
	assert.EqualValues(t, configWarnConfigMapEmpty, event)
	assert.Nil(t, ret)
}
