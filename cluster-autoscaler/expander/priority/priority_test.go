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
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

const (
	configOKMessage = "Normal PriorityConfigMapReloaded Successfully reloaded priority " +
		"configuration from configmap."
	configWarnGroupNotFoundMessage = "Warning PriorityConfigMapNotMatchedGroup Priority expander: node group " +
		"%s not found in priority expander configuration. The group won't be used."
	configWarnConfigMapDeleted = "Warning PriorityConfigMapDeleted Configmap for priority expander was deleted, " +
		"no updates will be processed until recreated."
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
	eoT2Micro = expander.Option{
		Debug: "t2.micro",
		NodeGroup: &testNodeGroup{
			id: "my-asg.t2.micro",
		},
	}
	eoT2Large = expander.Option{
		Debug: "t2.large",
		NodeGroup: &testNodeGroup{
			id: "my-asg.t2.large",
		},
	}
	eoT3Large = expander.Option{
		Debug: "t3.large",
		NodeGroup: &testNodeGroup{
			id: "my-asg.t3.large",
		},
	}
	eoM44XLarge = expander.Option{
		Debug: "m4.4xlarge",
		NodeGroup: &testNodeGroup{
			id: "my-asg.m4.4xlarge",
		},
	}
)

func getStrategyInstance(t *testing.T, config string) (expander.Strategy, chan watch.Event, *testEventRecorder, error) {
	c := make(chan watch.Event)

	r := newTestRecorder()
	s, err := NewStrategy(config, c, r)
	assert.Nil(t, err)
	return s, c, r, err
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
	s, c, r, _ := getStrategyInstance(t, oneEntryConfig)
	ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)
	assert.Equal(t, *ret, eoT2Large)

	event := <-r.recorder.Events
	assert.EqualValues(t, configOKMessage, event)
	for _, group := range []string{eoT3Large.NodeGroup.Id(), eoM44XLarge.NodeGroup.Id()} {
		event = <-r.recorder.Events
		assert.EqualValues(t, fmt.Sprintf(configWarnGroupNotFoundMessage, group), event)
	}

	c <- watch.Event{
		Type: watch.Modified,
		Object: &apiv1.ConfigMap{
			Data: map[string]string{
				ConfigMapKey: config,
			},
		},
	}
	priority := s.(*priority)
	for {
		if priority.okConfigUpdates == 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	ret = s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)
	assert.Equal(t, *ret, eoM44XLarge)
	event = <-r.recorder.Events
	assert.EqualValues(t, configOKMessage, event)
}

func TestPriorityExpanderCorrecltySkipsBadChangeConfig(t *testing.T) {
	s, c, r, _ := getStrategyInstance(t, oneEntryConfig)

	event := <-r.recorder.Events
	assert.EqualValues(t, configOKMessage, event)

	c <- watch.Event{
		Type:   watch.Deleted,
		Object: &apiv1.ConfigMap{},
	}
	priority := s.(*priority)
	for {
		if priority.badConfigUpdates == 1 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	assert.Equal(t, 1, priority.okConfigUpdates)
	event = <-r.recorder.Events
	assert.EqualValues(t, configWarnConfigMapDeleted, event)

	ret := s.BestOption([]expander.Option{eoT2Large, eoT3Large, eoM44XLarge}, nil)

	assert.Equal(t, *ret, eoT2Large)
	for _, group := range []string{eoT3Large.NodeGroup.Id(), eoM44XLarge.NodeGroup.Id()} {
		event = <-r.recorder.Events
		assert.EqualValues(t, fmt.Sprintf(configWarnGroupNotFoundMessage, group), event)
	}
}

func TestPriorityExpanderFailsToStartWithEmptyConfig(t *testing.T) {
	_, err := NewStrategy("", nil, &utils.LogEventRecorder{})
	assert.NotNil(t, err)
}

func TestPriorityExpanderFailsToStartWithBadConfig(t *testing.T) {
	_, err := NewStrategy("not_really_yaml: 34 : 43", nil, &utils.LogEventRecorder{})
	assert.NotNil(t, err)
}

type testNodeGroup struct {
	id    string
	debug string
}

func (t *testNodeGroup) MaxSize() int {
	return 10
}

func (t *testNodeGroup) MinSize() int {
	return 0
}

func (t *testNodeGroup) TargetSize() (int, error) {
	return 5, nil
}

func (t *testNodeGroup) IncreaseSize(delta int) error {
	return nil
}

func (t *testNodeGroup) DeleteNodes([]*apiv1.Node) error {
	return nil
}

func (t *testNodeGroup) DecreaseTargetSize(delta int) error {
	return nil
}

func (t *testNodeGroup) Id() string {
	return t.id
}

func (t *testNodeGroup) Debug() string {
	return t.debug
}

func (t *testNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return nil, nil
}

func (t *testNodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, nil
}

func (t *testNodeGroup) Exist() bool {
	return true
}

func (t *testNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, nil
}

func (t *testNodeGroup) Delete() error {
	return nil
}

func (t *testNodeGroup) Autoprovisioned() bool {
	return false
}

type testEventRecorder struct {
	recorder     *record.FakeRecorder
	statusObject runtime.Object
}

func newTestRecorder() *testEventRecorder {
	return &testEventRecorder{
		recorder:     record.NewFakeRecorder(100),
		statusObject: &apiv1.ConfigMap{},
	}
}

func (ler *testEventRecorder) Event(eventtype, reason, message string) {
	ler.recorder.Event(ler.statusObject, eventtype, reason, message)
}

func (ler *testEventRecorder) Eventf(eventtype, reason, message string, args ...interface{}) {
	ler.recorder.Eventf(ler.statusObject, eventtype, reason, message, args...)
}
