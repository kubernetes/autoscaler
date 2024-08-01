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

package orchestrator

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/nodegroupchange"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

// ScaleUpExecutor scales up node groups.
type scaleUpExecutor struct {
	autoscalingContext         *context.AutoscalingContext
	scaleStateNotifier         nodegroupchange.NodeGroupChangeObserver
	asyncNodeGroupStateChecker asyncnodegroups.AsyncNodeGroupStateChecker
}

// New returns new instance of scale up executor.
func newScaleUpExecutor(
	autoscalingContext *context.AutoscalingContext,
	scaleStateNotifier nodegroupchange.NodeGroupChangeObserver,
	asyncNodeGroupStateChecker asyncnodegroups.AsyncNodeGroupStateChecker,
) *scaleUpExecutor {
	return &scaleUpExecutor{
		autoscalingContext:         autoscalingContext,
		scaleStateNotifier:         scaleStateNotifier,
		asyncNodeGroupStateChecker: asyncNodeGroupStateChecker,
	}
}

// ExecuteScaleUps executes the scale ups, based on the provided scale up infos and options.
// May scale up groups concurrently when autoscler option is enabled.
// In case of issues returns an error and a scale up info which failed to execute.
// If there were multiple concurrent errors one combined error is returned.
func (e *scaleUpExecutor) ExecuteScaleUps(
	scaleUpInfos []nodegroupset.ScaleUpInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	now time.Time,
	atomic bool,
) (errors.AutoscalerError, []cloudprovider.NodeGroup) {
	options := e.autoscalingContext.AutoscalingOptions
	if options.ParallelScaleUp {
		return e.executeScaleUpsParallel(scaleUpInfos, nodeInfos, now, atomic)
	}
	return e.executeScaleUpsSync(scaleUpInfos, nodeInfos, now, atomic)
}

func (e *scaleUpExecutor) executeScaleUpsSync(
	scaleUpInfos []nodegroupset.ScaleUpInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	now time.Time,
	atomic bool,
) (errors.AutoscalerError, []cloudprovider.NodeGroup) {
	availableGPUTypes := e.autoscalingContext.CloudProvider.GetAvailableGPUTypes()
	for _, scaleUpInfo := range scaleUpInfos {
		nodeInfo, ok := nodeInfos[scaleUpInfo.Group.Id()]
		if !ok {
			klog.Errorf("ExecuteScaleUp: failed to get node info for node group %s", scaleUpInfo.Group.Id())
			continue
		}
		if aErr := e.executeScaleUp(scaleUpInfo, nodeInfo, availableGPUTypes, now, atomic); aErr != nil {
			return aErr, []cloudprovider.NodeGroup{scaleUpInfo.Group}
		}
	}
	return nil, nil
}

func (e *scaleUpExecutor) executeScaleUpsParallel(
	scaleUpInfos []nodegroupset.ScaleUpInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	now time.Time,
	atomic bool,
) (errors.AutoscalerError, []cloudprovider.NodeGroup) {
	if err := checkUniqueNodeGroups(scaleUpInfos); err != nil {
		return err, extractNodeGroups(scaleUpInfos)
	}
	type errResult struct {
		err  errors.AutoscalerError
		info *nodegroupset.ScaleUpInfo
	}
	scaleUpsLen := len(scaleUpInfos)
	errResults := make(chan errResult, scaleUpsLen)
	var wg sync.WaitGroup
	wg.Add(scaleUpsLen)
	availableGPUTypes := e.autoscalingContext.CloudProvider.GetAvailableGPUTypes()
	for _, scaleUpInfo := range scaleUpInfos {
		go func(info nodegroupset.ScaleUpInfo) {
			defer wg.Done()
			nodeInfo, ok := nodeInfos[info.Group.Id()]
			if !ok {
				klog.Errorf("ExecuteScaleUp: failed to get node info for node group %s", info.Group.Id())
				return
			}
			if aErr := e.executeScaleUp(info, nodeInfo, availableGPUTypes, now, atomic); aErr != nil {
				errResults <- errResult{err: aErr, info: &info}
			}
		}(scaleUpInfo)
	}
	wg.Wait()
	close(errResults)
	var results []errResult
	for err := range errResults {
		results = append(results, err)
	}
	if len(results) > 0 {
		failedNodeGroups := make([]cloudprovider.NodeGroup, len(results))
		scaleUpErrors := make([]errors.AutoscalerError, len(results))
		for i, result := range results {
			failedNodeGroups[i] = result.info.Group
			scaleUpErrors[i] = result.err
		}
		return combineConcurrentScaleUpErrors(scaleUpErrors), failedNodeGroups
	}
	return nil, nil
}

