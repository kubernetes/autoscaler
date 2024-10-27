/*
Copyright 2018 The Kubernetes Authors.

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

package status

import (
	"fmt"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
)

// ScaleUpStatus is the status of a scale-up attempt. This includes information
// on if scale-up happened, description of scale-up operation performed and
// status of pods that took part in the scale-up evaluation.
type ScaleUpStatus struct {
	Result                   ScaleUpResult
	ScaleUpError             *errors.AutoscalerError
	ScaleUpInfos             []nodegroupset.ScaleUpInfo
	PodsTriggeredScaleUp     []*apiv1.Pod
	PodsRemainUnschedulable  []NoScaleUpInfo
	PodsAwaitEvaluation      []*apiv1.Pod
	CreateNodeGroupResults   []nodegroups.CreateNodeGroupResult
	ConsideredNodeGroups     []cloudprovider.NodeGroup
	FailedCreationNodeGroups []cloudprovider.NodeGroup
	FailedResizeNodeGroups   []cloudprovider.NodeGroup
}

// NoScaleUpInfo contains information about a pod that didn't trigger scale-up.
type NoScaleUpInfo struct {
	Pod                *apiv1.Pod
	RejectedNodeGroups map[string]Reasons
	SkippedNodeGroups  map[string]Reasons
}

// ScaleUpResult represents the result of a scale up.
type ScaleUpResult int

const (
	// ScaleUpSuccessful - a scale-up successfully occurred.
	ScaleUpSuccessful ScaleUpResult = iota
	// ScaleUpError - an unexpected error occurred during the scale-up attempt.
	ScaleUpError
	// ScaleUpNoOptionsAvailable - there were no node groups that could be considered for the scale-up.
	ScaleUpNoOptionsAvailable
	// ScaleUpNotNeeded - there was no need for a scale-up e.g. because there were no unschedulable pods.
	ScaleUpNotNeeded
	// ScaleUpNotTried - the scale up wasn't even attempted, e.g. an autoscaling iteration was skipped, or
	// an error occurred before the scale up logic.
	ScaleUpNotTried
	// ScaleUpInCooldown - the scale up wasn't even attempted, because it's in a cooldown state (it's suspended for a scheduled period of time).
	ScaleUpInCooldown
)

// WasSuccessful returns true if the scale-up was successful.
func (s *ScaleUpStatus) WasSuccessful() bool {
	return s.Result == ScaleUpSuccessful
}

// Reasons interface provides a list of reasons for why something happened or didn't happen.
type Reasons interface {
	Reasons() []string
}

// ScaleUpStatusProcessor processes the status of the cluster after a scale-up.
type ScaleUpStatusProcessor interface {
	Process(context *context.AutoscalingContext, status *ScaleUpStatus)
	CleanUp()
}

// NewDefaultScaleUpStatusProcessor creates a default instance of ScaleUpStatusProcessor.
func NewDefaultScaleUpStatusProcessor() ScaleUpStatusProcessor {
	return &EventingScaleUpStatusProcessor{}
}

// NoOpScaleUpStatusProcessor is a ScaleUpStatusProcessor implementations useful for testing.
type NoOpScaleUpStatusProcessor struct{}

// Process processes the status of the cluster after a scale-up.
func (p *NoOpScaleUpStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleUpStatus) {
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpScaleUpStatusProcessor) CleanUp() {
}

// CombinedScaleUpStatusProcessor is a list of ScaleUpStatusProcessor
type CombinedScaleUpStatusProcessor struct {
	processors []ScaleUpStatusProcessor
}

// NewCombinedScaleUpStatusProcessor construct CombinedScaleUpStatusProcessor.
func NewCombinedScaleUpStatusProcessor(processors []ScaleUpStatusProcessor) *CombinedScaleUpStatusProcessor {
	var scaleUpProcessors []ScaleUpStatusProcessor
	for _, processor := range processors {
		if processor != nil {
			scaleUpProcessors = append(scaleUpProcessors, processor)
		}
	}
	return &CombinedScaleUpStatusProcessor{scaleUpProcessors}
}

// AddProcessor append processor to the list.
func (p *CombinedScaleUpStatusProcessor) AddProcessor(processor ScaleUpStatusProcessor) {
	if processor != nil {
		p.processors = append(p.processors, processor)
	}
}

// Process runs sub-processors sequentially in the same order of addition
func (p *CombinedScaleUpStatusProcessor) Process(ctx *context.AutoscalingContext, status *ScaleUpStatus) {
	for _, processor := range p.processors {
		processor.Process(ctx, status)
	}
}

// CleanUp cleans up the processor's internal structures.
func (p *CombinedScaleUpStatusProcessor) CleanUp() {
	for _, processor := range p.processors {
		processor.CleanUp()
	}
}

// UpdateScaleUpError updates ScaleUpStatus.
func UpdateScaleUpError(s *ScaleUpStatus, err errors.AutoscalerError) (*ScaleUpStatus, errors.AutoscalerError) {
	s.ScaleUpError = &err
	s.Result = ScaleUpError
	return s, err
}

// combinedStatusSet is a helper struct to combine multiple ScaleUpStatuses into one. It keeps track of the best result and all errors that occurred during the ScaleUp process.
type combinedStatusSet struct {
	Result                      ScaleUpResult
	ScaleupErrors               map[*errors.AutoscalerError]bool
	ScaleUpInfosSet             map[nodegroupset.ScaleUpInfo]bool
	PodsTriggeredScaleUpSet     map[*apiv1.Pod]bool
	PodsRemainUnschedulableSet  map[*NoScaleUpInfo]bool
	PodsAwaitEvaluationSet      map[*apiv1.Pod]bool
	CreateNodeGroupResultsSet   map[*nodegroups.CreateNodeGroupResult]bool
	ConsideredNodeGroupsSet     map[cloudprovider.NodeGroup]bool
	FailedCreationNodeGroupsSet map[cloudprovider.NodeGroup]bool
	FailedResizeNodeGroupsSet   map[cloudprovider.NodeGroup]bool
}

// Add adds a ScaleUpStatus to the combinedStatusSet.
func (c *combinedStatusSet) Add(status *ScaleUpStatus) {
	resultPriority := map[ScaleUpResult]int{
		ScaleUpNotTried:           0,
		ScaleUpNotNeeded:          1,
		ScaleUpNoOptionsAvailable: 2,
		ScaleUpError:              3,
		ScaleUpSuccessful:         4,
	}

	// If even one scaleUpSuccessful is present, the final result is ScaleUpSuccessful.
	// If no ScaleUpSuccessful is present, and even one ScaleUpError is present, the final result is ScaleUpError.
	// If no ScaleUpSuccessful or ScaleUpError is present, and even one ScaleUpNoOptionsAvailable is present, the final result is ScaleUpNoOptionsAvailable.
	// If no ScaleUpSuccessful, ScaleUpError or ScaleUpNoOptionsAvailable is present, the final result is ScaleUpNotTried.
	if resultPriority[c.Result] < resultPriority[status.Result] {
		c.Result = status.Result
	}
	if status.ScaleUpError != nil {
		if _, found := c.ScaleupErrors[status.ScaleUpError]; !found {
			c.ScaleupErrors[status.ScaleUpError] = true
		}
	}
	if status.ScaleUpInfos != nil {
		for _, scaleUpInfo := range status.ScaleUpInfos {
			if _, found := c.ScaleUpInfosSet[scaleUpInfo]; !found {
				c.ScaleUpInfosSet[scaleUpInfo] = true
			}
		}
	}
	if status.PodsTriggeredScaleUp != nil {
		for _, pod := range status.PodsTriggeredScaleUp {
			if _, found := c.PodsTriggeredScaleUpSet[pod]; !found {
				c.PodsTriggeredScaleUpSet[pod] = true
			}
		}
	}
	if status.PodsRemainUnschedulable != nil {
		for _, pod := range status.PodsRemainUnschedulable {
			if _, found := c.PodsRemainUnschedulableSet[&pod]; !found {
				c.PodsRemainUnschedulableSet[&pod] = true
			}
		}
	}
	if status.PodsAwaitEvaluation != nil {
		for _, pod := range status.PodsAwaitEvaluation {
			if _, found := c.PodsAwaitEvaluationSet[pod]; !found {
				c.PodsAwaitEvaluationSet[pod] = true
			}
		}
	}
	if status.CreateNodeGroupResults != nil {
		for _, createNodeGroupResult := range status.CreateNodeGroupResults {
			if _, found := c.CreateNodeGroupResultsSet[&createNodeGroupResult]; !found {
				c.CreateNodeGroupResultsSet[&createNodeGroupResult] = true
			}
		}
	}
	if status.ConsideredNodeGroups != nil {
		for _, nodeGroup := range status.ConsideredNodeGroups {
			if _, found := c.ConsideredNodeGroupsSet[nodeGroup]; !found {
				c.ConsideredNodeGroupsSet[nodeGroup] = true
			}
		}
	}
	if status.FailedCreationNodeGroups != nil {
		for _, nodeGroup := range status.FailedCreationNodeGroups {
			if _, found := c.FailedCreationNodeGroupsSet[nodeGroup]; !found {
				c.FailedCreationNodeGroupsSet[nodeGroup] = true
			}
		}
	}
	if status.FailedResizeNodeGroups != nil {
		for _, nodeGroup := range status.FailedResizeNodeGroups {
			if _, found := c.FailedResizeNodeGroupsSet[nodeGroup]; !found {
				c.FailedResizeNodeGroupsSet[nodeGroup] = true
			}
		}
	}
}

// formatMessageFromBatchErrors formats a message from a list of errors.
func (c *combinedStatusSet) formatMessageFromBatchErrors(errs []errors.AutoscalerError, printErrorTypes bool) string {
	firstErr := errs[0]
	var builder strings.Builder
	builder.WriteString(firstErr.Error())
	builder.WriteString(" ...and other concurrent errors: [")
	formattedErrs := map[errors.AutoscalerError]bool{
		firstErr: true,
	}
	for _, err := range errs {
		if _, has := formattedErrs[err]; has {
			continue
		}
		formattedErrs[err] = true
		var message string
		if printErrorTypes {
			message = fmt.Sprintf("[%s] %s", err.Type(), err.Error())
		} else {
			message = err.Error()
		}
		if len(formattedErrs) > 2 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", message))
	}
	builder.WriteString("]")
	return builder.String()
}

// combineBatchScaleUpErrors combines multiple errors into one. If there is only one error, it returns that error. If there are multiple errors, it combines them into one error with a message that contains all the errors.
func (c *combinedStatusSet) combineBatchScaleUpErrors() *errors.AutoscalerError {
	if len(c.ScaleupErrors) == 0 {
		return nil
	}
	if len(c.ScaleupErrors) == 1 {
		for err := range c.ScaleupErrors {
			return err
		}
	}
	uniqueMessages := make(map[string]bool)
	uniqueTypes := make(map[errors.AutoscalerErrorType]bool)
	for err := range c.ScaleupErrors {
		uniqueTypes[(*err).Type()] = true
		uniqueMessages[(*err).Error()] = true
	}
	if len(uniqueTypes) == 1 && len(uniqueMessages) == 1 {
		for err := range c.ScaleupErrors {
			return err
		}
	}
	// sort to stabilize the results and easier log aggregation
	errs := make([]errors.AutoscalerError, 0, len(c.ScaleupErrors))
	for err := range c.ScaleupErrors {
		errs = append(errs, *err)
	}
	sort.Slice(errs, func(i, j int) bool {
		errA := errs[i]
		errB := errs[j]
		if errA.Type() == errB.Type() {
			return errs[i].Error() < errs[j].Error()
		}
		return errA.Type() < errB.Type()
	})
	firstErr := errs[0]
	printErrorTypes := len(uniqueTypes) > 1
	message := c.formatMessageFromBatchErrors(errs, printErrorTypes)
	combinedErr := errors.NewAutoscalerError(firstErr.Type(), message)
	return &combinedErr
}

// Export converts the combinedStatusSet into a ScaleUpStatus.
func (c *combinedStatusSet) Export() (*ScaleUpStatus, errors.AutoscalerError) {
	result := &ScaleUpStatus{Result: c.Result}
	if len(c.ScaleupErrors) > 0 {
		result.ScaleUpError = c.combineBatchScaleUpErrors()
	}
	if len(c.ScaleUpInfosSet) > 0 {
		for scaleUpInfo := range c.ScaleUpInfosSet {
			result.ScaleUpInfos = append(result.ScaleUpInfos, scaleUpInfo)
		}
	}
	if len(c.PodsTriggeredScaleUpSet) > 0 {
		for pod := range c.PodsTriggeredScaleUpSet {
			result.PodsTriggeredScaleUp = append(result.PodsTriggeredScaleUp, pod)
		}
	}
	if len(c.PodsRemainUnschedulableSet) > 0 {
		for pod := range c.PodsRemainUnschedulableSet {
			result.PodsRemainUnschedulable = append(result.PodsRemainUnschedulable, *pod)
		}
	}
	if len(c.PodsAwaitEvaluationSet) > 0 {
		for pod := range c.PodsAwaitEvaluationSet {
			result.PodsAwaitEvaluation = append(result.PodsAwaitEvaluation, pod)
		}
	}
	if len(c.CreateNodeGroupResultsSet) > 0 {
		for createNodeGroupResult := range c.CreateNodeGroupResultsSet {
			result.CreateNodeGroupResults = append(result.CreateNodeGroupResults, *createNodeGroupResult)
		}
	}
	if len(c.ConsideredNodeGroupsSet) > 0 {
		for nodeGroup := range c.ConsideredNodeGroupsSet {
			result.ConsideredNodeGroups = append(result.ConsideredNodeGroups, nodeGroup)
		}
	}
	if len(c.FailedCreationNodeGroupsSet) > 0 {
		for nodeGroup := range c.FailedCreationNodeGroupsSet {
			result.FailedCreationNodeGroups = append(result.FailedCreationNodeGroups, nodeGroup)
		}
	}
	if len(c.FailedResizeNodeGroupsSet) > 0 {
		for nodeGroup := range c.FailedResizeNodeGroupsSet {
			result.FailedResizeNodeGroups = append(result.FailedResizeNodeGroups, nodeGroup)
		}
	}

	var resErr errors.AutoscalerError

	if result.Result == ScaleUpError {
		resErr = *result.ScaleUpError
	}

	return result, resErr
}

// NewCombinedStatusSet creates a new combinedStatusSet.
func NewCombinedStatusSet() combinedStatusSet {
	return combinedStatusSet{
		Result:                      ScaleUpNotTried,
		ScaleupErrors:               make(map[*errors.AutoscalerError]bool),
		ScaleUpInfosSet:             make(map[nodegroupset.ScaleUpInfo]bool),
		PodsTriggeredScaleUpSet:     make(map[*apiv1.Pod]bool),
		PodsRemainUnschedulableSet:  make(map[*NoScaleUpInfo]bool),
		PodsAwaitEvaluationSet:      make(map[*apiv1.Pod]bool),
		CreateNodeGroupResultsSet:   make(map[*nodegroups.CreateNodeGroupResult]bool),
		ConsideredNodeGroupsSet:     make(map[cloudprovider.NodeGroup]bool),
		FailedCreationNodeGroupsSet: make(map[cloudprovider.NodeGroup]bool),
		FailedResizeNodeGroupsSet:   make(map[cloudprovider.NodeGroup]bool),
	}
}
