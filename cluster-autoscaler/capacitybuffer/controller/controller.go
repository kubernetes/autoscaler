/*
Copyright 2025 The Kubernetes Authors.

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
	"time"

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/labels"
	buffersclient "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/listers/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	common "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	translators "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators"
	updater "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"

	client "k8s.io/client-go/kubernetes"
)

const loopInterval = time.Second * 5

// BufferController performs updates on Buffers and convert them to pods to be injected
type BufferController interface {
	// Run to run the reconciliation loop frequently every x seconds
	Run(stopCh <-chan struct{})
}

type bufferController struct {
	buffersLister  v1.CapacityBufferLister
	strategyFilter filters.Filter
	statusFilter   filters.Filter
	translator     translators.Translator
	updater        updater.StatusUpdater
	loopInterval   time.Duration
}

// NewBufferController creates new bufferController object
func NewBufferController(
	buffersLister v1.CapacityBufferLister,
	strategyFilter filters.Filter,
	statusFilter filters.Filter,
	translator translators.Translator,
	updater updater.StatusUpdater,
	loopInterval time.Duration,
) BufferController {
	return &bufferController{
		buffersLister:  buffersLister,
		strategyFilter: strategyFilter,
		statusFilter:   statusFilter,
		translator:     translator,
		updater:        updater,
		loopInterval:   loopInterval,
	}
}

// NewDefaultBufferController creates bufferController with default configs
func NewDefaultBufferController(
	listerRegistry kubernetes.ListerRegistry,
	capacityBufferClinet buffersclient.Clientset,
	nodeBufferListener v1.CapacityBufferLister,
	kubeClient client.Clientset,
) BufferController {
	return &bufferController{
		buffersLister: nodeBufferListener,
		// Accepting empty string as it represents nil value for ProvisioningStrategy
		strategyFilter: filters.NewStrategyFilter([]string{common.ActiveProvisioningStrategy, ""}),
		statusFilter: filters.NewStatusFilter(map[string]string{
			common.ReadyForProvisioningCondition: common.ConditionTrue,
			common.ProvisioningCondition:         common.ConditionTrue,
		}),
		translator: translators.NewCombinedTranslator(
			[]translators.Translator{
				translators.NewPodTemplateBufferTranslator(),
			},
		),
		updater:      *updater.NewStatusUpdater(&capacityBufferClinet),
		loopInterval: loopInterval,
	}
}

// Run to run the controller reconcile loop
func (c *bufferController) Run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case <-time.After(c.loopInterval):
			c.reconcile()
		}
	}
}

// Reconcile represents single iteration in the main-loop of Updater
func (c *bufferController) reconcile() {

	// List all capacity buffers objects
	buffers, err := c.buffersLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Capacity buffer controller failed to list buffers with error: %v", err.Error())
		return
	}
	klog.V(2).Infof("Capacity buffer controller listed [%v] buffers", len(buffers))

	// Filter the desired provisioning strategy
	filteredBuffers, _ := c.strategyFilter.Filter(buffers)
	klog.V(2).Infof("Capacity buffer controller filtered %v buffers with buffers strategy filter", len(filteredBuffers))

	// Filter the desired status
	toBeTranslatedBuffers, _ := c.statusFilter.Filter(filteredBuffers)
	klog.V(2).Infof("Capacity buffer controller filtered %v buffers with buffers status filter", len(filteredBuffers))

	// Extract pod specs and number of replicas from filtered buffers
	errors := c.translator.Translate(toBeTranslatedBuffers)
	logErrors(errors)

	// Update buffer status by calling API server
	errors = c.updater.Update(toBeTranslatedBuffers)
	logErrors(errors)
}

func logErrors(errors []error) {
	for _, error := range errors {
		klog.Errorf("Capacity buffer controller error: %v", error.Error())
	}
}
