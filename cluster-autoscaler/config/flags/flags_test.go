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

package flags

import (
	"flag"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseSingleGpuLimit(t *testing.T) {
	type testcase struct {
		input                string
		expectError          bool
		expectedLimits       config.GpuLimits
		expectedErrorMessage string
	}

	testcases := []testcase{
		{
			input:       "gpu:1:10",
			expectError: false,
			expectedLimits: config.GpuLimits{
				GpuType: "gpu",
				Min:     1,
				Max:     10,
			},
		},
		{
			input:                "gpu:1",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit specification: gpu:1",
		},
		{
			input:                "gpu:1:10:x",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit specification: gpu:1:10:x",
		},
		{
			input:                "gpu:x:10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is not integer: gpu:x:10",
		},
		{
			input:                "gpu:1:y",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - max is not integer: gpu:1:y",
		},
		{
			input:                "gpu:-1:10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is less than 0; gpu:-1:10",
		},
		{
			input:                "gpu:1:-10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - max is less than 0; gpu:1:-10",
		},
		{
			input:                "gpu:10:1",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is greater than max; gpu:10:1",
		},
	}

	for _, testcase := range testcases {
		limits, err := parseSingleGpuLimit(testcase.input)
		if testcase.expectError {
			assert.NotNil(t, err)
			if err != nil {
				assert.Equal(t, testcase.expectedErrorMessage, err.Error())
			}
		} else {
			assert.Equal(t, testcase.expectedLimits, limits)
		}
	}
}

func TestParseShutdownGracePeriodsAndPriorities(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []kubelet_config.ShutdownGracePeriodByPodPriority
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "Incorrect string - incorrect priority grace period pairs",
			input: "1:2,34",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing ,",
			input: "1:2, 3:4,",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing space",
			input: "1:2,3:4 ",
			want:  nil,
		},
		{
			name:  "Non integers - 1",
			input: "1:2,3:a",
			want:  nil,
		},
		{
			name:  "Non integers - 2",
			input: "1:2,3:23.2",
			want:  nil,
		},
		{
			name:  "parsable input",
			input: "1:2,3:4",
			want: []kubelet_config.ShutdownGracePeriodByPodPriority{
				{1, 2},
				{3, 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shutdownGracePeriodByPodPriority := parseShutdownGracePeriodsAndPriorities(tc.input)
			assert.Equal(t, tc.want, shutdownGracePeriodByPodPriority)
		})
	}
}

func TestCreateAutoscalingOptions(t *testing.T) {
	for _, tc := range []struct {
		testName            string
		flags               []string
		wantOptionsAsserter func(t *testing.T, gotOptions config.AutoscalingOptions)
	}{
		{
			testName: "DrainPriorityConfig defaults to an empty list when the flag isn't passed",
			flags:    []string{},
			wantOptionsAsserter: func(t *testing.T, gotOptions config.AutoscalingOptions) {
				if diff := cmp.Diff([]kubelet_config.ShutdownGracePeriodByPodPriority{}, gotOptions.DrainPriorityConfig, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("createAutoscalingOptions(): unexpected DrainPriorityConfig field (-want +got): %s", diff)
				}
			},
		},
		{
			testName: "DrainPriorityConfig is parsed correctly when the flag passed",
			flags:    []string{"--drain-priority-config", "5000:60,3000:50,0:40"},
			wantOptionsAsserter: func(t *testing.T, gotOptions config.AutoscalingOptions) {
				wantConfig := []kubelet_config.ShutdownGracePeriodByPodPriority{
					{Priority: 5000, ShutdownGracePeriodSeconds: 60},
					{Priority: 3000, ShutdownGracePeriodSeconds: 50},
					{Priority: 0, ShutdownGracePeriodSeconds: 40},
				}
				if diff := cmp.Diff(wantConfig, gotOptions.DrainPriorityConfig); diff != "" {
					t.Errorf("createAutoscalingOptions(): unexpected DrainPriorityConfig field (-want +got): %s", diff)
				}
			},
		},
		{
			testName: "max startup time is overridden to the highest value if it is smaller than max-inactivity or max-failing-time",
			flags:    []string{"--max-inactivity=10m", "--max-failing-time=15m", "--max-startup-time=5m"},
			wantOptionsAsserter: func(t *testing.T, gotOptions config.AutoscalingOptions) {
				if gotOptions.MaxStartupTime != 15*time.Minute {
					t.Errorf("got max startup time: %v, want %v", gotOptions.MaxStartupTime, 15*time.Minute)
				}
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet("test", pflag.ExitOnError)
			pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
			f := &AutoscalingFlags{}
			f.AddFlags(pflag.CommandLine)
			err := pflag.CommandLine.Parse(tc.flags)
			if err != nil {
				t.Errorf("pflag.CommandLine.Parse() got unexpected error: %v", err)
			}

			gotOptions, err := f.Options()
			if err != nil {
				t.Fatalf("Options() got unexpected error: %v", err)
			}
			tc.wantOptionsAsserter(t, gotOptions)
		})
	}
}

