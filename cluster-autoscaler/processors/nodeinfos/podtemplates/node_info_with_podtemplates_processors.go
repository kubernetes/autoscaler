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

package podtemplates

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfos"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	scheduler_utils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// PodTemplateDaemonSetLabelKey use as label key on PodTemplate corresponding to an extra Daemonset.
	PodTemplateDaemonSetLabelKey = "cluster-autoscaler.kubernetes.io/daemonset-pod"
	// PodTemplateDaemonSetLabelValueTrue use as PodTemplateDaemonSetLabelKey label value.
	PodTemplateDaemonSetLabelValueTrue = "true"
)

// NewNodeInfoWithPodTemplateProcessor returns a default instance of NodeInfoProcessor.
func NewNodeInfoWithPodTemplateProcessor(opts *core.AutoscalerOptions) nodeinfos.NodeInfoProcessor {
	internalContext, cancelFunc := context.WithCancel(context.Background())

	return &nodeInfoWithPodTemplateProcessor{
		ctx:               internalContext,
		cancelFunc:        cancelFunc,
		podTemplateLister: newPodTemplateLister(opts.KubeClient, internalContext.Done()),
	}
}

// nodeInfoWithPodTemplateProcessor add possible PodTemplates in nodeInfos.
type nodeInfoWithPodTemplateProcessor struct {
	podTemplateLister v1lister.PodTemplateLister

	ctx        context.Context
	cancelFunc func()
}

const templateNodeFromTemplatePrefix = core_utils.TemplateNodeForNamePrefix + "-" + core_utils.TemplateNodeForNameFromTemplatePrefix

// Process returns unchanged nodeInfos.
func (p *nodeInfoWithPodTemplateProcessor) Process(ctx *ca_context.AutoscalingContext, nodeInfosForNodeGroups map[string]*schedulerframework.NodeInfo) (map[string]*schedulerframework.NodeInfo, error) {
	// here we can use empty snapshot, since the NodeInfos that will be updated
	// are from CloudProvider NodeTemplates.
	clusterSnapshot := simulator.NewBasicClusterSnapshot()

	// retrieve only once the podTemplates list.
	// This list will be used for each NodeGroup.
	podTemplates, err := p.podTemplateLister.List(labels.Everything())
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}

	result := make(map[string]*schedulerframework.NodeInfo, len(nodeInfosForNodeGroups))
	var errs []error
	for id, nodeInfo := range nodeInfosForNodeGroups {
		var newNodeInfo *schedulerframework.NodeInfo

		// only runs getNodeInfoWithPodTemplates() for NodeInfo created from
		// cloudprovider.TemplateNodeInfo.
		// If not a NodeTemplate, Pods from PodTemplates should already be present
		// in the PodList attached to the Node.
		if strings.HasPrefix(nodeInfo.Node().Name, templateNodeFromTemplatePrefix) {
			var err error
			newNodeInfo, err = getNodeInfoWithPodTemplates(nodeInfo, podTemplates, clusterSnapshot, ctx.PredicateChecker)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			newNodeInfo = nodeInfosForNodeGroups[id]
		}

		result[id] = newNodeInfo
	}

	return result, utilerrors.NewAggregate(errs)
}

// CleanUp cleans up processor's internal structuxres.
func (p *nodeInfoWithPodTemplateProcessor) CleanUp() {
	p.cancelFunc()
}

func newPodTemplateLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.PodTemplateLister {
	podTemplateWatchOption := func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Everything().String()
		options.LabelSelector = labels.SelectorFromSet(getDaemonsetPodTemplateLabelSet()).String()
	}
	listWatcher := cache.NewFilteredListWatchFromClient(kubeClient.CoreV1().RESTClient(), "podtemplates", apiv1.NamespaceAll, podTemplateWatchOption)
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.PodTemplate{}, time.Hour)
	lister := v1lister.NewPodTemplateLister(store)
	go reflector.Run(stopchannel)
	return lister
}

const nodeInfoDeepCopySuffix = "podtemplate"

func getNodeInfoWithPodTemplates(baseNodeInfo *schedulerframework.NodeInfo, podTemplates []*apiv1.PodTemplate, clusterSnapshot *simulator.BasicClusterSnapshot, predicateChecker simulator.PredicateChecker) (*schedulerframework.NodeInfo, error) {
	// clone baseNodeInfo to not modify the input object.
	newNodeInfo := scheduler_utils.DeepCopyTemplateNode(baseNodeInfo, nodeInfoDeepCopySuffix)
	node := newNodeInfo.Node()
	var pods []*apiv1.Pod

	for _, podInfo := range baseNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}

	if err := clusterSnapshot.AddNodeWithPods(node, pods); err != nil {
		return nil, err
	}

	for _, podTpl := range podTemplates {
		newPod := newPod(podTpl, node.Name)
		err := predicateChecker.CheckPredicates(clusterSnapshot, newPod, node.Name)
		if err == nil {
			newNodeInfo.AddPod(newPod)
		} else if err.ErrorType() == simulator.NotSchedulablePredicateError {
			// ok; we are just skipping this daemonset
		} else {
			// unexpected error
			return nil, fmt.Errorf("unexpected error while calling PredicateChecker; %v", err)
		}
	}

	return newNodeInfo, nil
}

func getDaemonsetPodTemplateLabelSet() labels.Set {
	daemonsetPodTemplateLabels := map[string]string{
		PodTemplateDaemonSetLabelKey: PodTemplateDaemonSetLabelValueTrue,
	}
	return labels.Set(daemonsetPodTemplateLabels)
}

func newPod(pt *apiv1.PodTemplate, nodeName string) *apiv1.Pod {
	newPod := &apiv1.Pod{Spec: pt.Template.Spec, ObjectMeta: pt.Template.ObjectMeta}
	newPod.Namespace = pt.Namespace
	newPod.Name = fmt.Sprintf("%s-pod", pt.Name)
	newPod.Spec.NodeName = nodeName
	return newPod
}
