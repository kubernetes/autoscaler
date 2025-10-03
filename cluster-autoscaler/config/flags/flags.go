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

package flags

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce/localssdsize"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"
	scheduler_config "k8s.io/kubernetes/pkg/scheduler/apis/config"
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
	clusterName             = flag.String("cluster-name", "", "Autoscaled cluster name, if available")
	address                 = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubernetes              = flag.String("kubernetes", "", "Kubernetes master location. Leave blank for default")
	kubeConfigFile          = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
	kubeAPIContentType      = flag.String("kube-api-content-type", "application/vnd.kubernetes.protobuf", "Content type of requests sent to apiserver.")
	kubeClientBurst         = flag.Int("kube-client-burst", rest.DefaultBurst, "Burst value for kubernetes client.")
	kubeClientQPS           = flag.Float64("kube-client-qps", float64(rest.DefaultQPS), "QPS value for kubernetes client.")
	cloudConfig             = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	namespace               = flag.String("namespace", "kube-system", "Namespace in which cluster-autoscaler run.")
	enforceNodeGroupMinSize = flag.Bool("enforce-node-group-min-size", false, "Should CA scale up the node group to the configured min size if needed.")
	scaleDownEnabled        = flag.Bool("scale-down-enabled", true, "Should CA scale down the cluster")
	scaleDownUnreadyEnabled = flag.Bool("scale-down-unready-enabled", true, "Should CA scale down unready nodes of the cluster")
	scaleDownDelayAfterAdd  = flag.Duration("scale-down-delay-after-add", 10*time.Minute,
		"How long after scale up that scale down evaluation resumes")
	scaleDownDelayTypeLocal = flag.Bool("scale-down-delay-type-local", false,
		"Should --scale-down-delay-after-* flags be applied locally per nodegroup or globally across all nodegroups")
	scaleDownDelayAfterDelete = flag.Duration("scale-down-delay-after-delete", 0,
		"How long after node deletion that scale down evaluation resumes")
	scaleDownDelayAfterFailure = flag.Duration("scale-down-delay-after-failure", config.DefaultScaleDownDelayAfterFailure,
		"How long after scale down failure that scale down evaluation resumes")
	scaleDownUnneededTime = flag.Duration("scale-down-unneeded-time", config.DefaultScaleDownUnneededTime,
		"How long a node should be unneeded before it is eligible for scale down")
	scaleDownUnreadyTime = flag.Duration("scale-down-unready-time", config.DefaultScaleDownUnreadyTime,
		"How long an unready node should be unneeded before it is eligible for scale down")
	scaleDownUtilizationThreshold = flag.Float64("scale-down-utilization-threshold", config.DefaultScaleDownUtilizationThreshold,
		"The maximum value between the sum of cpu requests and sum of memory requests of all pods running on the node divided by node's corresponding allocatable resource, below which a node can be considered for scale down")
	scaleDownGpuUtilizationThreshold = flag.Float64("scale-down-gpu-utilization-threshold", config.DefaultScaleDownGpuUtilizationThreshold,
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
	schedulerConfigFile         = flag.String(config.SchedulerConfigFileFlag, "", "scheduler-config allows changing configuration of in-tree scheduler plugins acting on PreFilter and Filter extension points")
	nodeDeletionDelayTimeout    = flag.Duration("node-deletion-delay-timeout", 2*time.Minute, "Maximum time CA waits for removing delay-deletion.cluster-autoscaler.kubernetes.io/ annotations before deleting the node.")
	nodeDeletionBatcherInterval = flag.Duration("node-deletion-batcher-interval", 0*time.Second, "How long CA ScaleDown gather nodes to delete them in batch.")
	scanInterval                = flag.Duration("scan-interval", config.DefaultScanInterval, "How often cluster is reevaluated for scale up or down")
	maxNodesTotal               = flag.Int("max-nodes-total", 0, "Maximum number of nodes in all node groups. Cluster autoscaler will not grow the cluster beyond this number.")
	coresTotal                  = flag.String("cores-total", minMaxFlagString(0, config.DefaultMaxClusterCores), "Minimum and maximum number of cores in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers.")
	memoryTotal                 = flag.String("memory-total", minMaxFlagString(0, config.DefaultMaxClusterMemory), "Minimum and maximum number of gigabytes of memory in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers.")
	gpuTotal                    = multiStringFlag("gpu-total", "Minimum and maximum number of different GPUs in cluster, in the format <gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers. Can be passed multiple times. CURRENTLY THIS FLAG ONLY WORKS ON GKE.")
	cloudProviderFlag           = flag.String("cloud-provider", cloudBuilder.DefaultCloudProvider,
		"Cloud provider type. Available values: ["+strings.Join(cloudBuilder.AvailableCloudProviders, ",")+"]")
	maxBulkSoftTaintCount      = flag.Int("max-bulk-soft-taint-count", 10, "Maximum number of nodes that can be tainted/untainted PreferNoSchedule at the same time. Set to 0 to turn off such tainting.")
	maxBulkSoftTaintTime       = flag.Duration("max-bulk-soft-taint-time", 3*time.Second, "Maximum duration of tainting/untainting nodes as PreferNoSchedule at the same time.")
	maxGracefulTerminationFlag = flag.Int("max-graceful-termination-sec", 10*60, "Maximum number of seconds CA waits for pod termination when trying to scale down a node. "+
		"This flag is mutually exclusion with drain-priority-config flag which allows more configuration options.")
	maxTotalUnreadyPercentage = flag.Float64("max-total-unready-percentage", 45, "Maximum percentage of unready nodes in the cluster.  After this is exceeded, CA halts operations")
	okTotalUnreadyCount       = flag.Int("ok-total-unready-count", 3, "Number of allowed unready nodes, irrespective of max-total-unready-percentage")
	scaleUpFromZero           = flag.Bool("scale-up-from-zero", true, "Should CA scale up when there are 0 ready nodes.")
	parallelScaleUp           = flag.Bool("parallel-scale-up", false, "Whether to allow parallel node groups scale up. Experimental: may not work on some cloud providers, enable at your own risk.")
	maxNodeProvisionTime      = flag.Duration("max-node-provision-time", 15*time.Minute, "The default maximum time CA waits for node to be provisioned - the value can be overridden per node group")
	maxPodEvictionTime        = flag.Duration("max-pod-eviction-time", 2*time.Minute, "Maximum time CA tries to evict a pod before giving up")
	nodeGroupsFlag            = multiStringFlag(
		"nodes",
		"sets min,max size and other configuration data for a node group in a format accepted by cloud provider. Can be used multiple times. Format: <min>:<max>:<other...>")
	nodeGroupAutoDiscoveryFlag = multiStringFlag(
		"node-group-auto-discovery",
		"One or more definition(s) of node group auto-discovery. "+
			"A definition is expressed `<name of discoverer>:[<key>[=<value>]]`. "+
			"The `aws`, `gce`, and `azure` cloud providers are currently supported. AWS matches by ASG tags, e.g. `asg:tag=tagKey,anotherTagKey`. "+
			"GCE matches by IG name prefix, and requires you to specify min and max nodes per IG, e.g. `mig:namePrefix=pfx,min=0,max=10` "+
			"Azure matches by VMSS tags, similar to AWS. And you can optionally specify a default min and max size, e.g. `label:tag=tagKey,anotherTagKey=bar,min=0,max=600`. "+
			"Can be used multiple times.")

	estimatorFlag = flag.String("estimator", estimator.BinpackingEstimatorName,
		"Type of resource estimator to be used in scale up. Available values: ["+strings.Join(estimator.AvailableEstimators, ",")+"]")

	expanderFlag = flag.String("expander", expander.LeastWasteExpanderName, "Type of node group expander to be used in scale up. Available values: ["+strings.Join(expander.AvailableExpanders, ",")+"]. Specifying multiple values separated by commas will call the expanders in succession until there is only one option remaining. Ties still existing after this process are broken randomly.")

	grpcExpanderCert = flag.String("grpc-expander-cert", "", "Path to cert used by gRPC server over TLS")
	grpcExpanderURL  = flag.String("grpc-expander-url", "", "URL to reach gRPC expander server.")

	ignoreDaemonSetsUtilization = flag.Bool("ignore-daemonsets-utilization", false,
		"Should CA ignore DaemonSet pods when calculating resource utilization for scaling down")
	ignoreMirrorPodsUtilization = flag.Bool("ignore-mirror-pods-utilization", false,
		"Should CA ignore Mirror pods when calculating resource utilization for scaling down")

	writeStatusConfigMapFlag     = flag.Bool("write-status-configmap", true, "Should CA write status information to a configmap")
	statusConfigMapName          = flag.String("status-config-map-name", "cluster-autoscaler-status", "Status configmap name")
	maxInactivityTimeFlag        = flag.Duration("max-inactivity", 10*time.Minute, "Maximum time from last recorded autoscaler activity before automatic restart")
	maxBinpackingTimeFlag        = flag.Duration("max-binpacking-time", 5*time.Minute, "Maximum time spend on binpacking for a single scale-up. If binpacking is limited by this, scale-up will continue with the already calculated scale-up options.")
	maxFailingTimeFlag           = flag.Duration("max-failing-time", 15*time.Minute, "Maximum time from last recorded successful autoscaler run before automatic restart")
	balanceSimilarNodeGroupsFlag = flag.Bool("balance-similar-node-groups", false, "Detect similar node groups and balance the number of nodes between them")

	unremovableNodeRecheckTimeout = flag.Duration("unremovable-node-recheck-timeout", 5*time.Minute, "The timeout before we check again a node that couldn't be removed before")
	expendablePodsPriorityCutoff  = flag.Int("expendable-pods-priority-cutoff", -10, "Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.")
	regional                      = flag.Bool("regional", false, "Cluster is regional.")
	newPodScaleUpDelay            = flag.Duration("new-pod-scale-up-delay", 0*time.Second, "Pods less than this old will not be considered for scale-up. Can be increased for individual pods through annotation 'cluster-autoscaler.kubernetes.io/pod-scale-up-delay'.")

	startupTaintsFlag         = multiStringFlag("startup-taint", "Specifies a taint to ignore in node templates when considering to scale a node group (Equivalent to ignore-taint)")
	statusTaintsFlag          = multiStringFlag("status-taint", "Specifies a taint to ignore in node templates when considering to scale a node group but nodes will not be treated as unready")
	balancingIgnoreLabelsFlag = multiStringFlag("balancing-ignore-label", "Specifies a label to ignore in addition to the basic and cloud-provider set of labels when comparing if two node groups are similar")
	balancingLabelsFlag       = multiStringFlag("balancing-label", "Specifies a label to use for comparing if two node groups are similar, rather than the built in heuristics. Setting this flag disables all other comparison logic, and cannot be combined with --balancing-ignore-label.")
	awsUseStaticInstanceList  = flag.Bool("aws-use-static-instance-list", false, "Should CA fetch instance types in runtime or use a static list. AWS only")

	// GCE specific flags
	concurrentGceRefreshes             = flag.Int("gce-concurrent-refreshes", 1, "Maximum number of concurrent refreshes per cloud object type.")
	gceMigInstancesMinRefreshWaitTime  = flag.Duration("gce-mig-instances-min-refresh-wait-time", 5*time.Second, "The minimum time which needs to pass before GCE MIG instances from a given MIG can be refreshed.")
	bulkGceMigInstancesListingEnabled  = flag.Bool("bulk-mig-instances-listing-enabled", false, "Fetch GCE mig instances in bulk instead of per mig")
	enableProfiling                    = flag.Bool("profiling", false, "Is debug/pprof endpoint enabled")
	clusterAPICloudConfigAuthoritative = flag.Bool("clusterapi-cloud-config-authoritative", false, "Treat the cloud-config flag authoritatively (do not fallback to using kubeconfig flag). ClusterAPI only")
	cordonNodeBeforeTerminate          = flag.Bool("cordon-node-before-terminating", true, "Should CA cordon nodes before terminating during downscale process")
	daemonSetEvictionForEmptyNodes     = flag.Bool("daemonset-eviction-for-empty-nodes", false, "DaemonSet pods will be gracefully terminated from empty nodes")
	daemonSetEvictionForOccupiedNodes  = flag.Bool("daemonset-eviction-for-occupied-nodes", true, "DaemonSet pods will be gracefully terminated from non-empty nodes")
	userAgent                          = flag.String("user-agent", "cluster-autoscaler", "User agent used for HTTP calls.")
	emitPerNodeGroupMetrics            = flag.Bool("emit-per-nodegroup-metrics", false, "If true, emit per node group metrics.")
	debuggingSnapshotEnabled           = flag.Bool("debugging-snapshot-enabled", false, "Whether the debugging snapshot of cluster autoscaler feature is enabled")
	nodeInfoCacheExpireTime            = flag.Duration("node-info-cache-expire-time", 87600*time.Hour, "Node Info cache expire time for each item. Default value is 10 years.")

	initialNodeGroupBackoffDuration = flag.Duration("initial-node-group-backoff-duration", 5*time.Minute,
		"initialNodeGroupBackoffDuration is the duration of first backoff after a new node failed to start.")
	maxNodeGroupBackoffDuration = flag.Duration("max-node-group-backoff-duration", 30*time.Minute,
		"maxNodeGroupBackoffDuration is the maximum backoff duration for a NodeGroup after new nodes failed to start.")
	nodeGroupBackoffResetTimeout = flag.Duration("node-group-backoff-reset-timeout", 3*time.Hour,
		"nodeGroupBackoffResetTimeout is the time after last failed scale-up when the backoff duration is reset.")
	maxScaleDownParallelismFlag             = flag.Int("max-scale-down-parallelism", 10, "Maximum number of nodes (both empty and needing drain) that can be deleted in parallel.")
	maxDrainParallelismFlag                 = flag.Int("max-drain-parallelism", 1, "Maximum number of nodes needing drain, that can be drained and deleted in parallel.")
	recordDuplicatedEvents                  = flag.Bool("record-duplicated-events", false, "enable duplication of similar events within a 5 minute window.")
	maxNodesPerScaleUp                      = flag.Int("max-nodes-per-scaleup", 1000, "Max nodes added in a single scale-up. This is intended strictly for optimizing CA algorithm latency and not a tool to rate-limit scale-up throughput.")
	maxNodeGroupBinpackingDuration          = flag.Duration("max-nodegroup-binpacking-duration", 10*time.Second, "Maximum time that will be spent in binpacking simulation for each NodeGroup.")
	skipNodesWithSystemPods                 = flag.Bool("skip-nodes-with-system-pods", true, "If true cluster autoscaler will wait for --blocking-system-pod-distruption-timeout before deleting nodes with pods from kube-system (except for DaemonSet or mirror pods)")
	skipNodesWithLocalStorage               = flag.Bool("skip-nodes-with-local-storage", true, "If true cluster autoscaler will never delete nodes with pods with local storage, e.g. EmptyDir or HostPath")
	skipNodesWithCustomControllerPods       = flag.Bool("skip-nodes-with-custom-controller-pods", true, "If true cluster autoscaler will never delete nodes with pods owned by custom controllers")
	minReplicaCount                         = flag.Int("min-replica-count", 0, "Minimum number or replicas that a replica set or replication controller should have to allow their pods deletion in scale down")
	bspDisruptionTimeout                    = flag.Duration("blocking-system-pod-distruption-timeout", time.Hour, "The timeout after which CA will evict non-pdb-assigned blocking system pods, applicable only when --skip-nodes-with-system-pods is set to true")
	nodeDeleteDelayAfterTaint               = flag.Duration("node-delete-delay-after-taint", 5*time.Second, "How long to wait before deleting a node after tainting it")
	scaleDownSimulationTimeout              = flag.Duration("scale-down-simulation-timeout", 30*time.Second, "How long should we run scale down simulation.")
	maxCapacityMemoryDifferenceRatio        = flag.Float64("memory-difference-ratio", config.DefaultMaxCapacityMemoryDifferenceRatio, "Maximum difference in memory capacity between two similar node groups to be considered for balancing. Value is a ratio of the smaller node group's memory capacity.")
	maxFreeDifferenceRatio                  = flag.Float64("max-free-difference-ratio", config.DefaultMaxFreeDifferenceRatio, "Maximum difference in free resources between two similar node groups to be considered for balancing. Value is a ratio of the smaller node group's free resource.")
	maxAllocatableDifferenceRatio           = flag.Float64("max-allocatable-difference-ratio", config.DefaultMaxAllocatableDifferenceRatio, "Maximum difference in allocatable resources between two similar node groups to be considered for balancing. Value is a ratio of the smaller node group's allocatable resource.")
	forceDaemonSets                         = flag.Bool("force-ds", false, "Blocks scale-up of node groups too small for all suitable Daemon Sets pods.")
	dynamicNodeDeleteDelayAfterTaintEnabled = flag.Bool("dynamic-node-delete-delay-after-taint-enabled", false, "Enables dynamic adjustment of NodeDeleteDelayAfterTaint based of the latency between CA and api-server")
	bypassedSchedulers                      = pflag.StringSlice("bypassed-scheduler-names", []string{}, "Names of schedulers to bypass. If set to non-empty value, CA will not wait for pods to reach a certain age before triggering a scale-up.")
	drainPriorityConfig                     = flag.String("drain-priority-config", "",
		"List of ',' separated pairs (priority:terminationGracePeriodSeconds) of integers separated by ':' enables priority evictor. Priority evictor groups pods into priority groups based on pod priority and evict pods in the ascending order of group priorities"+
			"--max-graceful-termination-sec flag should not be set when this flag is set. Not setting this flag will use unordered evictor by default."+
			"Priority evictor reuses the concepts of drain logic in kubelet(https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/2712-pod-priority-based-graceful-node-shutdown#migration-from-the-node-graceful-shutdown-feature)."+
			"Eg. flag usage:  '10000:20,1000:100,0:60'")
	provisioningRequestsEnabled                  = flag.Bool("enable-provisioning-requests", false, "Whether the clusterautoscaler will be handling the ProvisioningRequest CRs.")
	provisioningRequestInitialBackoffTime        = flag.Duration("provisioning-request-initial-backoff-time", 1*time.Minute, "Initial backoff time for ProvisioningRequest retry after failed ScaleUp.")
	provisioningRequestMaxBackoffTime            = flag.Duration("provisioning-request-max-backoff-time", 10*time.Minute, "Max backoff time for ProvisioningRequest retry after failed ScaleUp.")
	provisioningRequestMaxBackoffCacheSize       = flag.Int("provisioning-request-max-backoff-cache-size", 1000, "Max size for ProvisioningRequest cache size used for retry backoff mechanism.")
	frequentLoopsEnabled                         = flag.Bool("frequent-loops-enabled", false, "Whether clusterautoscaler triggers new iterations more frequently when it's needed")
	asyncNodeGroupsEnabled                       = flag.Bool("async-node-groups", false, "Whether clusterautoscaler creates and deletes node groups asynchronously. Experimental: requires cloud provider supporting async node group operations, enable at your own risk.")
	proactiveScaleupEnabled                      = flag.Bool("enable-proactive-scaleup", false, "Whether to enable/disable proactive scale-ups, defaults to false")
	podInjectionLimit                            = flag.Int("pod-injection-limit", 5000, "Limits total number of pods while injecting fake pods. If unschedulable pods already exceeds the limit, pod injection is disabled but pods are not truncated.")
	checkCapacityBatchProcessing                 = flag.Bool("check-capacity-batch-processing", false, "Whether to enable batch processing for check capacity requests.")
	checkCapacityProvisioningRequestMaxBatchSize = flag.Int("check-capacity-provisioning-request-max-batch-size", 10, "Maximum number of provisioning requests to process in a single batch.")
	checkCapacityProvisioningRequestBatchTimebox = flag.Duration("check-capacity-provisioning-request-batch-timebox", 10*time.Second, "Maximum time to process a batch of provisioning requests.")
	forceDeleteLongUnregisteredNodes             = flag.Bool("force-delete-unregistered-nodes", false, "Whether to enable force deletion of long unregistered nodes, regardless of the min size of the node group the belong to.")
	forceDeleteFailedNodes                       = flag.Bool("force-delete-failed-nodes", false, "Whether to enable force deletion of failed nodes, regardless of the min size of the node group the belong to.")
	enableDynamicResourceAllocation              = flag.Bool("enable-dynamic-resource-allocation", false, "Whether logic for handling DRA (Dynamic Resource Allocation) objects is enabled.")
	clusterSnapshotParallelism                   = flag.Int("cluster-snapshot-parallelism", 16, "Maximum parallelism of cluster snapshot creation.")
	checkCapacityProcessorInstance               = flag.String("check-capacity-processor-instance", "", "Name of the processor instance. Only ProvisioningRequests that define this name in their parameters with the key \"processorInstance\" will be processed by this CA instance. It only refers to check capacity ProvisioningRequests, but if not empty, best-effort atomic ProvisioningRequests processing is disabled in this instance. Not recommended: Until CA 1.35, ProvisioningRequests with this name as prefix in their class will be also processed.")
	nodeDeletionCandidateTTL                     = flag.Duration("node-deletion-candidate-ttl", time.Duration(0), "Maximum time a node can be marked as removable before the marking becomes stale. This sets the TTL of Cluster-Autoscaler's state if the Cluste-Autoscaler deployment becomes inactive")
	capacitybufferControllerEnabled              = flag.Bool("capacity-buffer-controller-enabled", false, "Whether to enable the default controller for capacity buffers or not")
	capacitybufferPodInjectionEnabled            = flag.Bool("capacity-buffer-pod-injection-enabled", false, "Whether to enable pod list processor that processes ready capacity buffers and injects fake pods accordingly")

	// Deprecated flags
	ignoreTaintsFlag = multiStringFlag("ignore-taint", "Specifies a taint to ignore in node templates when considering to scale a node group (Deprecated, use startup-taints instead)")
)

