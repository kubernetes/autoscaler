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
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce/localssdsize"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/podinjection"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/emptycandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/previouscandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	provreqorchestrator "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	kube_flag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	componentopts "k8s.io/component-base/config/options"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
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
	leaseResourceName       = flag.String("lease-resource-name", "cluster-autoscaler", "The lease resource to use in leader election.")
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
		"How long after node deletion that scale down evaluation resumes, defaults to scanInterval")
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
	maxEmptyBulkDeleteFlag     = flag.Int("max-empty-bulk-delete", 10, "Maximum number of empty nodes that can be deleted at the same time. DEPRECATED: Use --max-scale-down-parallelism instead.")
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

	expanderFlag = flag.String("expander", expander.RandomExpanderName, "Type of node group expander to be used in scale up. Available values: ["+strings.Join(expander.AvailableExpanders, ",")+"]. Specifying multiple values separated by commas will call the expanders in succession until there is only one option remaining. Ties still existing after this process are broken randomly.")

	grpcExpanderCert = flag.String("grpc-expander-cert", "", "Path to cert used by gRPC server over TLS")
	grpcExpanderURL  = flag.String("grpc-expander-url", "", "URL to reach gRPC expander server.")

	ignoreDaemonSetsUtilization = flag.Bool("ignore-daemonsets-utilization", false,
		"Should CA ignore DaemonSet pods when calculating resource utilization for scaling down")
	ignoreMirrorPodsUtilization = flag.Bool("ignore-mirror-pods-utilization", false,
		"Should CA ignore Mirror pods when calculating resource utilization for scaling down")

	writeStatusConfigMapFlag         = flag.Bool("write-status-configmap", true, "Should CA write status information to a configmap")
	statusConfigMapName              = flag.String("status-config-map-name", "cluster-autoscaler-status", "Status configmap name")
	maxInactivityTimeFlag            = flag.Duration("max-inactivity", 10*time.Minute, "Maximum time from last recorded autoscaler activity before automatic restart")
	maxBinpackingTimeFlag            = flag.Duration("max-binpacking-time", 5*time.Minute, "Maximum time spend on binpacking for a single scale-up. If binpacking is limited by this, scale-up will continue with the already calculated scale-up options.")
	maxFailingTimeFlag               = flag.Duration("max-failing-time", 15*time.Minute, "Maximum time from last recorded successful autoscaler run before automatic restart")
	balanceSimilarNodeGroupsFlag     = flag.Bool("balance-similar-node-groups", false, "Detect similar node groups and balance the number of nodes between them")
	nodeAutoprovisioningEnabled      = flag.Bool("node-autoprovisioning-enabled", false, "Should CA autoprovision node groups when needed.This flag is deprecated and will be removed in future releases.")
	maxAutoprovisionedNodeGroupCount = flag.Int("max-autoprovisioned-node-group-count", 15, "The maximum number of autoprovisioned groups in the cluster.This flag is deprecated and will be removed in future releases.")

	unremovableNodeRecheckTimeout = flag.Duration("unremovable-node-recheck-timeout", 5*time.Minute, "The timeout before we check again a node that couldn't be removed before")
	expendablePodsPriorityCutoff  = flag.Int("expendable-pods-priority-cutoff", -10, "Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.")
	regional                      = flag.Bool("regional", false, "Cluster is regional.")
	newPodScaleUpDelay            = flag.Duration("new-pod-scale-up-delay", 0*time.Second, "Pods less than this old will not be considered for scale-up. Can be increased for individual pods through annotation 'cluster-autoscaler.kubernetes.io/pod-scale-up-delay'.")

	ignoreTaintsFlag          = multiStringFlag("ignore-taint", "Specifies a taint to ignore in node templates when considering to scale a node group (Deprecated, use startup-taints instead)")
	startupTaintsFlag         = multiStringFlag("startup-taint", "Specifies a taint to ignore in node templates when considering to scale a node group (Equivalent to ignore-taint)")
	statusTaintsFlag          = multiStringFlag("status-taint", "Specifies a taint to ignore in node templates when considering to scale a node group but nodes will not be treated as unready")
	balancingIgnoreLabelsFlag = multiStringFlag("balancing-ignore-label", "Specifies a label to ignore in addition to the basic and cloud-provider set of labels when comparing if two node groups are similar")
	balancingLabelsFlag       = multiStringFlag("balancing-label", "Specifies a label to use for comparing if two node groups are similar, rather than the built in heuristics. Setting this flag disables all other comparison logic, and cannot be combined with --balancing-ignore-label.")
	awsUseStaticInstanceList  = flag.Bool("aws-use-static-instance-list", false, "Should CA fetch instance types in runtime or use a static list. AWS only")

	// GCE specific flags
	concurrentGceRefreshes             = flag.Int("gce-concurrent-refreshes", 1, "Maximum number of concurrent refreshes per cloud object type.")
	gceMigInstancesMinRefreshWaitTime  = flag.Duration("gce-mig-instances-min-refresh-wait-time", 5*time.Second, "The minimum time which needs to pass before GCE MIG instances from a given MIG can be refreshed.")
	_                                  = flag.Bool("gce-expander-ephemeral-storage-support", true, "Whether scale-up takes ephemeral storage resources into account for GCE cloud provider (Deprecated, to be removed in 1.30+)")
	bulkGceMigInstancesListingEnabled  = flag.Bool("bulk-mig-instances-listing-enabled", false, "Fetch GCE mig instances in bulk instead of per mig")
	enableProfiling                    = flag.Bool("profiling", false, "Is debug/pprof endpoint enabled")
	clusterAPICloudConfigAuthoritative = flag.Bool("clusterapi-cloud-config-authoritative", false, "Treat the cloud-config flag authoritatively (do not fallback to using kubeconfig flag). ClusterAPI only")
	cordonNodeBeforeTerminate          = flag.Bool("cordon-node-before-terminating", false, "Should CA cordon nodes before terminating during downscale process")
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
	skipNodesWithSystemPods                 = flag.Bool("skip-nodes-with-system-pods", true, "If true cluster autoscaler will never delete nodes with pods from kube-system (except for DaemonSet or mirror pods)")
	skipNodesWithLocalStorage               = flag.Bool("skip-nodes-with-local-storage", true, "If true cluster autoscaler will never delete nodes with pods with local storage, e.g. EmptyDir or HostPath")
	skipNodesWithCustomControllerPods       = flag.Bool("skip-nodes-with-custom-controller-pods", true, "If true cluster autoscaler will never delete nodes with pods owned by custom controllers")
	minReplicaCount                         = flag.Int("min-replica-count", 0, "Minimum number or replicas that a replica set or replication controller should have to allow their pods deletion in scale down")
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
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
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

	// in order to avoid inconsistent deletion thresholds for the legacy planner and the new actuator, the max-empty-bulk-delete,
	// and max-scale-down-parallelism flags must be set to the same value.
	if isFlagPassed("max-empty-bulk-delete") && !isFlagPassed("max-scale-down-parallelism") {
		*maxScaleDownParallelismFlag = *maxEmptyBulkDeleteFlag
		klog.Warning("The max-empty-bulk-delete flag will be deprecated in k8s version 1.29. Please use max-scale-down-parallelism instead.")
		klog.Infof("Setting max-scale-down-parallelism to %d, based on the max-empty-bulk-delete value %d", *maxScaleDownParallelismFlag, *maxEmptyBulkDeleteFlag)
	} else if !isFlagPassed("max-empty-bulk-delete") && isFlagPassed("max-scale-down-parallelism") {
		*maxEmptyBulkDeleteFlag = *maxScaleDownParallelismFlag
	}

	if isFlagPassed("node-autoprovisioning-enabled") {
		klog.Warning("The node-autoprovisioning-enabled flag is deprecated and will be removed in k8s version 1.31.")
	}

	if isFlagPassed("max-autoprovisioned-node-group-count") {
		klog.Warning("The max-autoprovisioned-node-group-count flag is deprecated and will be removed in k8s version 1.31.")
	}

	var parsedSchedConfig *scheduler_config.KubeSchedulerConfiguration
	// if scheduler config flag was set by the user
	if pflag.CommandLine.Changed(config.SchedulerConfigFileFlag) {
		parsedSchedConfig, err = scheduler_util.ConfigFromPath(*schedulerConfigFile)
	}
	if err != nil {
		klog.Fatalf("Failed to get scheduler config: %v", err)
	}

	if isFlagPassed("drain-priority-config") && isFlagPassed("max-graceful-termination-sec") {
		klog.Fatalf("Invalid configuration, could not use --drain-priority-config together with --max-graceful-termination-sec")
	}

	var drainPriorityConfigMap []kubelet_config.ShutdownGracePeriodByPodPriority
	if isFlagPassed("drain-priority-config") {
		drainPriorityConfigMap = actuation.ParseShutdownGracePeriodsAndPriorities(*drainPriorityConfig)
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
		NodeAutoprovisioningEnabled:      *nodeAutoprovisioningEnabled,
		MaxAutoprovisionedNodeGroupCount: *maxAutoprovisionedNodeGroupCount,
		UnremovableNodeRecheckTimeout:    *unremovableNodeRecheckTimeout,
		ExpendablePodsPriorityCutoff:     *expendablePodsPriorityCutoff,
		Regional:                         *regional,
		NewPodScaleUpDelay:               *newPodScaleUpDelay,
		StartupTaints:                    append(*ignoreTaintsFlag, *startupTaintsFlag...),
		StatusTaints:                     *statusTaintsFlag,
		BalancingExtraIgnoredLabels:      *balancingIgnoreLabelsFlag,
		BalancingLabels:                  *balancingLabelsFlag,
		KubeClientOpts: config.KubeClientOptions{
			Master:         *kubernetes,
			KubeConfigPath: *kubeConfigFile,
			APIContentType: *kubeAPIContentType,
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
	}
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

func buildAutoscaler(context ctx.Context, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) (core.Autoscaler, *loop.LoopTrigger, error) {
	// Create basic config from flags.
	autoscalingOptions := createAutoscalingOptions()

	autoscalingOptions.KubeClientOpts.KubeClientBurst = int(*kubeClientBurst)
	autoscalingOptions.KubeClientOpts.KubeClientQPS = float32(*kubeClientQPS)
	kubeClient := kube_util.CreateKubeClient(autoscalingOptions.KubeClientOpts)

	// Informer transform to trim ManagedFields for memory efficiency.
	trim := func(obj interface{}) (interface{}, error) {
		if accessor, err := meta.Accessor(obj); err == nil {
			accessor.SetManagedFields(nil)
		}
		return obj, nil
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, informers.WithTransform(trim))

	fwHandle, err := framework.NewHandle(informerFactory, autoscalingOptions.SchedulerConfig)
	if err != nil {
		return nil, nil, err
	}
	deleteOptions := options.NewNodeDeleteOptions(autoscalingOptions)
	drainabilityRules := rules.Default(deleteOptions)

	opts := core.AutoscalerOptions{
		AutoscalingOptions:   autoscalingOptions,
		FrameworkHandle:      fwHandle,
		ClusterSnapshot:      clustersnapshot.NewDeltaClusterSnapshot(),
		KubeClient:           kubeClient,
		InformerFactory:      informerFactory,
		DebuggingSnapshotter: debuggingSnapshotter,
		PredicateChecker:     predicatechecker.NewSchedulerBasedPredicateChecker(fwHandle),
		DeleteOptions:        deleteOptions,
		DrainabilityRules:    drainabilityRules,
		ScaleUpOrchestrator:  orchestrator.New(),
	}

	opts.Processors = ca_processors.DefaultProcessors(autoscalingOptions)
	opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nodeInfoCacheExpireTime, *forceDaemonSets)
	podListProcessor := podlistprocessor.NewDefaultPodListProcessor(opts.PredicateChecker, scheduling.ScheduleAnywhere)

	var ProvisioningRequestInjector *provreq.ProvisioningRequestPodsInjector
	if autoscalingOptions.ProvisioningRequestEnabled {
		podListProcessor.AddProcessor(provreq.NewProvisioningRequestPodsFilter(provreq.NewDefautlEventManager()))

		restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
		client, err := provreqclient.NewProvisioningRequestClient(restConfig)
		if err != nil {
			return nil, nil, err
		}

		ProvisioningRequestInjector, err = provreq.NewProvisioningRequestPodsInjector(restConfig, opts.ProvisioningRequestInitialBackoffTime, opts.ProvisioningRequestMaxBackoffTime, opts.ProvisioningRequestMaxBackoffCacheSize)
		if err != nil {
			return nil, nil, err
		}
		podListProcessor.AddProcessor(ProvisioningRequestInjector)

		var provisioningRequestPodsInjector *provreq.ProvisioningRequestPodsInjector
		if autoscalingOptions.CheckCapacityBatchProcessing {
			klog.Infof("Batch processing for check capacity requests is enabled. Passing provisioning request injector to check capacity processor.")
			provisioningRequestPodsInjector = ProvisioningRequestInjector
		}

		provreqOrchestrator := provreqorchestrator.New(client, []provreqorchestrator.ProvisioningClass{
			checkcapacity.New(client, provisioningRequestPodsInjector),
			besteffortatomic.New(client),
		})

		scaleUpOrchestrator := provreqorchestrator.NewWrapperOrchestrator(provreqOrchestrator)
		opts.ScaleUpOrchestrator = scaleUpOrchestrator
		provreqProcesor := provreq.NewProvReqProcessor(client, opts.PredicateChecker)
		opts.LoopStartNotifier = loopstart.NewObserversList([]loopstart.Observer{provreqProcesor})

		podListProcessor.AddProcessor(provreqProcesor)
	}

	if *proactiveScaleupEnabled {
		podInjectionBackoffRegistry := podinjectionbackoff.NewFakePodControllerRegistry()

		podInjectionPodListProcessor := podinjection.NewPodInjectionPodListProcessor(podInjectionBackoffRegistry)
		enforceInjectedPodsLimitProcessor := podinjection.NewEnforceInjectedPodsLimitProcessor(*podInjectionLimit)

		podListProcessor = pods.NewCombinedPodListProcessor([]pods.PodListProcessor{podInjectionPodListProcessor, podListProcessor, enforceInjectedPodsLimitProcessor})

		// FakePodsScaleUpStatusProcessor processor needs to be the first processor in ScaleUpStatusProcessor before the default processor
		// As it filters out fake pods from Scale Up status so that we don't emit events.
		opts.Processors.ScaleUpStatusProcessor = status.NewCombinedScaleUpStatusProcessor([]status.ScaleUpStatusProcessor{podinjection.NewFakePodsScaleUpStatusProcessor(podInjectionBackoffRegistry), opts.Processors.ScaleUpStatusProcessor})
	}

	opts.Processors.PodListProcessor = podListProcessor
	sdCandidatesSorting := previouscandidates.NewPreviousCandidates()
	scaleDownCandidatesComparers := []scaledowncandidates.CandidatesComparer{
		emptycandidates.NewEmptySortingProcessor(emptycandidates.NewNodeInfoGetter(opts.ClusterSnapshot), deleteOptions, drainabilityRules),
		sdCandidatesSorting,
	}
	opts.Processors.ScaleDownCandidatesNotifier.Register(sdCandidatesSorting)

	cp := scaledowncandidates.NewCombinedScaleDownCandidatesProcessor()
	cp.Register(scaledowncandidates.NewScaleDownCandidatesSortingProcessor(scaleDownCandidatesComparers))

	if autoscalingOptions.ScaleDownDelayTypeLocal {
		sdp := scaledowncandidates.NewScaleDownCandidatesDelayProcessor()
		cp.Register(sdp)
		opts.Processors.ScaleStateNotifier.Register(sdp)

	}
	opts.Processors.ScaleDownNodeProcessor = cp

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
	autoscaler, err := core.NewAutoscaler(opts, informerFactory)
	if err != nil {
		return nil, nil, err
	}

	// Start informers. This must come after fully constructing the autoscaler because
	// additional informers might have been registered in the factory during NewAutoscaler.
	stop := make(chan struct{})
	informerFactory.Start(stop)

	podObserver := loop.StartPodObserver(context, kube_util.CreateKubeClient(autoscalingOptions.KubeClientOpts))

	// A ProvisioningRequestPodsInjector is used as provisioningRequestProcessingTimesGetter here to obtain the last time a
	// ProvisioningRequest was processed. This is because the ProvisioningRequestPodsInjector in addition to injecting pods
	// also marks the ProvisioningRequest as accepted or failed.
	trigger := loop.NewLoopTrigger(autoscaler, ProvisioningRequestInjector, podObserver, *scanInterval)

	return autoscaler, trigger, nil
}

