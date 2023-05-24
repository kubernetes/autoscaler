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

package main

import (
	ctx "context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/emptycandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/previouscandidates"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	kube_flag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/component-base/config/options"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
)

func createAutoscalingOptions() config.AutoscalingOptions {
	minCoresTotal, maxCoresTotal, err := parseMinMaxFlag(*coresTotal)
	if err != nil {
		klog.Fatalf("Failed to parse flags: %v", err)
	}
	minMemoryTotal, maxMemoryTotal, err := parseMinMaxFlag(*memoryTotal)
	if err != nil {
		klog.Fatalf("Failed to parse flags: %v", err)
	}
	// Convert memory limits to bytes.
	minMemoryTotal = minMemoryTotal * units.GiB
	maxMemoryTotal = maxMemoryTotal * units.GiB

	parsedGpuTotal, err := parseMultipleGpuLimits(*gpuTotal)
	if err != nil {
		klog.Fatalf("Failed to parse flags: %v", err)
	}
	if *maxDrainParallelismFlag > 1 && !*parallelDrain {
		klog.Fatalf("Invalid configuration, could not use --max-drain-parallelism > 1 if --parallel-drain is false")
	}
	return config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    *scaleDownUtilizationThreshold,
			ScaleDownGpuUtilizationThreshold: *scaleDownGpuUtilizationThreshold,
			ScaleDownUnneededTime:            *scaleDownUnneededTime,
			ScaleDownUnreadyTime:             *scaleDownUnreadyTime,
			MaxNodeProvisionTime:             *maxNodeProvisionTime,
		},
		CloudConfig:                      *cloudConfig,
		CloudProviderName:                *cloudProviderFlag,
		NodeGroupAutoDiscovery:           *nodeGroupAutoDiscoveryFlag,
		MaxTotalUnreadyPercentage:        *maxTotalUnreadyPercentage,
		OkTotalUnreadyCount:              *okTotalUnreadyCount,
		ScaleUpFromZero:                  *scaleUpFromZero,
		ParallelScaleUp:                  *parallelScaleUp,
		EstimatorName:                    *estimatorFlag,
		ExpanderNames:                    *expanderFlag,
		GRPCExpanderCert:                 *grpcExpanderCert,
		GRPCExpanderURL:                  *grpcExpanderURL,
		IgnoreDaemonSetsUtilization:      *ignoreDaemonSetsUtilization,
		IgnoreMirrorPodsUtilization:      *ignoreMirrorPodsUtilization,
		MaxBulkSoftTaintCount:            *maxBulkSoftTaintCount,
		MaxBulkSoftTaintTime:             *maxBulkSoftTaintTime,
		MaxEmptyBulkDelete:               *maxEmptyBulkDeleteFlag,
		MaxGracefulTerminationSec:        *maxGracefulTerminationFlag,
		MaxPodEvictionTime:               *maxPodEvictionTime,
		MaxNodesTotal:                    *maxNodesTotal,
		MaxCoresTotal:                    maxCoresTotal,
		MinCoresTotal:                    minCoresTotal,
		MaxMemoryTotal:                   maxMemoryTotal,
		MinMemoryTotal:                   minMemoryTotal,
		GpuTotal:                         parsedGpuTotal,
		NodeGroups:                       *nodeGroupsFlag,
		EnforceNodeGroupMinSize:          *enforceNodeGroupMinSize,
		ScaleDownDelayAfterAdd:           *scaleDownDelayAfterAdd,
		ScaleDownDelayAfterDelete:        *scaleDownDelayAfterDelete,
		ScaleDownDelayAfterFailure:       *scaleDownDelayAfterFailure,
		ScaleDownEnabled:                 *scaleDownEnabled,
		ScaleDownUnreadyEnabled:          *scaleDownUnreadyEnabled,
		ScaleDownNonEmptyCandidatesCount: *scaleDownNonEmptyCandidatesCount,
		ScaleDownCandidatesPoolRatio:     *scaleDownCandidatesPoolRatio,
		ScaleDownCandidatesPoolMinCount:  *scaleDownCandidatesPoolMinCount,
		WriteStatusConfigMap:             *writeStatusConfigMapFlag,
		StatusConfigMapName:              *statusConfigMapName,
		BalanceSimilarNodeGroups:         *balanceSimilarNodeGroupsFlag,
		ConfigNamespace:                  *namespace,
		ClusterName:                      *clusterName,
		NodeAutoprovisioningEnabled:      *nodeAutoprovisioningEnabled,
		MaxAutoprovisionedNodeGroupCount: *maxAutoprovisionedNodeGroupCount,
		UnremovableNodeRecheckTimeout:    *unremovableNodeRecheckTimeout,
		ExpendablePodsPriorityCutoff:     *expendablePodsPriorityCutoff,
		Regional:                         *regional,
		NewPodScaleUpDelay:               *newPodScaleUpDelay,
		IgnoredTaints:                    *ignoreTaintsFlag,
		BalancingExtraIgnoredLabels:      *balancingIgnoreLabelsFlag,
		BalancingLabels:                  *balancingLabelsFlag,
		KubeConfigPath:                   *kubeConfigFile,
		KubeClientBurst:                  *kubeClientBurst,
		KubeClientQPS:                    *kubeClientQPS,
		NodeDeletionDelayTimeout:         *nodeDeletionDelayTimeout,
		AWSUseStaticInstanceList:         *awsUseStaticInstanceList,
		GCEOptions: config.GCEOptions{
			ConcurrentRefreshes:             *concurrentGceRefreshes,
			MigInstancesMinRefreshWaitTime:  *gceMigInstancesMinRefreshWaitTime,
			ExpanderEphemeralStorageSupport: *gceExpanderEphemeralStorageSupport,
		},
		ClusterAPICloudConfigAuthoritative: *clusterAPICloudConfigAuthoritative,
		CordonNodeBeforeTerminate:          *cordonNodeBeforeTerminate,
		DaemonSetEvictionForEmptyNodes:     *daemonSetEvictionForEmptyNodes,
		DaemonSetEvictionForOccupiedNodes:  *daemonSetEvictionForOccupiedNodes,
		UserAgent:                          *userAgent,
		InitialNodeGroupBackoffDuration:    *initialNodeGroupBackoffDuration,
		MaxNodeGroupBackoffDuration:        *maxNodeGroupBackoffDuration,
		NodeGroupBackoffResetTimeout:       *nodeGroupBackoffResetTimeout,
		MaxScaleDownParallelism:            *maxScaleDownParallelismFlag,
		MaxDrainParallelism:                *maxDrainParallelismFlag,
		RecordDuplicatedEvents:             *recordDuplicatedEvents,
		MaxNodesPerScaleUp:                 *maxNodesPerScaleUp,
		MaxNodeGroupBinpackingDuration:     *maxNodeGroupBinpackingDuration,
		NodeDeletionBatcherInterval:        *nodeDeletionBatcherInterval,
		SkipNodesWithSystemPods:            *skipNodesWithSystemPods,
		SkipNodesWithLocalStorage:          *skipNodesWithLocalStorage,
		MinReplicaCount:                    *minReplicaCount,
		NodeDeleteDelayAfterTaint:          *nodeDeleteDelayAfterTaint,
		ScaleDownSimulationTimeout:         *scaleDownSimulationTimeout,
		ParallelDrain:                      *parallelDrain,
		SkipNodesWithCustomControllerPods:  *skipNodesWithCustomControllerPods,
		NodeGroupSetRatios: config.NodeGroupDifferenceRatios{
			MaxCapacityMemoryDifferenceRatio: *maxCapacityMemoryDifferenceRatio,
			MaxAllocatableDifferenceRatio:    *maxAllocatableDifferenceRatio,
			MaxFreeDifferenceRatio:           *maxFreeDifferenceRatio,
		},
	}
}