var autoscalingOptions *config.AutoscalingOptions

// AutoscalingOptions returns the singleton instance of AutoscalingOptions, initializing it if necessary.
func AutoscalingOptions() config.AutoscalingOptions {
	if autoscalingOptions == nil {
		newAutoscalingOptions := createAutoscalingOptions()
		autoscalingOptions = &newAutoscalingOptions
	}
	return *autoscalingOptions
}

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

	var parsedSchedConfig *scheduler_config.KubeSchedulerConfiguration
	// if scheduler config flag was set by the user
	if pflag.CommandLine.Changed(config.SchedulerConfigFileFlag) {
		parsedSchedConfig, err = scheduler_util.ConfigFromPath(*schedulerConfigFile)
	}
	if err != nil {
		klog.Fatalf("Failed to get scheduler config: %v", err)
	}

	if pflag.CommandLine.Changed("drain-priority-config") && pflag.CommandLine.Changed("max-graceful-termination-sec") {
		klog.Fatalf("Invalid configuration, could not use --drain-priority-config together with --max-graceful-termination-sec")
	}

	var drainPriorityConfigMap []kubelet_config.ShutdownGracePeriodByPodPriority
	if pflag.CommandLine.Changed("drain-priority-config") {
		drainPriorityConfigMap = parseShutdownGracePeriodsAndPriorities(*drainPriorityConfig)
		if len(drainPriorityConfigMap) == 0 {
			klog.Fatalf("Invalid configuration, parsing --drain-priority-config")
		}
	}

	return config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    *scaleDownUtilizationThreshold,
			ScaleDownGpuUtilizationThreshold: *scaleDownGpuUtilizationThreshold,
			ScaleDownUnneededTime:            *scaleDownUnneededTime,
			ScaleDownUnreadyTime:             *scaleDownUnreadyTime,
			IgnoreDaemonSetsUtilization:      *ignoreDaemonSetsUtilization,
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
		IgnoreMirrorPodsUtilization:      *ignoreMirrorPodsUtilization,
		MaxBulkSoftTaintCount:            *maxBulkSoftTaintCount,
		MaxBulkSoftTaintTime:             *maxBulkSoftTaintTime,
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
		ScaleDownDelayTypeLocal:          *scaleDownDelayTypeLocal,
		ScaleDownDelayAfterDelete:        *scaleDownDelayAfterDelete,
		ScaleDownDelayAfterFailure:       *scaleDownDelayAfterFailure,
		ScaleDownEnabled:                 *scaleDownEnabled,
		ScaleDownUnreadyEnabled:          *scaleDownUnreadyEnabled,
		ScaleDownNonEmptyCandidatesCount: *scaleDownNonEmptyCandidatesCount,
		ScaleDownCandidatesPoolRatio:     *scaleDownCandidatesPoolRatio,
		ScaleDownCandidatesPoolMinCount:  *scaleDownCandidatesPoolMinCount,
		DrainPriorityConfig:              drainPriorityConfigMap,
		SchedulerConfig:                  parsedSchedConfig,
		WriteStatusConfigMap:             *writeStatusConfigMapFlag,
		StatusConfigMapName:              *statusConfigMapName,
		BalanceSimilarNodeGroups:         *balanceSimilarNodeGroupsFlag,
		ConfigNamespace:                  *namespace,
		ClusterName:                      *clusterName,
		UnremovableNodeRecheckTimeout:    *unremovableNodeRecheckTimeout,
		ExpendablePodsPriorityCutoff:     *expendablePodsPriorityCutoff,
		Regional:                         *regional,
		NewPodScaleUpDelay:               *newPodScaleUpDelay,
		StartupTaints:                    append(*ignoreTaintsFlag, *startupTaintsFlag...),
		StatusTaints:                     *statusTaintsFlag,
		BalancingExtraIgnoredLabels:      *balancingIgnoreLabelsFlag,
		BalancingLabels:                  *balancingLabelsFlag,
		KubeClientOpts: config.KubeClientOptions{
			Master:          *kubernetes,
			KubeConfigPath:  *kubeConfigFile,
			APIContentType:  *kubeAPIContentType,
			KubeClientBurst: int(*kubeClientBurst),
			KubeClientQPS:   float32(*kubeClientQPS),
		},
		NodeDeletionDelayTimeout: *nodeDeletionDelayTimeout,
		AWSUseStaticInstanceList: *awsUseStaticInstanceList,
		GCEOptions: config.GCEOptions{
			ConcurrentRefreshes:            *concurrentGceRefreshes,
			MigInstancesMinRefreshWaitTime: *gceMigInstancesMinRefreshWaitTime,
			LocalSSDDiskSizeProvider:       localssdsize.NewSimpleLocalSSDProvider(),
			BulkMigInstancesListingEnabled: *bulkGceMigInstancesListingEnabled,
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
		MaxBinpackingTime:                  *maxBinpackingTimeFlag,
		NodeDeletionBatcherInterval:        *nodeDeletionBatcherInterval,
		SkipNodesWithSystemPods:            *skipNodesWithSystemPods,
		SkipNodesWithLocalStorage:          *skipNodesWithLocalStorage,
		MinReplicaCount:                    *minReplicaCount,
		BspDisruptionTimeout:               *bspDisruptionTimeout,
		NodeDeleteDelayAfterTaint:          *nodeDeleteDelayAfterTaint,
		ScaleDownSimulationTimeout:         *scaleDownSimulationTimeout,
		SkipNodesWithCustomControllerPods:  *skipNodesWithCustomControllerPods,
		NodeGroupSetRatios: config.NodeGroupDifferenceRatios{
			MaxCapacityMemoryDifferenceRatio: *maxCapacityMemoryDifferenceRatio,
			MaxAllocatableDifferenceRatio:    *maxAllocatableDifferenceRatio,
			MaxFreeDifferenceRatio:           *maxFreeDifferenceRatio,
		},
		DynamicNodeDeleteDelayAfterTaintEnabled:      *dynamicNodeDeleteDelayAfterTaintEnabled,
		BypassedSchedulers:                           scheduler_util.GetBypassedSchedulersMap(*bypassedSchedulers),
		ProvisioningRequestEnabled:                   *provisioningRequestsEnabled,
		AsyncNodeGroupsEnabled:                       *asyncNodeGroupsEnabled,
		ProvisioningRequestInitialBackoffTime:        *provisioningRequestInitialBackoffTime,
		ProvisioningRequestMaxBackoffTime:            *provisioningRequestMaxBackoffTime,
		ProvisioningRequestMaxBackoffCacheSize:       *provisioningRequestMaxBackoffCacheSize,
		CheckCapacityBatchProcessing:                 *checkCapacityBatchProcessing,
		CheckCapacityProvisioningRequestMaxBatchSize: *checkCapacityProvisioningRequestMaxBatchSize,
		CheckCapacityProvisioningRequestBatchTimebox: *checkCapacityProvisioningRequestBatchTimebox,
		ForceDeleteLongUnregisteredNodes:             *forceDeleteLongUnregisteredNodes,
		ForceDeleteFailedNodes:                       *forceDeleteFailedNodes,
		DynamicResourceAllocationEnabled:             *enableDynamicResourceAllocation,
		ClusterSnapshotParallelism:                   *clusterSnapshotParallelism,
		CheckCapacityProcessorInstance:               *checkCapacityProcessorInstance,
		MaxInactivityTime:                            *maxInactivityTimeFlag,
		MaxFailingTime:                               *maxFailingTimeFlag,
		DebuggingSnapshotEnabled:                     *debuggingSnapshotEnabled,
		EnableProfiling:                              *enableProfiling,
		Address:                                      *address,
		EmitPerNodeGroupMetrics:                      *emitPerNodeGroupMetrics,
		FrequentLoopsEnabled:                         *frequentLoopsEnabled,
		ScanInterval:                                 *scanInterval,
		ForceDaemonSets:                              *forceDaemonSets,
		NodeInfoCacheExpireTime:                      *nodeInfoCacheExpireTime,
		ProactiveScaleupEnabled:                      *proactiveScaleupEnabled,
		PodInjectionLimit:                            *podInjectionLimit,
		NodeDeletionCandidateTTL:                     *nodeDeletionCandidateTTL,
		CapacitybufferControllerEnabled:              *capacitybufferControllerEnabled,
		CapacitybufferPodInjectionEnabled:            *capacitybufferPodInjectionEnabled,
	}
}