func TestAutoscalingFlagsAllPossible(t *testing.T) {
	tests := map[string]struct {
		Flags    []string
		Validate func(t *testing.T, opts config.AutoscalingOptions)
	}{
		"AllPossibleFlagsSet": {
			Flags: []string{
				"--cluster-name=my-cluster",
				"--address=:8086",
				"--kubernetes=https://apiserver",
				"--kubeconfig=/etc/kubeconfig",
				"--kube-api-content-type=application/json",
				"--kube-client-burst=50",
				"--kube-client-qps=25.5",
				"--cloud-config=/etc/cloud.conf",
				"--cloud-provider=gce",
				"--namespace=custom-ns",
				"--enforce-node-group-min-size=true",
				"--scale-down-unready-enabled=false",
				"--scale-down-delay-after-add=5m",
				"--scale-down-delay-type-local=true",
				"--scale-down-delay-after-delete=1m",
				"--scale-down-delay-after-failure=2m",
				"--scale-down-unneeded-time=3m",
				"--scale-down-unready-time=4m",
				"--scale-down-utilization-threshold=0.4",
				"--scale-down-gpu-utilization-threshold=0.3",
				"--scale-down-non-empty-candidates-count=25",
				"--scale-down-candidates-pool-ratio=0.2",
				"--scale-down-candidates-pool-min-count=100",
				"--node-deletion-delay-timeout=1m",
				"--node-deletion-batcher-interval=10s",
				"--scan-interval=5s",
				"--max-nodes-total=100",
				"--max-bulk-soft-taint-count=5",
				"--max-bulk-soft-taint-time=1s",
				"--max-total-unready-percentage=30",
				"--ok-total-unready-count=5",
				"--scale-up-from-zero=false",
				"--parallel-scale-up=true",
				"--max-node-provision-time=10m",
				"--max-node-startup-time=10m",
				"--max-pod-eviction-time=1m",
				"--estimator=binpacking",
				"--expander=least-waste",
				"--grpc-expander-cert=/path/to/cert",
				"--grpc-expander-url=https://grpc",
				"--ignore-daemonsets-utilization=true",
				"--ignore-mirror-pods-utilization=true",
				"--write-status-configmap=false",
				"--status-config-map-name=custom-status",
				"--max-inactivity=5m",
				"--max-binpacking-time=2m",
				"--max-failing-time=5m",
				"--max-startup-time=10m",
				"--balance-similar-node-groups=true",
				"--unremovable-node-recheck-timeout=2m",
				"--expendable-pods-priority-cutoff=0",
				"--regional=true",
				"--new-pod-scale-up-delay=1m",
				"--scale-from-unschedulable=true",
				"--profiling=true",
				"--clusterapi-cloud-config-authoritative=true",
				"--cordon-node-before-terminating=false",
				"--daemonset-eviction-for-empty-nodes=true",
				"--daemonset-eviction-for-occupied-nodes=false",
				"--user-agent=custom-ca",
				"--emit-per-nodegroup-metrics=true",
				"--debugging-snapshot-enabled=true",
				"--node-info-cache-expire-time=24h",
				"--initial-node-group-backoff-duration=1m",
				"--max-node-group-backoff-duration=10m",
				"--node-group-backoff-reset-timeout=1h",
				"--max-scale-down-parallelism=5",
				"--max-drain-parallelism=2",
				"--record-duplicated-events=true",
				"--max-nodes-per-scaleup=500",
				"--max-nodegroup-binpacking-duration=5s",
				"--fastpath-binpacking-enabled=true",
				"--skip-nodes-with-system-pods=false",
				"--skip-nodes-with-local-storage=false",
				"--skip-nodes-with-custom-controller-pods=false",
				"--min-replica-count=1",
				"--blocking-system-pod-distruption-timeout=30m",
				"--node-delete-delay-after-taint=2s",
				"--scale-down-simulation-timeout=10s",
				"--memory-difference-ratio=0.02",
				"--max-free-difference-ratio=0.03",
				"--max-allocatable-difference-ratio=0.04",
				"--force-ds=true",
				"--dynamic-node-delete-delay-after-taint-enabled=true",
				"--enable-provisioning-requests=true",
				"--provisioning-request-initial-backoff-time=2m",
				"--provisioning-request-max-backoff-time=15m",
				"--provisioning-request-max-backoff-cache-size=500",
				"--frequent-loops-enabled=false",
				"--async-node-groups=true",
				"--enable-proactive-scaleup=true",
				"--salvo-scale-up=true",
				"--salvo-scale-up-budget=30s",
				"--scaleup-simulation-for-skipped-node-groups-enabled=true",
				"--bulk-mig-instances-listing-enabled=true",
				"--gce-concurrent-refreshes=2",
				"--gce-mig-instances-min-refresh-wait-time=2s",
				"--aws-use-static-instance-list=true",
				"--balancing-label=app",
				"--startup-taint-prefix=transient",
				"--pod-injection-limit=1000",
				"--check-capacity-batch-processing=true",
				"--check-capacity-provisioning-request-max-batch-size=5",
				"--check-capacity-provisioning-request-batch-timebox=5s",
				"--force-delete-unregistered-nodes=true",
				"--force-delete-failed-nodes=true",
				"--enable-csi-node-aware-scheduling=true",
				"--predicate-parallelism=8",
				"--check-capacity-processor-instance=inst",
				"--node-deletion-candidate-ttl=1m",
				"--capacity-buffer-controller-enabled=true",
				"--capacity-buffer-pod-injection-enabled=true",
				"--capacity-buffer-pod-dry-run-enabled=false",
				"--node-removal-latency-tracking-enabled=true",
				"--max-node-skip-eval-time-tracker-enabled=true",
				"--capacity-quotas-enabled=true",
				"--cores-total=2:100",
				"--memory-total=4:200",
				"--gpu-total=nvidia:1:10",
				"--max-graceful-termination-sec=300",
			},
			Validate: func(t *testing.T, opts config.AutoscalingOptions) {
				assert.Equal(t, "my-cluster", opts.ClusterName)
				assert.Equal(t, ":8086", opts.Address)
				assert.Equal(t, "https://apiserver", opts.KubeClientOpts.Master)
				assert.Equal(t, "/etc/kubeconfig", opts.KubeClientOpts.KubeConfigPath)
				assert.Equal(t, "application/json", opts.KubeClientOpts.APIContentType)
				assert.Equal(t, 50, opts.KubeClientOpts.KubeClientBurst)
				assert.Equal(t, float32(25.5), opts.KubeClientOpts.KubeClientQPS)
				assert.Equal(t, "/etc/cloud.conf", opts.CloudConfig)
				assert.Equal(t, "gce", opts.CloudProviderName)
				assert.Equal(t, "custom-ns", opts.ConfigNamespace)
				assert.True(t, opts.EnforceNodeGroupMinSize)
				assert.False(t, opts.ScaleDownUnreadyEnabled)
				assert.Equal(t, 5*time.Minute, opts.ScaleDownDelayAfterAdd)
				assert.True(t, opts.ScaleDownDelayTypeLocal)
				assert.Equal(t, 1*time.Minute, opts.ScaleDownDelayAfterDelete)
				assert.Equal(t, 2*time.Minute, opts.ScaleDownDelayAfterFailure)
				assert.Equal(t, 3*time.Minute, opts.NodeGroupDefaults.ScaleDownUnneededTime)
				assert.Equal(t, 4*time.Minute, opts.NodeGroupDefaults.ScaleDownUnreadyTime)
				assert.Equal(t, 0.4, opts.NodeGroupDefaults.ScaleDownUtilizationThreshold)
				assert.Equal(t, 0.3, opts.NodeGroupDefaults.ScaleDownGpuUtilizationThreshold)
				assert.Equal(t, 25, opts.ScaleDownNonEmptyCandidatesCount)
				assert.Equal(t, 0.2, opts.ScaleDownCandidatesPoolRatio)
				assert.Equal(t, 100, opts.ScaleDownCandidatesPoolMinCount)
				assert.Equal(t, 1*time.Minute, opts.NodeDeletionDelayTimeout)
				assert.Equal(t, 10*time.Second, opts.NodeDeletionBatcherInterval)
				assert.Equal(t, 5*time.Second, opts.ScanInterval)
				assert.Equal(t, 100, opts.MaxNodesTotal)
				assert.Equal(t, 5, opts.MaxBulkSoftTaintCount)
				assert.Equal(t, 1*time.Second, opts.MaxBulkSoftTaintTime)
				assert.Equal(t, float64(30), opts.MaxTotalUnreadyPercentage)
				assert.Equal(t, 5, opts.OkTotalUnreadyCount)
				assert.False(t, opts.ScaleUpFromZero)
				assert.True(t, opts.ParallelScaleUp)
				assert.Equal(t, 10*time.Minute, opts.NodeGroupDefaults.MaxNodeProvisionTime)
				assert.Equal(t, 10*time.Minute, opts.NodeGroupDefaults.MaxNodeStartupTime)
				assert.Equal(t, 1*time.Minute, opts.MaxPodEvictionTime)
				assert.Equal(t, "binpacking", opts.EstimatorName)
				assert.Equal(t, "least-waste", opts.ExpanderNames)
				assert.Equal(t, "/path/to/cert", opts.GRPCExpanderCert)
				assert.Equal(t, "https://grpc", opts.GRPCExpanderURL)
				assert.True(t, opts.NodeGroupDefaults.IgnoreDaemonSetsUtilization)
				assert.True(t, opts.IgnoreMirrorPodsUtilization)
				assert.False(t, opts.WriteStatusConfigMap)
				assert.Equal(t, "custom-status", opts.StatusConfigMapName)
				assert.Equal(t, 10*time.Minute, opts.MaxStartupTime) // Overridden by max-startup-time (10m) >= MaxInactivity (5m) and MaxFailing (5m)
				assert.True(t, opts.BalanceSimilarNodeGroups)
				assert.Equal(t, 2*time.Minute, opts.UnremovableNodeRecheckTimeout)
				assert.Equal(t, 0, opts.ExpendablePodsPriorityCutoff)
				assert.True(t, opts.Regional)
				assert.Equal(t, 1*time.Minute, opts.NewPodScaleUpDelay)
				assert.True(t, opts.ScaleFromUnschedulable)
				assert.True(t, opts.EnableProfiling)
				assert.True(t, opts.ClusterAPICloudConfigAuthoritative)
				assert.False(t, opts.CordonNodeBeforeTerminate)
				assert.True(t, opts.DaemonSetEvictionForEmptyNodes)
				assert.False(t, opts.DaemonSetEvictionForOccupiedNodes)
				assert.Equal(t, "custom-ca", opts.UserAgent)
				assert.True(t, opts.EmitPerNodeGroupMetrics)
				assert.True(t, opts.DebuggingSnapshotEnabled)
				assert.Equal(t, 24*time.Hour, opts.NodeInfoCacheExpireTime)
				assert.Equal(t, 1*time.Minute, opts.InitialNodeGroupBackoffDuration)
				assert.Equal(t, 10*time.Minute, opts.MaxNodeGroupBackoffDuration)
				assert.Equal(t, 1*time.Hour, opts.NodeGroupBackoffResetTimeout)
				assert.Equal(t, 5, opts.MaxScaleDownParallelism)
				assert.Equal(t, 2, opts.MaxDrainParallelism)
				assert.True(t, opts.RecordDuplicatedEvents)
				assert.Equal(t, 500, opts.MaxNodesPerScaleUp)
				assert.Equal(t, 5*time.Second, opts.MaxNodeGroupBinpackingDuration)
				assert.True(t, opts.FastpathBinpackingEnabled)
				assert.False(t, opts.SkipNodesWithSystemPods)
				assert.False(t, opts.SkipNodesWithLocalStorage)
				assert.False(t, opts.SkipNodesWithCustomControllerPods)
				assert.Equal(t, 1, opts.MinReplicaCount)
				assert.Equal(t, 30*time.Minute, opts.BspDisruptionTimeout)
				assert.Equal(t, 2*time.Second, opts.NodeDeleteDelayAfterTaint)
				assert.Equal(t, 10*time.Second, opts.ScaleDownSimulationTimeout)
				assert.Equal(t, 0.02, opts.NodeGroupSetRatios.MaxCapacityMemoryDifferenceRatio)
				assert.Equal(t, 0.03, opts.NodeGroupSetRatios.MaxFreeDifferenceRatio)
				assert.Equal(t, 0.04, opts.NodeGroupSetRatios.MaxAllocatableDifferenceRatio)
				assert.True(t, opts.ForceDaemonSets)
				assert.True(t, opts.DynamicNodeDeleteDelayAfterTaintEnabled)
				assert.True(t, opts.ProvisioningRequestEnabled)
				assert.Equal(t, 2*time.Minute, opts.ProvisioningRequestInitialBackoffTime)
				assert.Equal(t, 15*time.Minute, opts.ProvisioningRequestMaxBackoffTime)
				assert.Equal(t, 500, opts.ProvisioningRequestMaxBackoffCacheSize)
				assert.False(t, opts.FrequentLoopsEnabled)
				assert.True(t, opts.AsyncNodeGroupsEnabled)
				assert.True(t, opts.ProactiveScaleupEnabled)
				assert.True(t, opts.SalvoScaleUp)
				assert.Equal(t, 30*time.Second, opts.SalvoScaleUpBudget)
				assert.True(t, opts.ScaleUpSimulationForSkippedNodeGroupsEnabled)
				assert.True(t, opts.GCEOptions.BulkMigInstancesListingEnabled)
				assert.Equal(t, 2, opts.GCEOptions.ConcurrentRefreshes)
				assert.Equal(t, 2*time.Second, opts.GCEOptions.MigInstancesMinRefreshWaitTime)
				assert.True(t, opts.AWSUseStaticInstanceList)
				assert.Equal(t, []string{"app"}, opts.BalancingLabels)
				assert.Equal(t, []string{"transient"}, opts.StartupTaintPrefixes)
				assert.Equal(t, 1000, opts.PodInjectionLimit)
				assert.True(t, opts.CheckCapacityBatchProcessing)
				assert.Equal(t, 5, opts.CheckCapacityProvisioningRequestMaxBatchSize)
				assert.Equal(t, 5*time.Second, opts.CheckCapacityProvisioningRequestBatchTimebox)
				assert.True(t, opts.ForceDeleteLongUnregisteredNodes)
				assert.True(t, opts.ForceDeleteFailedNodes)
				assert.True(t, opts.CSINodeAwareSchedulingEnabled)
				assert.Equal(t, 8, opts.PredicateParallelism)
				assert.Equal(t, "inst", opts.CheckCapacityProcessorInstance)
				assert.Equal(t, 1*time.Minute, opts.NodeDeletionCandidateTTL)
				assert.True(t, opts.CapacitybufferControllerEnabled)
				assert.True(t, opts.CapacitybufferPodInjectionEnabled)
				assert.False(t, opts.CapacityBufferPodDryRunEnabled)
				assert.True(t, opts.NodeRemovalLatencyTrackingEnabled)
				assert.True(t, opts.MaxNodeSkipEvalTimeTrackerEnabled)
				assert.True(t, opts.CapacityQuotasEnabled)
				assert.Equal(t, int64(2), opts.MinCoresTotal)
				assert.Equal(t, int64(100), opts.MaxCoresTotal)
				assert.Equal(t, int64(4*1024*1024*1024), opts.MinMemoryTotal)
				assert.Equal(t, int64(200*1024*1024*1024), opts.MaxMemoryTotal)
				assert.Equal(t, []config.GpuLimits{{"nvidia", 1, 10}}, opts.GpuTotal)
				assert.Equal(t, 300, opts.MaxGracefulTerminationSec)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			f := &AutoscalingFlags{}
			f.AddFlags(fs)
			err := fs.Parse(tc.Flags)
			if err != nil {
				t.Fatalf("unexpected Parse error: %v", err)
			}
			opts, err := f.Options()
			if err != nil {
				t.Fatalf("unexpected Options error: %v", err)
			}
			tc.Validate(t, opts)
		})
	}
}