func getKubeConfig() *rest.Config {
	if *kubeConfigFile != "" {
		klog.V(1).Infof("Using kubeconfig file: %s", *kubeConfigFile)
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigFile)
		if err != nil {
			klog.Fatalf("Failed to build config: %v", err)
		}
		return config
	}
	url, err := url.Parse(*kubernetes)
	if err != nil {
		klog.Fatalf("Failed to parse Kubernetes url: %v", err)
	}

	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		klog.Fatalf("Failed to build Kubernetes client configuration: %v", err)
	}

	return kubeConfig
}

func createKubeClient(kubeConfig *rest.Config) kube_client.Interface {
	return kube_client.NewForConfigOrDie(kubeConfig)
}

func registerSignalHandlers(autoscaler core.Autoscaler) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	klog.V(1).Info("Registered cleanup signal handler")

	go func() {
		<-sigs
		klog.V(1).Info("Received signal, attempting cleanup")
		autoscaler.ExitCleanUp()
		klog.V(1).Info("Cleaned up, exiting...")
		klog.Flush()
		os.Exit(0)
	}()
}

func buildAutoscaler(debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) (core.Autoscaler, error) {
	// Create basic config from flags.
	autoscalingOptions := createAutoscalingOptions()

	kubeClientConfig := getKubeConfig()
	kubeClientConfig.Burst = autoscalingOptions.KubeClientBurst
	kubeClientConfig.QPS = float32(autoscalingOptions.KubeClientQPS)
	kubeClient := createKubeClient(kubeClientConfig)

	eventsKubeClient := createKubeClient(getKubeConfig())

	predicateChecker, err := predicatechecker.NewSchedulerBasedPredicateChecker(kubeClient, make(chan struct{}))
	if err != nil {
		return nil, err
	}

	opts := core.AutoscalerOptions{
		AutoscalingOptions:   autoscalingOptions,
		ClusterSnapshot:      clustersnapshot.NewDeltaClusterSnapshot(),
		KubeClient:           kubeClient,
		EventsKubeClient:     eventsKubeClient,
		DebuggingSnapshotter: debuggingSnapshotter,
		PredicateChecker:     predicateChecker,
	}

	opts.Processors = ca_processors.DefaultProcessors()
	opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nodeInfoCacheExpireTime, *forceDaemonSets)
	opts.Processors.PodListProcessor = podlistprocessor.NewDefaultPodListProcessor(opts.PredicateChecker)
	scaleDownCandidatesComparers := []scaledowncandidates.CandidatesComparer{}
	if autoscalingOptions.ParallelDrain {
		sdCandidatesSorting := previouscandidates.NewPreviousCandidates()
		scaleDownCandidatesComparers = []scaledowncandidates.CandidatesComparer{
			emptycandidates.NewEmptySortingProcessor(&autoscalingOptions, emptycandidates.NewNodeInfoGetter(opts.ClusterSnapshot)),
			sdCandidatesSorting,
		}
		opts.Processors.ScaleDownCandidatesNotifier.Register(sdCandidatesSorting)
	}
	sdProcessor := scaledowncandidates.NewScaleDownCandidatesSortingProcessor(scaleDownCandidatesComparers)
	opts.Processors.ScaleDownNodeProcessor = sdProcessor

	var nodeInfoComparator nodegroupset.NodeInfoComparator
	if len(autoscalingOptions.BalancingLabels) > 0 {
		nodeInfoComparator = nodegroupset.CreateLabelNodeInfoComparator(autoscalingOptions.BalancingLabels)
	} else {
		nodeInfoComparatorBuilder := nodegroupset.CreateGenericNodeInfoComparator
		if autoscalingOptions.CloudProviderName == cloudprovider.AzureProviderName {
			nodeInfoComparatorBuilder = nodegroupset.CreateAzureNodeInfoComparator
		} else if autoscalingOptions.CloudProviderName == cloudprovider.AwsProviderName {
			nodeInfoComparatorBuilder = nodegroupset.CreateAwsNodeInfoComparator
			opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewAsgTagResourceNodeInfoProvider(nodeInfoCacheExpireTime, *forceDaemonSets)
		} else if autoscalingOptions.CloudProviderName == cloudprovider.GceProviderName {
			nodeInfoComparatorBuilder = nodegroupset.CreateGceNodeInfoComparator
			opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewAnnotationNodeInfoProvider(nodeInfoCacheExpireTime, *forceDaemonSets)
		}
		nodeInfoComparator = nodeInfoComparatorBuilder(autoscalingOptions.BalancingExtraIgnoredLabels, autoscalingOptions.NodeGroupSetRatios)
	}

	opts.Processors.NodeGroupSetProcessor = &nodegroupset.BalancingNodeGroupSetProcessor{
		Comparator: nodeInfoComparator,
	}

	// These metrics should be published only once.
	metrics.UpdateNapEnabled(autoscalingOptions.NodeAutoprovisioningEnabled)
	metrics.UpdateCPULimitsCores(autoscalingOptions.MinCoresTotal, autoscalingOptions.MaxCoresTotal)
	metrics.UpdateMemoryLimitsBytes(autoscalingOptions.MinMemoryTotal, autoscalingOptions.MaxMemoryTotal)

	// Create autoscaler.
	return core.NewAutoscaler(opts)
}

