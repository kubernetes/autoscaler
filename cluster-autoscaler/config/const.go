/*
Copyright 2017 The Kubernetes Authors.

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

package config

import "time"

const (
	// SchedulerConfigFileFlag is the name of the flag
	// for passing in custom scheduler config for in-tree scheduelr plugins
	SchedulerConfigFileFlag = "scheduler-config-file"

	// DefaultMaxClusterCores is the default maximum number of cores in the cluster.
	DefaultMaxClusterCores = 5000 * 64
	// DefaultMaxClusterMemory is the default maximum number of gigabytes of memory in cluster.
	DefaultMaxClusterMemory = 5000 * 64 * 20

	// DefaultScaleDownUtilizationThresholdKey identifies ScaleDownUtilizationThreshold autoscaling option
	DefaultScaleDownUtilizationThresholdKey = "scaledownutilizationthreshold"
	// DefaultScaleDownGpuUtilizationThresholdKey identifies ScaleDownGpuUtilizationThreshold autoscaling option
	DefaultScaleDownGpuUtilizationThresholdKey = "scaledowngpuutilizationthreshold"
	// DefaultScaleDownUnneededTimeKey identifies ScaleDownUnneededTime autoscaling option
	DefaultScaleDownUnneededTimeKey = "scaledownunneededtime"
	// DefaultScaleDownUnreadyTimeKey identifies ScaleDownUnreadyTime autoscaling option
	DefaultScaleDownUnreadyTimeKey = "scaledownunreadytime"
	// DefaultMaxNodeProvisionTimeKey identifies MaxNodeProvisionTime autoscaling option
	DefaultMaxNodeProvisionTimeKey = "maxnodeprovisiontime"
	// DefaultIgnoreDaemonSetsUtilizationKey identifies IgnoreDaemonSetsUtilization autoscaling option
	DefaultIgnoreDaemonSetsUtilizationKey = "ignoredaemonsetsutilization"

	// DefaultScaleDownUnneededTime is the default time duration for which CA waits before deleting an unneeded node
	DefaultScaleDownUnneededTime = 10 * time.Minute
	// DefaultScaleDownUnreadyTime identifies ScaleDownUnreadyTime autoscaling option
	DefaultScaleDownUnreadyTime = 20 * time.Minute
	// DefaultScaleDownUtilizationThreshold identifies ScaleDownUtilizationThreshold autoscaling option
	DefaultScaleDownUtilizationThreshold = 0.5
	// DefaultScaleDownGpuUtilizationThreshold identifies ScaleDownGpuUtilizationThreshold autoscaling option
	DefaultScaleDownGpuUtilizationThreshold = 0.5
	// DefaultScaleDownDelayAfterFailure is the default value for ScaleDownDelayAfterFailure autoscaling option
	DefaultScaleDownDelayAfterFailure = 3 * time.Minute
	// DefaultScanInterval is the default scan interval for CA
	DefaultScanInterval = 10 * time.Second
)