func TestAutoscalingFlagsValidationEdgeCases(t *testing.T) {
	tests := map[string]struct {
		Flags   []string
		WantErr bool
	}{
		"ValidMinimal": {
			Flags:   []string{},
			WantErr: false,
		},
		"BalancingMutualExclusivity": {
			Flags:   []string{"--balancing-label=app", "--balancing-ignore-label=tier"},
			WantErr: true,
		},
		"DrainPriorityMutualExclusivity": {
			Flags:   []string{"--drain-priority-config=1000:10", "--max-graceful-termination-sec=300"},
			WantErr: true,
		},
		"InvalidPredicateParallelism": {
			Flags:   []string{"--predicate-parallelism=0"},
			WantErr: true,
		},
		"InvalidDRA": {
			Flags:   []string{"--enable-dynamic-resource-allocation=false"},
			WantErr: true,
		},
		"BypassedSchedulerSubset": {
			Flags:   []string{"--allowed-scheduler-names=default-scheduler", "--bypassed-scheduler-names=custom-scheduler"},
			WantErr: true,
		},
		"InvalidCoresTotal": {
			Flags:   []string{"--cores-total=10:5"},
			WantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			f := &AutoscalingFlags{}
			f.AddFlags(fs)
			err := fs.Parse(tc.Flags)
			if err != nil {
				t.Fatalf("unexpected Parse error: %v", err)
			}
			_, err = f.Options()
			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