func run(healthCheck *metrics.HealthCheck, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) {
	metrics.RegisterAll(*emitPerNodeGroupMetrics)
	context, cancel := ctx.WithCancel(ctx.Background())
	defer cancel()

	autoscaler, trigger, err := buildAutoscaler(context, debuggingSnapshotter)
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
	if *frequentLoopsEnabled {
		lastRun := time.Now()
		for {
			trigger.Wait(lastRun)
			lastRun = time.Now()
			loop.RunAutoscalerOnce(autoscaler, healthCheck, lastRun)
		}
	} else {
		for {
			time.Sleep(*scanInterval)
			loop.RunAutoscalerOnce(autoscaler, healthCheck, time.Now())
		}
	}
}

func main() {
	klog.InitFlags(nil)

	featureGate := utilfeature.DefaultMutableFeatureGate
	loggingConfig := logsapi.NewLoggingConfiguration()

	if err := logsapi.AddFeatureGates(featureGate); err != nil {
		klog.Fatalf("Failed to add logging feature flags: %v", err)
	}

	logsapi.AddFlags(loggingConfig, pflag.CommandLine)
	featureGate.AddFlag(pflag.CommandLine)
	kube_flag.InitFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = true
	componentopts.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	logs.InitLogs()
	if err := logsapi.ValidateAndApply(loggingConfig, featureGate); err != nil {
		klog.Fatalf("Failed to validate and apply logging configuration: %v", err)
	}

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

		kubeClient := kube_util.CreateKubeClient(createAutoscalingOptions().KubeClientOpts)

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
		ResourceName:  *leaseResourceName,
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
