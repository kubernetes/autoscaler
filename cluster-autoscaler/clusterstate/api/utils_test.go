/*
Copyright 2017 The Kubernetes Authors.

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

package api

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareConditions() (health, scaleUp ClusterAutoscalerCondition) {
	healthCondition := ClusterAutoscalerCondition{
		Type:    ClusterAutoscalerHealth,
		Status:  ClusterAutoscalerHealthy,
		Message: "HEALTH_MESSAGE"}
	scaleUpCondition := ClusterAutoscalerCondition{
		Type:    ClusterAutoscalerScaleUp,
		Status:  ClusterAutoscalerNotNeeded,
		Message: "SCALE_UP_MESSAGE"}
	return healthCondition, scaleUpCondition
}

func TestGetStringForEmptyStatus(t *testing.T) {
	var empty ClusterAutoscalerStatus
	assert.Regexp(t, regexp.MustCompile("\\s*Health:\\s*<unknown>"), empty.GetReadableString())
}

func TestGetStringNothingGoingOn(t *testing.T) {
	var status ClusterAutoscalerStatus
	healthCondition, scaleUpCondition := prepareConditions()
	status.ClusterwideConditions = append(status.ClusterwideConditions, healthCondition)
	status.ClusterwideConditions = append(status.ClusterwideConditions, scaleUpCondition)

	// Make sure everything is printed
	result := status.GetReadableString()
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%v:\\s*%v", ClusterAutoscalerHealth, ClusterAutoscalerHealthy)), result)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%v.*HEALTH_MESSAGE", ClusterAutoscalerHealth)), result)
	assert.NotRegexp(t, regexp.MustCompile(fmt.Sprintf("%v.*SCALE_UP_MESSAGE", ClusterAutoscalerHealth)), result)
	assert.NotRegexp(t, regexp.MustCompile("NodeGroups"), result)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%v:\\s*%v", ClusterAutoscalerScaleUp, ClusterAutoscalerNotNeeded)), result)

	// Check if reordering fields doesn't change output
	var reorderedStatus ClusterAutoscalerStatus
	reorderedStatus.ClusterwideConditions = append(status.ClusterwideConditions, scaleUpCondition)
	reorderedStatus.ClusterwideConditions = append(status.ClusterwideConditions, healthCondition)
	reorderedResult := reorderedStatus.GetReadableString()
	assert.Equal(t, result, reorderedResult)
}

func TestGetStringScalingUp(t *testing.T) {
	var status ClusterAutoscalerStatus
	healthCondition, scaleUpCondition := prepareConditions()
	scaleUpCondition.Status = ClusterAutoscalerInProgress
	status.ClusterwideConditions = append(status.ClusterwideConditions, healthCondition)
	status.ClusterwideConditions = append(status.ClusterwideConditions, scaleUpCondition)
	result := status.GetReadableString()
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%v:\\s*%v.*SCALE_UP_MESSAGE", ClusterAutoscalerScaleUp, ClusterAutoscalerInProgress)), result)
}

func TestGetStringNodeGroups(t *testing.T) {
	var status ClusterAutoscalerStatus
	healthCondition, scaleUpCondition := prepareConditions()
	status.ClusterwideConditions = append(status.ClusterwideConditions, healthCondition)
	status.ClusterwideConditions = append(status.ClusterwideConditions, scaleUpCondition)
	var ng1, ng2 NodeGroupStatus
	ng1.ProviderID = "ng1"
	ng1.Conditions = status.ClusterwideConditions
	ng2.ProviderID = "ng2"
	ng2.Conditions = status.ClusterwideConditions
	status.NodeGroupStatuses = append(status.NodeGroupStatuses, ng1)
	status.NodeGroupStatuses = append(status.NodeGroupStatuses, ng2)
	result := status.GetReadableString()
	assert.Regexp(t, regexp.MustCompile("(?ms)NodeGroups:.*Name:\\s*ng1"), result)
	assert.Regexp(t, regexp.MustCompile("(?ms)NodeGroups:.*Name:\\s*ng2"), result)
}
