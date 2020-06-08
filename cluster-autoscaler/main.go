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
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
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
	klog "k8s.io/klog/v2"
)

// MultiStringFlag is a flag for passing multiple parameters using same flag
type MultiStringFlag []string

// String returns string representation of the node groups.
func (flag *MultiStringFlag) String() string {
	return "[" + strings.Join(*flag, " ") + "]"
}

// Set adds a new configuration.
func (flag *MultiStringFlag) Set(value string) error {
	*flag = append(*flag, value)
	return nil
}

func multiStringFlag(name string, usage string) *MultiStringFlag {
	value := new(MultiStringFlag)
	flag.Var(value, name, usage)
	return value
}

var (
	clusterName            = flag.String("cluster-name", "", "Autoscaled cluster name, if available")
	address                = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubernetes             = flag.String("kubernetes", "", "Kubernetes master location. Leave blank for default")
	kubeConfigFile         = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
	cloudConfig            = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	namespace              = flag.String("namespace", "kube-system", "Namespace in which cluster-autoscaler run.")
	scaleDownEnabled       = flag.Bool("scale-down-enabled", true, "Should CA scale down the cluster")
	scaleDownDelayAfterAdd = flag.Duration("scale-down-delay-after-add", 10*time.Minute,
		"How long after scale up that scale down evaluation resumes")
	scaleDownDelayAfterDelete = flag.Duration("scale-down-delay-after-delete", 0,
		"How long after node deletion that scale down evaluation resumes, defaults to scanInterval")
	scaleDownDelayAfterFailure = flag.Duration("scale-down-delay-after-failure", 3*time.Minute,
		"How long after scale down failure that scale down evaluation resumes")
	scaleDownUnneededTime = flag.Duration("scale-down-unneeded-time", 10*time.Minute,
		"How long a node should be unneeded before it is eligible for scale down")
	scaleDownUnreadyTime = flag.Duration("scale-down-unready-time", 20*time.Minute,
		"How long an unready node should be unneeded before it is eligible for scale down")
	scaleDownUtilizationThreshold = flag.Float64("scale-down-utilization-threshold", 0.5,
		"Sum of cpu or memory of all pods running on the node divided by node's corresponding allocatable resource, below which a node can be considered for scale down")
	scaleDownGpuUtilizationThreshold = flag.Float64("scale-down-gpu-utilization-threshold", 0.5,
		"Sum of gpu requests of all pods running on the node divided by node's allocatable resource, below which a node can be considered for scale down."+
			"Utilization calculation only cares about gpu resource for accelerator node. cpu and memory utilization will be ignored.")
	scaleDownNonEmptyCandidatesCount = flag.Int("scale-down-non-empty-candidates-count", 30,
		"Maximum number of non empty nodes considered in one iteration as candidates for scale down with drain."+
			"Lower value means better CA responsiveness but possible slower scale down latency."+
			"Higher value can affect CA performance with big clusters (hundreds of nodes)."+
			"Set to non positive value to turn this heuristic off - CA will not limit the number of nodes it considers.")
	scaleDownCandidatesPoolRatio = flag.Float64("scale-down-candidates-pool-ratio", 0.1,
		"A ratio of nodes that are considered as additional non empty candidates for"+
			"scale down when some candidates from previous iteration are no longer valid."+
			"Lower value means better CA responsiveness but possible slower scale down latency."+
			"Higher value can affect CA performance with big clusters (hundreds of nodes)."+
			"Set to 1.0 to turn this heuristics off - CA will take all nodes as additional candidates.")
	scaleDownCandidatesPoolMinCount = flag.Int("scale-down-candidates-pool-min-count", 50,
		"Minimum number of nodes that are considered as additional non empty candidates"+
			"for scale down when some candidates from previous iteration are no longer valid."+
			"When calculating the pool size for additional candidates we take"+
			"max(#nodes * scale-down-candidates-pool-ratio, scale-down-candidates-pool-min-count).")
	nodeDeletionDelayTimeout = flag.Duration("node-deletion-delay-timeout", 2*time.Minute, "Maximum time CA waits for removing delay-deletion.cluster-autoscaler.kubernetes.io/ annotations before deleting the node.")
	scanInterval             = flag.Duration("scan-interval", 10*time.Second, "How often cluster is reevaluated for scale up or down")
	maxNodesTotal            = flag.Int("max-nodes-total", 0, "Maximum number of nodes in all node groups. Cluster autoscaler will not grow the cluster beyond this number.")
	coresTotal               = flag.String("cores-total", minMaxFlagString(0, config.DefaultMaxClusterCores), "Minimum and maximum number of cores in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers.")
	memoryTotal              = flag.String("memory-total", minMaxFlagString(0, config.DefaultMaxClusterMemory), "Minimum and maximum number of gigabytes of memory in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers.")
	gpuTotal                 = multiStringFlag("gpu-total", "Minimum and maximum number of different GPUs in cluster, in the format <gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers. Can be passed multiple times. CURRENTLY THIS FLAG ONLY WORKS ON GKE.")
	cloudProviderFlag        = flag.String("cloud-provider", cloudBuilder.DefaultCloudProvider,
		"Cloud provider type. Available values: ["+strings.Join(cloudBuilder.AvailableCloudProviders, ",")+"]")
	maxBulkSoftTaintCount      = flag.Int("max-bulk-soft-taint-count", 10, "Maximum number of nodes that can be tainted/untainted PreferNoSchedule at the same time. Set to 0 to turn off such tainting.")
	maxBulkSoftTaintTime       = flag.Duration("max-bulk-soft-taint-time", 3*time.Second, "Maximum duration of tainting/untainting nodes as PreferNoSchedule at the same time.")
	maxEmptyBulkDeleteFlag     = flag.Int("max-empty-bulk-delete", 10, "Maximum number of empty nodes that can be deleted at the same time.")
	maxGracefulTerminationFlag = flag.Int("max-graceful-termination-sec", 10*60, "Maximum number of seconds CA waits for pod termination when trying to scale down a node.")
	maxTotalUnreadyPercentage  = flag.Float64("max-total-unready-percentage", 45, "Maximum percentage of unready nodes in the cluster.  After this is exceeded, CA halts operations")
	okTotalUnreadyCount        = flag.Int("ok-total-unready-count", 3, "Number of allowed unready nodes, irrespective of max-total-unready-percentage")
	scaleUpFromZero            = flag.Bool("scale-up-from-zero", true, "Should CA scale up when there 0 ready nodes.")
	maxNodeProvisionTime       = flag.Duration("max-node-provision-time", 15*time.Minute, "Maximum time CA waits for node to be provisioned")
	nodeGroupsFlag             = multiStringFlag(
		"nodes",
		"sets min,max size and other configuration data for a node group in a format accepted by cloud provider. Can be used multiple times. Format: <min>:<max>:<other...>")
	nodeGroupAutoDiscoveryFlag = multiStringFlag(
		"node-group-auto-discovery",
		"One or more definition(s) of node group auto-discovery. "+
			"A definition is expressed `<name of discoverer>:[<key>[=<value>]]`. "+
			"The `aws` and `gce` cloud providers are currently supported. AWS matches by ASG tags, e.g. `asg:tag=tagKey,anotherTagKey`. "+
			"GCE matches by IG name prefix, and requires you to specify min and max nodes per IG, e.g. `mig:namePrefix=pfx,min=0,max=10` "+
			"Can be used multiple times.")

	estimatorFlag = flag.String("estimator", estimator.BinpackingEstimatorName,
		"Type of resource estimator to be used in scale up. Available values: ["+strings.Join(estimator.AvailableEstimators, ",")+"]")

	expanderFlag = flag.String("expander", expander.RandomExpanderName,
		"Type of node group expander to be used in scale up. Available values: ["+strings.Join(expander.AvailableExpanders, ",")+"]")

	ignoreDaemonSetsUtilization = flag.Bool("ignore-daemonsets-utilization", false,
		"Should CA ignore DaemonSet pods when calculating resource utilization for scaling down")
	ignoreMirrorPodsUtilization = flag.Bool("ignore-mirror-pods-utilization", false,
		"Should CA ignore Mirror pods when calculating resource utilization for scaling down")

	writeStatusConfigMapFlag         = flag.Bool("write-status-configmap", true, "Should CA write status information to a configmap")
	maxInactivityTimeFlag            = flag.Duration("max-inactivity", 10*time.Minute, "Maximum time from last recorded autoscaler activity before automatic restart")
	maxFailingTimeFlag               = flag.Duration("max-failing-time", 15*time.Minute, "Maximum time from last recorded successful autoscaler run before automatic restart")
	balanceSimilarNodeGroupsFlag     = flag.Bool("balance-similar-node-groups", false, "Detect similar node groups and balance the number of nodes between them")
	nodeAutoprovisioningEnabled      = flag.Bool("node-autoprovisioning-enabled", false, "Should CA autoprovision node groups when needed")
	maxAutoprovisionedNodeGroupCount = flag.Int("max-autoprovisioned-node-group-count", 15, "The maximum number of autoprovisioned groups in the cluster.")

	unremovableNodeRecheckTimeout = flag.Duration("unremovable-node-recheck-timeout", 5*time.Minute, "The timeout before we check again a node that couldn't be removed before")
	expendablePodsPriorityCutoff  = flag.Int("expendable-pods-priority-cutoff", -10, "Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.")
	regional                      = flag.Bool("regional", false, "Cluster is regional.")
	newPodScaleUpDelay            = flag.Duration("new-pod-scale-up-delay", 0*time.Second, "Pods less than this old will not be considered for scale-up.")

	ignoreTaintsFlag          = multiStringFlag("ignore-taint", "Specifies a taint to ignore in node templates when considering to scale a node group")
	balancingIgnoreLabelsFlag = multiStringFlag("balancing-ignore-label", "Specifies a label to ignore in addition to the basic and cloud-provider set of labels when comparing if two node groups are similar")
	awsUseStaticInstanceList  = flag.Bool("aws-use-static-instance-list", false, "Should CA fetch instance types in runtime or use a static list. AWS only")
	enableProfiling           = flag.Bool("profiling", false, "Is debug/pprof endpoint enabled")
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
	return config.AutoscalingOptions{
		CloudConfig:                      *cloudConfig,
		CloudProviderName:                *cloudProviderFlag,
		NodeGroupAutoDiscovery:           *nodeGroupAutoDiscoveryFlag,
		MaxTotalUnreadyPercentage:        *maxTotalUnreadyPercentage,
		OkTotalUnreadyCount:              *okTotalUnreadyCount,
		ScaleUpFromZero:                  *scaleUpFromZero,
		EstimatorName:                    *estimatorFlag,
		ExpanderName:                     *expanderFlag,
		IgnoreDaemonSetsUtilization:      *ignoreDaemonSetsUtilization,
		IgnoreMirrorPodsUtilization:      *ignoreMirrorPodsUtilization,
		MaxBulkSoftTaintCount:            *maxBulkSoftTaintCount,
		MaxBulkSoftTaintTime:             *maxBulkSoftTaintTime,
		MaxEmptyBulkDelete:               *maxEmptyBulkDeleteFlag,
		MaxGracefulTerminationSec:        *maxGracefulTerminationFlag,
		MaxNodeProvisionTime:             *maxNodeProvisionTime,
		MaxNodesTotal:                    *maxNodesTotal,
		MaxCoresTotal:                    maxCoresTotal,
		MinCoresTotal:                    minCoresTotal,
		MaxMemoryTotal:                   maxMemoryTotal,
		MinMemoryTotal:                   minMemoryTotal,
		GpuTotal:                         parsedGpuTotal,
		NodeGroups:                       *nodeGroupsFlag,
		ScaleDownDelayAfterAdd:           *scaleDownDelayAfterAdd,
		ScaleDownDelayAfterDelete:        *scaleDownDelayAfterDelete,
		ScaleDownDelayAfterFailure:       *scaleDownDelayAfterFailure,
		ScaleDownEnabled:                 *scaleDownEnabled,
		ScaleDownUnneededTime:            *scaleDownUnneededTime,
		ScaleDownUnreadyTime:             *scaleDownUnreadyTime,
		ScaleDownUtilizationThreshold:    *scaleDownUtilizationThreshold,
		ScaleDownGpuUtilizationThreshold: *scaleDownGpuUtilizationThreshold,
		ScaleDownNonEmptyCandidatesCount: *scaleDownNonEmptyCandidatesCount,
		ScaleDownCandidatesPoolRatio:     *scaleDownCandidatesPoolRatio,
		ScaleDownCandidatesPoolMinCount:  *scaleDownCandidatesPoolMinCount,
		WriteStatusConfigMap:             *writeStatusConfigMapFlag,
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
		KubeConfigPath:                   *kubeConfigFile,
		NodeDeletionDelayTimeout:         *nodeDeletionDelayTimeout,
		AWSUseStaticInstanceList:         *awsUseStaticInstanceList,
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

func buildAutoscaler() (core.Autoscaler, error) {
	// Create basic config from flags.
	autoscalingOptions := createAutoscalingOptions()
	kubeClient := createKubeClient(getKubeConfig())
	eventsKubeClient := createKubeClient(getKubeConfig())

	opts := core.AutoscalerOptions{
		AutoscalingOptions: autoscalingOptions,
		ClusterSnapshot:    simulator.NewDeltaClusterSnapshot(),
		KubeClient:         kubeClient,
		EventsKubeClient:   eventsKubeClient,
	}

	opts.Processors = ca_processors.DefaultProcessors()
	opts.Processors.PodListProcessor = core.NewFilterOutSchedulablePodListProcessor()

	nodeInfoComparatorBuilder := nodegroupset.CreateGenericNodeInfoComparator
	if autoscalingOptions.CloudProviderName == cloudprovider.AzureProviderName {
		nodeInfoComparatorBuilder = nodegroupset.CreateAzureNodeInfoComparator
	} else if autoscalingOptions.CloudProviderName == cloudprovider.AwsProviderName {
		nodeInfoComparatorBuilder = nodegroupset.CreateAwsNodeInfoComparator
	}

	opts.Processors.NodeGroupSetProcessor = &nodegroupset.BalancingNodeGroupSetProcessor{
		Comparator: nodeInfoComparatorBuilder(autoscalingOptions.BalancingExtraIgnoredLabels),
	}

	// This metric should be published only once.
	metrics.UpdateNapEnabled(autoscalingOptions.NodeAutoprovisioningEnabled)

	// Create autoscaler.
	return core.NewAutoscaler(opts)
}

func run(healthCheck *metrics.HealthCheck) {
	metrics.RegisterAll()

	autoscaler, err := buildAutoscaler()
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
	kube_flag.InitFlags()
	healthCheck := metrics.NewHealthCheck(*maxInactivityTimeFlag, *maxFailingTimeFlag)

	klog.V(1).Infof("Cluster Autoscaler %s", version.ClusterAutoscalerVersion)

	go func() {
		pathRecorderMux := mux.NewPathRecorderMux("cluster-autoscaler")
		defaultMetricsHandler := legacyregistry.Handler().ServeHTTP
		pathRecorderMux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
			defaultMetricsHandler(w, req)
		})
		pathRecorderMux.HandleFunc("/health-check", healthCheck.ServeHTTP)
		if *enableProfiling {
			routes.Profiling{}.Install(pathRecorderMux)
		}
		err := http.ListenAndServe(*address, pathRecorderMux)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()

	if !leaderElection.LeaderElect {
		run(healthCheck)
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
			"cluster-autoscaler",
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity:      id,
				EventRecorder: kube_util.CreateEventRecorder(kubeClient),
			},
		)
		if err != nil {
			klog.Fatalf("Unable to create leader election lock: %v", err)
		}

		leaderelection.RunOrDie(ctx.TODO(), leaderelection.LeaderElectionConfig{
			Lock:          lock,
			LeaseDuration: leaderElection.LeaseDuration.Duration,
			RenewDeadline: leaderElection.RenewDeadline.Duration,
			RetryPeriod:   leaderElection.RetryPeriod.Duration,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ ctx.Context) {
					// Since we are committing a suicide after losing
					// mastership, we can safely ignore the argument.
					run(healthCheck)
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
