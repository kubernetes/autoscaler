/*
Copyright The Kubernetes Authors.

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

package metrics

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	pausingInterval = time.Minute
	loggingQuota    = 5
)

type capacityBufferMetrics interface {
	ObserveCapacityBuffersProcessingIntervalSeconds(isNewBuffer bool, duration time.Duration)
}

type capacityBufferClient interface {
	ListCapacityBuffers(namespace string) ([]*v1beta1.CapacityBuffer, error)
}

// processingIntervalMetricReporter is a struct that contains a capacitybuffer client.
type processingIntervalMetricReporter struct {
	client                 capacityBufferClient
	processedBuffers       *ProcessingCache
	supportedBuffersFilter filters.Filter
	metrics                capacityBufferMetrics
	clock                  clock.Clock
}

// NewProcessingIntervalMetricReporter creates a new instance of processingIntervalMetricReporter
func NewProcessingIntervalMetricReporter(client *cbclient.CapacityBufferClient, supportedBuffersFilter filters.Filter, processedBuffers *ProcessingCache, clock clock.Clock) *processingIntervalMetricReporter {
	return &processingIntervalMetricReporter{
		client:                 client,
		processedBuffers:       processedBuffers,
		supportedBuffersFilter: supportedBuffersFilter,
		metrics:                metrics.DefaultMetrics,
		clock:                  clock,
	}
}

// Run is the goroutine that runs every running interval (1 min) to calculate and publish the processing interval metric
func (r *processingIntervalMetricReporter) Run(stopCh <-chan struct{}) {
	klog.V(1).Infof("Running capacity buffers processing time metric reporter")
	wait.NonSlidingUntil(r.loop, pausingInterval, stopCh)
	klog.Info("Stopping capacity buffers processing time metric reporter")
}

func (r *processingIntervalMetricReporter) loop() {
	var maximumNewBuffersProcessingTime *time.Duration
	var maximumProcessedBuffersProcessingTime *time.Duration

	// List all capacity buffers
	buffers, err := r.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("Failed to list capacity buffers with error: %v", err.Error())
		return
	}
	supportedBuffers, _ := r.supportedBuffersFilter.Filter(buffers)
	processedBuffersSnapshot := r.processedBuffers.Snapshot()

	supportedUIDs := make([]string, 0, len(supportedBuffers))
	for _, buffer := range supportedBuffers {
		supportedUIDs = append(supportedUIDs, string(buffer.UID))
	}
	// Delete unsupported buffers resulting from buffers deleting or change in provisioning strategy
	r.processedBuffers.Prune(supportedUIDs)

	loggingQuota := klogx.NewLoggingQuota(loggingQuota)
	for _, buffer := range supportedBuffers {
		bufferProcessingDuration, isNewBuffer, skipped := r.calculateBufferProcessingDuration(buffer, processedBuffersSnapshot)
		if skipped {
			klogx.V(2).UpTo(loggingQuota).Infof("Skipping capacity buffer %s/%s from capacity_buffer_processing_interval_seconds metric: "+
				"buffer is not registered in processed buffers cache while having conditions", buffer.Namespace, buffer.Name)
			continue
		}
		if isNewBuffer {
			if maximumNewBuffersProcessingTime == nil || bufferProcessingDuration > *maximumNewBuffersProcessingTime {
				d := bufferProcessingDuration
				maximumNewBuffersProcessingTime = &d
			}
		} else {
			if maximumProcessedBuffersProcessingTime == nil || bufferProcessingDuration > *maximumProcessedBuffersProcessingTime {
				d := bufferProcessingDuration
				maximumProcessedBuffersProcessingTime = &d
			}
		}
	}
	klogx.V(2).Over(loggingQuota).Infof("Skipped %d other capacity buffers from capacity_buffer_processing_interval_seconds metric "+
		"for not being registered in processed buffers cache while having conditions", -loggingQuota.Left())

	newDuration := time.Duration(0)
	if maximumNewBuffersProcessingTime != nil {
		newDuration = *maximumNewBuffersProcessingTime
	}
	processedDuration := time.Duration(0)
	if maximumProcessedBuffersProcessingTime != nil {
		processedDuration = *maximumProcessedBuffersProcessingTime
	}
	r.reportProcessingIntervalMetric(newDuration, processedDuration)
}

// calculateBufferProcessingDuration calculates and returns the duration since buffer last processed
// or since creation if buffer was not processed with a flag indicating if the buffer is new and another
// flag indicates buffer is skipped
func (r *processingIntervalMetricReporter) calculateBufferProcessingDuration(buffer *v1beta1.CapacityBuffer, processedBuffersSnapshot map[string]time.Time) (time.Duration, bool, bool) {
	isNewBuffer := false
	lastProcessingTime, exists := processedBuffersSnapshot[string(buffer.UID)]
	hasConditions := bufferHasConditions(buffer)

	// non-cached buffers with no conditions are considered new buffers
	if !exists && !hasConditions {
		lastProcessingTime = buffer.CreationTimestamp.Time
		isNewBuffer = true

		// non-cached buffers with conditions are skipped (CA restart)
	} else if !exists && hasConditions {
		return 0, false, true
	}
	processingTime := r.clock.Since(lastProcessingTime)
	return processingTime, isNewBuffer, false
}

func bufferHasConditions(buffer *v1beta1.CapacityBuffer) bool {
	return len(buffer.Status.Conditions) != 0
}

func (r *processingIntervalMetricReporter) reportProcessingIntervalMetric(maximumNewBuffersProcessingTime time.Duration, maximumProcessedBuffersProcessingTime time.Duration) {
	r.metrics.ObserveCapacityBuffersProcessingIntervalSeconds(true, maximumNewBuffersProcessingTime)
	r.metrics.ObserveCapacityBuffersProcessingIntervalSeconds(false, maximumProcessedBuffersProcessingTime)
}
