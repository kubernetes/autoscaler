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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
)

func setConditionsBasedOnError(balancer *v1alpha1.Balancer, err *BalancerError, now time.Time) {
	if err == nil {
		balancer.Status.Conditions = setConditionInList(balancer.Status.Conditions,
			now, v1alpha1.BalancerConditionRunning, metav1.ConditionTrue,
			"Completed", "Balancer running OK")
	} else {
		balancer.Status.Conditions = setConditionInList(balancer.Status.Conditions,
			now, v1alpha1.BalancerConditionRunning, metav1.ConditionFalse,
			string(err.phase), err.Error())
	}
}

// setConditionInList sets the specific condition type on the given Balancer to
// the specified value with the given reason and message. The message and args
// are treated like a format string. The condition will be added if
// it is not present. The condition will be overwritten if it already exists
// (and LastTransitionState update if condition's status is different)
// A new list will be returned.
func setConditionInList(inputList []metav1.Condition, now time.Time, conditionType string,
	status metav1.ConditionStatus, reason, message string, args ...interface{}) []metav1.Condition {
	resList := inputList
	var existingCond *metav1.Condition
	for i, condition := range resList {
		if condition.Type == conditionType {
			// can't take a pointer to an iteration variable
			existingCond = &resList[i]
			break
		}
	}
	if existingCond == nil {
		resList = append(resList, metav1.Condition{
			Type: conditionType,
		})
		existingCond = &resList[len(resList)-1]
	}
	if existingCond.Status != status {
		existingCond.LastTransitionTime = metav1.NewTime(now)
	}
	existingCond.Status = status
	existingCond.Reason = reason
	existingCond.Message = fmt.Sprintf(message, args...)
	return resList
}