func run(healthCheck *metrics.HealthCheck, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) {
	metrics.RegisterAll(*emitPerNodeGroupMetrics)

	autoscaler, err := buildAutoscaler(debuggingSnapshotter)
	if err != nil {
		klog.Fatalf("Failed to create autoscaler: %v", err)
	}

	// Register signal handlers for graceful shutdown.
	registerSignalHandlers(autoscaler)

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	// Start components running in background.
	if err := autoscaler.Start(); err != nil {
		klog.Fatalf("Failed to autoscaler background components: %v", err)
	}

	// Autoscale ad infinitum.
	for {
		select {
		case <-time.After(*scanInterval):
			{
				loopStart := time.Now()
				metrics.UpdateLastTime(metrics.Main, loopStart)
				healthCheck.UpdateLastActivity(loopStart)

				err := autoscaler.RunOnce(loopStart)
				if err != nil && err.Type() != errors.TransientError {
					metrics.RegisterError(err)
				} else {
					healthCheck.UpdateLastSuccessfulRun(time.Now())
				}

				metrics.UpdateDurationFromStart(metrics.Main, loopStart)
			}
		}
	}
}

func main() {
	klog.InitFlags(nil)

	leaderElection := defaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = true

	options.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)
	utilfeature.DefaultMutableFeatureGate.AddFlag(pflag.CommandLine)
	kube_flag.InitFlags()

	healthCheck := metrics.NewHealthCheck(*maxInactivityTimeFlag, *maxFailingTimeFlag)

	klog.V(1).Infof("Cluster Autoscaler %s", version.ClusterAutoscalerVersion)

	debuggingSnapshotter := debuggingsnapshot.NewDebuggingSnapshotter(*debuggingSnapshotEnabled)

	go func() {
		pathRecorderMux := mux.NewPathRecorderMux("cluster-autoscaler")
		defaultMetricsHandler := legacyregistry.Handler().ServeHTTP
		pathRecorderMux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
			defaultMetricsHandler(w, req)
		})
		if *debuggingSnapshotEnabled {
			pathRecorderMux.HandleFunc("/snapshotz", debuggingSnapshotter.ResponseHandler)
		}
		pathRecorderMux.HandleFunc("/health-check", healthCheck.ServeHTTP)
		if *enableProfiling {
			routes.Profiling{}.Install(pathRecorderMux)
		}
		err := http.ListenAndServe(*address, pathRecorderMux)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()

	if !leaderElection.LeaderElect {
		run(healthCheck, debuggingSnapshotter)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Unable to get hostname: %v", err)
		}

		kubeClient := createKubeClient(getKubeConfig())

		// Validate that the client is ok.
		_, err = kubeClient.CoreV1().Nodes().List(ctx.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("Failed to get nodes from apiserver: %v", err)
		}

		lock, err := resourcelock.New(
			leaderElection.ResourceLock,
			*namespace,
			leaderElection.ResourceName,
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity:      id,
				EventRecorder: kube_util.CreateEventRecorder(kubeClient, *recordDuplicatedEvents),
			},
		)
		if err != nil {
			klog.Fatalf("Unable to create leader election lock: %v", err)
		}

		leaderelection.RunOrDie(ctx.TODO(), leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ ctx.Context) {
					// Since we are committing a suicide after losing
					// mastership, we can safely ignore the argument.
					run(healthCheck, debuggingSnapshotter)
				},
				OnStoppedLeading: func() {
					klog.Fatalf("lost master")
				},
			},
		})
	}
}

func defaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.LeasesResourceLock,
		ResourceName:  "cluster-autoscaler",
	}
}

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

func parseMinMaxFlag(flag string) (int64, int64, error) {
	tokens := strings.SplitN(flag, ":", 2)
	if len(tokens) != 2 {
		return 0, 0, fmt.Errorf("wrong nodes configuration: %s", flag)
	}

	min, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to set min size: %s, expected integer, err: %v", tokens[0], err)
	}

	max, err := strconv.ParseInt(tokens[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to set max size: %s, expected integer, err: %v", tokens[1], err)
	}

	err = validateMinMaxFlag(min, max)
	if err != nil {
		return 0, 0, err
	}

	return min, max, nil
}

func validateMinMaxFlag(min, max int64) error {
	if min < 0 {
		return fmt.Errorf("min size must be greater or equal to  0")
	}
	if max < min {
		return fmt.Errorf("max size must be greater or equal to min size")
	}
	return nil
}

func minMaxFlagString(min, max int64) string {
	return fmt.Sprintf("%v:%v", min, max)
}

func parseMultipleGpuLimits(flags MultiStringFlag) ([]config.GpuLimits, error) {
	parsedFlags := make([]config.GpuLimits, 0, len(flags))
	for _, flag := range flags {
		parsedFlag, err := parseSingleGpuLimit(flag)
		if err != nil {
			return nil, err
		}
		parsedFlags = append(parsedFlags, parsedFlag)
	}
	return parsedFlags, nil
}

func parseSingleGpuLimit(limits string) (config.GpuLimits, error) {
	parts := strings.Split(limits, ":")
	if len(parts) != 3 {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit specification: %v", limits)
	}
	gpuType := parts[0]
	minVal, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit - min is not integer: %v", limits)
	}
	maxVal, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit - max is not integer: %v", limits)
	}
	if minVal < 0 {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit - min is less than 0; %v", limits)
	}
	if maxVal < 0 {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit - max is less than 0; %v", limits)
	}
	if minVal > maxVal {
		return config.GpuLimits{}, fmt.Errorf("incorrect gpu limit - min is greater than max; %v", limits)
	}
	parsedGpuLimits := config.GpuLimits{
		GpuType: gpuType,
		Min:     minVal,
		Max:     maxVal,
	}
	return parsedGpuLimits, nil
}
