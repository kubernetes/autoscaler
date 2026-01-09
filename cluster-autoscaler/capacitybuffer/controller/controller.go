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
	"sort"
	"time"

	"k8s.io/klog/v2"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	translators "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators"
	updater "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"
)

const defaultloopInterval = time.Second * 5
const defaultResyncInterval = time.Minute * 5

// BufferController performs updates on Buffers and convert them to pods to be injected
type BufferController interface {
	// Run to run the reconciliation loop frequently every x seconds
	Run(stopCh <-chan struct{})
}

type bufferController struct {
	client         *cbclient.CapacityBufferClient
	strategyFilter filters.Filter
	translator     translators.Translator
	updater        updater.StatusUpdater
	loopInterval   time.Duration
	resyncInterval time.Duration
	tracker        *DirtyNamespaceTracker
}

// NewBufferController creates new bufferController object
func NewBufferController(
	client *cbclient.CapacityBufferClient,
	strategyFilter filters.Filter,
	translator translators.Translator,
	updater updater.StatusUpdater,
	loopInterval time.Duration,
	resyncInterval time.Duration,
) BufferController {
	return &bufferController{
		client:         client,
		strategyFilter: strategyFilter,
		translator:     translator,
		updater:        updater,
		loopInterval:   loopInterval,
		resyncInterval: resyncInterval,
		tracker:        NewDirtyNamespaceTracker(client),
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
		translator: translators.NewCombinedTranslator(
			[]translators.Translator{
				translators.NewPodTemplateBufferTranslator(client),
				translators.NewDefaultScalableObjectsTranslator(client),
				translators.NewResourceLimitsTranslator(client),
				translators.NewResourceQuotasTranslator(client),
			},
		),
		updater:        *updater.NewStatusUpdater(client),
		loopInterval:   defaultloopInterval,
		resyncInterval: defaultResyncInterval,
		tracker:        NewDirtyNamespaceTracker(client),
	}
}

// Run to run the controller reconcile loop
func (c *bufferController) Run(stopCh <-chan struct{}) {
	c.reconcileAll()

	loopTicker := time.NewTicker(c.loopInterval)
	defer loopTicker.Stop()
	resyncTicker := time.NewTicker(c.resyncInterval)
	defer resyncTicker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-resyncTicker.C:
			c.reconcileAll()
		case <-loopTicker.C:
			c.reconcile()
		}
	}
}

// reconcileAll performs a full resync of all namespaces
func (c *bufferController) reconcileAll() {
	klog.V(2).Infof("Capacity buffer controller starting full resync")
	allBuffers, err := c.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("Capacity buffer controller failed to list all buffers for full resync: %v", err)
		return
	}

	namespaces := make(map[string]bool)
	for _, buffer := range allBuffers {
		namespaces[buffer.Namespace] = true
	}

	klog.V(2).Infof("Capacity buffer controller full resync processing %d namespaces", len(namespaces))
	for namespace := range namespaces {
		c.reconcileNamespace(namespace)
	}
}

// Reconcile represents single iteration in the main-loop of Updater
func (c *bufferController) reconcile() {
	namespacesToProcess := make(map[string]bool)

	// 1. Add Dirty Namespaces (Event-Driven)
	dirtyNamespaces := c.tracker.GetAndClearDirtyNamespaces()
	for _, ns := range dirtyNamespaces {
		namespacesToProcess[ns] = true
	}

	// 2. Add Retry Namespaces (Polling for unstable buffers)
	retryNamespaces := c.getNamespacesNeedingRetry()
	for _, ns := range retryNamespaces {
		namespacesToProcess[ns] = true
	}

	if len(namespacesToProcess) == 0 {
		return
	}

	klog.V(2).Infof("Capacity buffer controller processing %d namespaces", len(namespacesToProcess))

	for namespace := range namespacesToProcess {
		c.reconcileNamespace(namespace)
	}
}

func (c *bufferController) getNamespacesNeedingRetry() []string {
	// List all buffers (from cache) to find those that are not in a stable state
	// We use empty string namespace to list all
	allBuffers, err := c.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("Capacity buffer controller failed to list all buffers for retry check: %v", err)
		return nil
	}

	namespaces := make(map[string]bool)
	for _, buffer := range allBuffers {
		if !isBufferStable(buffer) {
			namespaces[buffer.Namespace] = true
		}
	}

	result := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		result = append(result, ns)
	}
	return result
}

func isBufferStable(buffer *v1.CapacityBuffer) bool {
	ready := false
	provisioning := false

	for _, cond := range buffer.Status.Conditions {
		if cond.Type == common.ReadyForProvisioningCondition && cond.Status == common.ConditionTrue {
			ready = true
		}
		if cond.Type == common.ProvisioningCondition && cond.Status == common.ConditionTrue {
			provisioning = true
		}
	}
	return ready || provisioning
}

func (c *bufferController) reconcileNamespace(namespace string) {
	klog.V(2).Infof("Reconciling namespace: %s", namespace)
	// List all capacity buffers in the target namespace
	buffers, err := c.client.ListCapacityBuffers(namespace)
	if err != nil {
		klog.Errorf("Capacity buffer controller failed to list buffers in namespace %s with error: %v", namespace, err.Error())
		return
	}

	// Filter the desired provisioning strategy
	// Note: We process ALL buffers in the namespace that match the strategy.
	filteredBuffers, _ := c.strategyFilter.Filter(buffers)

	if len(filteredBuffers) == 0 {
		return
	}

	// Sort buffers deterministically by CreationTimestamp, then Name
	sort.Slice(filteredBuffers, func(i, j int) bool {
		if filteredBuffers[i].CreationTimestamp.Time.Equal(filteredBuffers[j].CreationTimestamp.Time) {
			return filteredBuffers[i].Name < filteredBuffers[j].Name
		}
		return filteredBuffers[i].CreationTimestamp.Before(&filteredBuffers[j].CreationTimestamp)
	})

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
