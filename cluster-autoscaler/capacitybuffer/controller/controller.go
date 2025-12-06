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

	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	filter "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	translators "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators"
	updater "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"
)

const defaultloopInterval = time.Second * 5
const defaultIterationsToReprocessAll = 60

// BufferController performs updates on Buffers and convert them to pods to be injected
type BufferController interface {
	// Run to run the reconciliation loop frequently every x seconds
	Run(stopCh <-chan struct{})
}

type bufferController struct {
	client                   *cbclient.CapacityBufferClient
	strategyFilter           filters.Filter
	statusFilter             filters.Filter
	translator               translators.Translator
	updater                  updater.StatusUpdater
	loopInterval             time.Duration
	iterationsToReprocessAll int
	currentIteration         int
}

// NewBufferController creates new bufferController object
func NewBufferController(
	client *cbclient.CapacityBufferClient,
	strategyFilter filters.Filter,
	statusFilter filters.Filter,
	translator translators.Translator,
	updater updater.StatusUpdater,
	loopInterval time.Duration,
	iterationsToReprocessAll int,
) BufferController {
	return &bufferController{
		client:                   client,
		strategyFilter:           strategyFilter,
		statusFilter:             statusFilter,
		translator:               translator,
		updater:                  updater,
		loopInterval:             loopInterval,
		iterationsToReprocessAll: iterationsToReprocessAll,
	}
}

// NewDefaultBufferController creates bufferController with default configs
func NewDefaultBufferController(
	client *cbclient.CapacityBufferClient,
) BufferController {
	return &bufferController{
		client: client,
		// Accepting empty string as it represents nil value for ProvisioningStrategy
		strategyFilter: filters.NewStrategyFilter([]string{common.ActiveProvisioningStrategy, ""}),
		statusFilter: filter.NewCombinedAnyFilter(
			[]filters.Filter{
				filters.NewStatusFilter(map[string]string{
					common.ReadyForProvisioningCondition: common.ConditionTrue,
					common.ProvisioningCondition:         common.ConditionTrue,
				}),
				filters.NewBufferGenerationChangedFilter(),
				filters.NewPodTemplateGenerationChangedFilter(client),
			},
		),
		translator: translators.NewCombinedTranslator(
			[]translators.Translator{
				translators.NewPodTemplateBufferTranslator(client),
				translators.NewDefaultScalableObjectsTranslator(client),
				translators.NewResourceLimitsTranslator(client),
			},
		),
		updater:                  *updater.NewStatusUpdater(client),
		loopInterval:             defaultloopInterval,
		iterationsToReprocessAll: defaultIterationsToReprocessAll,
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
	c.currentIteration += 1

	// List all capacity buffers objects
	buffers, err := c.client.ListCapacityBuffers()
	if err != nil {
		klog.Errorf("Capacity buffer controller failed to list buffers with error: %v", err.Error())
		return
	}
	klog.V(2).Infof("Capacity buffer controller listed [%v] buffers", len(buffers))

	// Filter the desired provisioning strategy
	filteredBuffers, _ := c.strategyFilter.Filter(buffers)
	klog.V(2).Infof("Capacity buffer controller filtered %v buffers with buffers strategy filter", len(filteredBuffers))

	// Filter the desired status
	if c.currentIteration < c.iterationsToReprocessAll {
		filteredBuffers, _ = c.statusFilter.Filter(filteredBuffers)
		klog.V(2).Infof("Capacity buffer controller filtered %v buffers with buffers status filter", len(filteredBuffers))
	} else {
		c.currentIteration = 0
		klog.V(2).Infof("Capacity buffer controller skipped buffers status filter, translating all %v buffers", len(filteredBuffers))
	}

	// Extract pod specs and number of replicas from filtered buffers
	errors := c.translator.Translate(filteredBuffers)
	logErrors(errors)

	// Update buffer status by calling API server
	errors = c.updater.Update(filteredBuffers)
	logErrors(errors)
}

func logErrors(errors []error) {
	for _, error := range errors {
		klog.Errorf("Capacity buffer controller error: %v", error.Error())
	}
}