func minMaxFlagString(min, max int64) string {
	return fmt.Sprintf("%v:%v", min, max)
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

// parseShutdownGracePeriodsAndPriorities parse priorityGracePeriodStr and returns an array of ShutdownGracePeriodByPodPriority if succeeded.
// Otherwise, returns an empty list
func parseShutdownGracePeriodsAndPriorities(priorityGracePeriodStr string) []kubelet_config.ShutdownGracePeriodByPodPriority {
	var priorityGracePeriodMap, emptyMap []kubelet_config.ShutdownGracePeriodByPodPriority

	if priorityGracePeriodStr == "" {
		return emptyMap
	}
	priorityGracePeriodStrArr := strings.Split(priorityGracePeriodStr, ",")
	for _, item := range priorityGracePeriodStrArr {
		priorityAndPeriod := strings.Split(item, ":")
		if len(priorityAndPeriod) != 2 {
			klog.Errorf("Parsing shutdown grace periods failed because '%s' is not a priority and grace period couple separated by ':'", item)
			return emptyMap
		}
		priority, err := strconv.Atoi(priorityAndPeriod[0])
		if err != nil {
			klog.Errorf("Parsing shutdown grace periods and priorities failed: %v", err)
			return emptyMap
		}
		shutDownGracePeriod, err := strconv.Atoi(priorityAndPeriod[1])
		if err != nil {
			klog.Errorf("Parsing shutdown grace periods and priorities failed: %v", err)
			return emptyMap
		}
		priorityGracePeriodMap = append(priorityGracePeriodMap, kubelet_config.ShutdownGracePeriodByPodPriority{
			Priority:                   int32(priority),
			ShutdownGracePeriodSeconds: int64(shutDownGracePeriod),
		})
	}
	return priorityGracePeriodMap
}