func (e *scaleUpExecutor) increaseSize(nodeGroup cloudprovider.NodeGroup, increase int, atomic bool) error {
	if atomic {
		if err := nodeGroup.AtomicIncreaseSize(increase); err != cloudprovider.ErrNotImplemented {
			return err
		}
		// If error is cloudprovider.ErrNotImplemented, fall back to non-atomic
		// increase - cloud provider doesn't support it.
	}
	return nodeGroup.IncreaseSize(increase)
}

func (e *scaleUpExecutor) executeScaleUp(
	info nodegroupset.ScaleUpInfo,
	nodeInfo *schedulerframework.NodeInfo,
	availableGPUTypes map[string]struct{},
	now time.Time,
	atomic bool,
) errors.AutoscalerError {
	gpuConfig := e.autoscalingContext.CloudProvider.GetNodeGpuConfig(nodeInfo.Node())
	gpuResourceName, gpuType := gpu.GetGpuInfoForMetrics(gpuConfig, availableGPUTypes, nodeInfo.Node(), nil)
	klog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	e.autoscalingContext.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: setting group %s size to %d instead of %d (max: %d)", info.Group.Id(), info.NewSize, info.CurrentSize, info.MaxSize)
	increase := info.NewSize - info.CurrentSize
	if err := e.increaseSize(info.Group, increase, atomic); err != nil {
		e.autoscalingContext.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToScaleUpGroup", "Scale-up failed for group %s: %v", info.Group.Id(), err)
		aerr := errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to increase node group size: ")
		e.scaleStateNotifier.RegisterFailedScaleUp(info.Group, string(aerr.Type()), aerr.Error(), gpuResourceName, gpuType, now)
		return aerr
	}
	if increase < 0 {
		return errors.NewAutoscalerError(errors.InternalError, fmt.Sprintf("increase in number of nodes cannot be negative, got: %v", increase))
	}
	if e.asyncNodeGroupStateChecker.IsUpcoming(info.Group) {
		// Don't emit scale up event for upcoming node group as it will be generated after
		// the node group is created, during initial scale up.
		klog.V(0).Infof("Scale-up: group %s is an upcoming node group, skipping emit scale up event", info.Group.Id())
		return nil
	}
	e.scaleStateNotifier.RegisterScaleUp(info.Group, increase, time.Now())
	metrics.RegisterScaleUp(increase, gpuResourceName, gpuType)
	e.autoscalingContext.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d instead of %d (max: %d)", info.Group.Id(), info.NewSize, info.CurrentSize, info.MaxSize)
	return nil
}

func combineConcurrentScaleUpErrors(errs []errors.AutoscalerError) errors.AutoscalerError {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	uniqueMessages := make(map[string]bool)
	uniqueTypes := make(map[errors.AutoscalerErrorType]bool)
	for _, err := range errs {
		uniqueTypes[err.Type()] = true
		uniqueMessages[err.Error()] = true
	}
	if len(uniqueTypes) == 1 && len(uniqueMessages) == 1 {
		return errs[0]
	}
	// sort to stabilize the results and easier log aggregation
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
	message := formatMessageFromConcurrentErrors(errs, printErrorTypes)
	return errors.NewAutoscalerError(firstErr.Type(), message)
}

func formatMessageFromConcurrentErrors(errs []errors.AutoscalerError, printErrorTypes bool) string {
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

// Checks if all groups are scaled only once.
// Scaling one group multiple times concurrently may cause problems.
func checkUniqueNodeGroups(scaleUpInfos []nodegroupset.ScaleUpInfo) errors.AutoscalerError {
	uniqueGroups := make(map[string]bool)
	for _, info := range scaleUpInfos {
		if uniqueGroups[info.Group.Id()] {
			return errors.NewAutoscalerError(
				errors.InternalError,
				"assertion failure: detected group double scaling: %s", info.Group.Id(),
			)
		}
		uniqueGroups[info.Group.Id()] = true
	}
	return nil
}

func extractNodeGroups(scaleUpInfos []nodegroupset.ScaleUpInfo) []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, len(scaleUpInfos))
	for i, info := range scaleUpInfos {
		groups[i] = info.Group
	}
	return groups
}
