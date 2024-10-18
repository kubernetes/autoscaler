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

package test

const (
	// Custom scheduler configs for testing

	// SchedulerConfigNodeResourcesFitDisabled is scheduler config
	// with `NodeResourcesFit` plugin disabled
	SchedulerConfigNodeResourcesFitDisabled = `
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
profiles:
- pluginConfig:
  plugins:
    multiPoint:
      disabled:
      - name: NodeResourcesFit
        weight: 1
  schedulerName: custom-scheduler`

	// SchedulerConfigTaintTolerationDisabled is scheduler config
	// with `TaintToleration` plugin disabled
	SchedulerConfigTaintTolerationDisabled = `
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
profiles:
- pluginConfig:
  plugins:
    multiPoint:
      disabled:
      - name: TaintToleration
        weight: 1
  schedulerName: custom-scheduler`

	// SchedulerConfigMultiProfiles is scheduler config
	// with multiple profiles,
	// default-scheduler all plugin enabled, custom-scheduler `NodeResourcesFit` plugin disabled
	SchedulerConfigMultiProfiles = `
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
profiles:
- schedulerName: default-scheduler
- pluginConfig:
  plugins:
    multiPoint:
      disabled:
      - name: NodeResourcesFit
        weight: 1
  schedulerName: custom-scheduler`

	// SchedulerConfigMinimalCorrect is the minimal
	// correct scheduler config
	SchedulerConfigMinimalCorrect = `
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration`

	// SchedulerConfigDecodeErr is the scheduler config
	// which throws decoding error when we try to load it
	SchedulerConfigDecodeErr = `
kind: KubeSchedulerConfiguration`

	// SchedulerConfigInvalid is invalid scheduler config
	// because we specify percentageOfNodesToScore > 100
	SchedulerConfigInvalid = `
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
# percentageOfNodesToScore has to be between 0 and 100
percentageOfNodesToScore: 130`
)
