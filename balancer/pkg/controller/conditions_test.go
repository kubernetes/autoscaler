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

package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	balancerapi "k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
)

func TestSetConditionsBasedOnProcessError(t *testing.T) {
	balancer := newBalancer(5)
	now := time.Now()
	setConditionsBasedOnError(balancer, nil, now)

	assert.Len(t, balancer.Status.Conditions, 1)
	assert.Equal(t, balancerapi.BalancerConditionRunning, balancer.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, balancer.Status.Conditions[0].Status)
	assert.Equal(t, now, balancer.Status.Conditions[0].LastTransitionTime.Time)

	setConditionsBasedOnError(balancer, nil, now.Add(time.Minute))
	assert.Len(t, balancer.Status.Conditions, 1)
	assert.Equal(t, now, balancer.Status.Conditions[0].LastTransitionTime.Time)

	now = now.Add(time.Hour)

	setConditionsBasedOnError(balancer, newBalancerError(ScaleSubresourcePolling, fmt.Errorf("bum")), now)
	assert.Len(t, balancer.Status.Conditions, 1)
	assert.Equal(t, balancerapi.BalancerConditionRunning, balancer.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, balancer.Status.Conditions[0].Status)
	assert.Equal(t, now, balancer.Status.Conditions[0].LastTransitionTime.Time)
}
