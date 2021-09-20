/*
Copyright 2021 The Kubernetes Authors.

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

package podtemplate

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	v1lister "k8s.io/client-go/listers/core/v1"
)

// NewPodTemplateProcessor returns a default instance of NodeInfoProcessor.
func NewPodTemplateProcessor(opts *core.AutoscalerOptions) Interface {
	if opts == nil || !opts.NodeInfosProcessorPodTemplates {
		return &dummyPodTemplateProcessor{}
	}

	internalContext, cancelFunc := context.WithCancel(context.Background())
	return &podTemplateProcessor{
		ctx:               internalContext,
		cancelFunc:        cancelFunc,
		podTemplateLister: newPodTemplateLister(opts.KubeClient, internalContext.Done()),
	}
}

// Interface define the PodTemplateProcess interface to allow having a several implementation depending of the core.AutoscalerOptions.
type Interface interface {
	GetDaemonSetPodsFromPodTemplateForNode(baseNodeInfo *schedulerframework.NodeInfo, predicateChecker predicatechecker.PredicateChecker, ignoredTaints taints.TaintKeySet) ([]*apiv1.Pod, error)
	CleanUp()
}

// podTemplateProcessor add possible PodTemplates in nodeInfos.
type podTemplateProcessor struct {
	podTemplateLister v1lister.PodTemplateLister

	ctx        context.Context
	cancelFunc func()
}

// GetDaemonSetPodsFromPodTemplateForNode return a list of apiv1.Pod from PodTemplate present in the cluster.
func (p *podTemplateProcessor) GetDaemonSetPodsFromPodTemplateForNode(baseNodeInfo *schedulerframework.NodeInfo, predicateChecker predicatechecker.PredicateChecker, ignoredTaints taints.TaintKeySet) ([]*apiv1.Pod, error) {
	// retrieve only once the podTemplates list.
	// This list will be used for each NodeGroup.
	podTemplates, err := p.podTemplateLister.List(labels.Everything())
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	if len(podTemplates) == 0 {
		return nil, nil
	}

	var result []*apiv1.Pod

	// here we can use empty snapshot
	clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()

	// add a node with pods - node info is created by cloud provider,
	// we don't know whether it'll have pods or not.
	var pods []*apiv1.Pod
	for _, podInfo := range baseNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	if err := clusterSnapshot.AddNodeWithPods(baseNodeInfo.Node(), pods); err != nil {
		return nil, err
	}

	var errs []error
	for _, podTpl := range podTemplates {
		pod := newPod(podTpl, baseNodeInfo.Node().Name)
		err := predicateChecker.CheckPredicates(clusterSnapshot, pod, baseNodeInfo.Node().Name)
		if err == nil {
			result = append(result, pod)
		} else if err.ErrorType() == predicatechecker.NotSchedulablePredicateError {
			// ok; we are just skipping this daemonset
		} else {
			// unexpected error
			errs = append(errs, fmt.Errorf("unexpected error while calling PredicateChecker; %v", err))
		}
	}
	klog.V(6).Infof("the podTemplateProcessor has created %d from PodTemplate(s) for the Node %s", len(result), baseNodeInfo.Node().GetName())

	return result, utilerrors.NewAggregate(errs)
}

func newPod(pt *apiv1.PodTemplate, nodeName string) *apiv1.Pod {
	newPod := &apiv1.Pod{Spec: pt.Template.Spec, ObjectMeta: pt.Template.ObjectMeta}
	newPod.Namespace = pt.Namespace
	newPod.Name = fmt.Sprintf("%s-pod", pt.Name)
	newPod.Spec.NodeName = nodeName
	return newPod
}

// CleanUp cleans up processor's internal structuxres.
func (p *podTemplateProcessor) CleanUp() {
	p.cancelFunc()
}

// dummyPodTemplateProcessor dummy implementation when the feature is not enabled
type dummyPodTemplateProcessor struct{}

func (p *dummyPodTemplateProcessor) GetDaemonSetPodsFromPodTemplateForNode(baseNodeInfo *schedulerframework.NodeInfo, predicateChecker predicatechecker.PredicateChecker, ignoredTaints taints.TaintKeySet) ([]*apiv1.Pod, error) {
	return nil, nil
}

func (p *dummyPodTemplateProcessor) CleanUp() {
}
