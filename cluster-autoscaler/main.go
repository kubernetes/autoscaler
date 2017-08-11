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
	"flag"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_leaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/kubernetes/pkg/apis/componentconfig"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
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

var (
	nodeGroupsFlag          MultiStringFlag
	address                 = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubernetes              = flag.String("kubernetes", "", "Kubernetes master location. Leave blank for default")
	cloudConfig             = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	configMapName           = flag.String("configmap", "", "The name of the ConfigMap containing settings used for dynamic reconfiguration. Empty string for no ConfigMap.")
	namespace               = flag.String("namespace", "kube-system", "Namespace in which cluster-autoscaler run. If a --configmap flag is also provided, ensure that the configmap exists in this namespace before CA runs.")
	nodeGroupAutoDiscovery  = flag.String("node-group-auto-discovery", "", "One or more definition(s) of node group auto-discovery. A definition is expressed `<name of discoverer per cloud provider>:[<key>[=<value>]]`. Only the `aws` cloud provider is currently supported. The only valid discoverer for it is `asg` and the valid key is `tag`. For example, specifying `--cloud-provider aws` and `--node-group-auto-discovery asg:tag=cluster-autoscaler/auto-discovery/enabled,kubernetes.io/cluster/<YOUR CLUSTER NAME>` results in ASGs tagged with `cluster-autoscaler/auto-discovery/enabled` and `kubernetes.io/cluster/<YOUR CLUSTER NAME>` to be considered as target node groups")
	verifyUnschedulablePods = flag.Bool("verify-unschedulable-pods", true,
		"If enabled CA will ensure that each pod marked by Scheduler as unschedulable actually can't be scheduled on any node."+
			"This prevents from adding unnecessary nodes in situation when CA and Scheduler have different configuration.")
	scaleDownEnabled = flag.Bool("scale-down-enabled", true, "Should CA scale down the cluster")
	scaleDownDelay   = flag.Duration("scale-down-delay", 10*time.Minute,
		"Duration from the last scale up to the time when CA starts to check scale down options")
	scaleDownUnneededTime = flag.Duration("scale-down-unneeded-time", 10*time.Minute,
		"How long a node should be unneeded before it is eligible for scale down")
	scaleDownUnreadyTime = flag.Duration("scale-down-unready-time", 20*time.Minute,
		"How long an unready node should be unneeded before it is eligible for scale down")
	scaleDownUtilizationThreshold = flag.Float64("scale-down-utilization-threshold", 0.5,
		"Node utilization level, defined as sum of requested resources divided by capacity, below which a node can be considered for scale down")
	scaleDownTrialInterval = flag.Duration("scale-down-trial-interval", 1*time.Minute,
		"How often scale down possiblity is check")
	scanInterval                = flag.Duration("scan-interval", 10*time.Second, "How often cluster is reevaluated for scale up or down")
	maxNodesTotal               = flag.Int("max-nodes-total", 0, "Maximum number of nodes in all node groups. Cluster autoscaler will not grow the cluster beyond this number.")
	cloudProviderFlag           = flag.String("cloud-provider", "gce", "Cloud provider type. Allowed values: gce, aws, azure")
	maxEmptyBulkDeleteFlag      = flag.Int("max-empty-bulk-delete", 10, "Maximum number of empty nodes that can be deleted at the same time.")
	maxGracefulTerminationFlag  = flag.Int("max-graceful-termination-sec", 60, "Maximum number of seconds CA waints for pod termination when trying to scale down a node.")
	maxTotalUnreadyPercentage   = flag.Float64("max-total-unready-percentage", 33, "Maximum percentage of unready nodes after which CA halts operations")
	okTotalUnreadyCount         = flag.Int("ok-total-unready-count", 3, "Number of allowed unready nodes, irrespective of max-total-unready-percentage")
	maxNodeProvisionTime        = flag.Duration("max-node-provision-time", 15*time.Minute, "Maximum time CA waits for node to be provisioned")
	unregisteredNodeRemovalTime = flag.Duration("unregistered-node-removal-time", 15*time.Minute, "Time that CA waits before removing nodes that are not registered in Kubernetes")

	estimatorFlag = flag.String("estimator", estimator.BinpackingEstimatorName,
		"Type of resource estimator to be used in scale up. Available values: ["+strings.Join(estimator.AvailableEstimators, ",")+"]")

	expanderFlag = flag.String("expander", expander.RandomExpanderName,
		"Type of node group expander to be used in scale up. Available values: ["+strings.Join(expander.AvailableExpanders, ",")+"]")

	writeStatusConfigMapFlag     = flag.Bool("write-status-configmap", true, "Should CA write status information to a configmap")
	maxInactivityTimeFlag        = flag.Duration("max-inactivity", 10*time.Minute, "Maximum time from last recorded autoscaler activity before automatic restart")
	maxFailingTimeFlag           = flag.Duration("max-failing-time", 15*time.Minute, "Maximum time from last recorded successful autoscaler run before automatic restart")
	balanceSimilarNodeGroupsFlag = flag.Bool("balance-similar-node-groups", false, "Detect similar node groups and balance the number of nodes between them")
)

func createAutoscalerOptions() core.AutoscalerOptions {
	autoscalingOpts := core.AutoscalingOptions{
		CloudConfig:                   *cloudConfig,
		CloudProviderName:             *cloudProviderFlag,
		NodeGroupAutoDiscovery:        *nodeGroupAutoDiscovery,
		MaxTotalUnreadyPercentage:     *maxTotalUnreadyPercentage,
		OkTotalUnreadyCount:           *okTotalUnreadyCount,
		EstimatorName:                 *estimatorFlag,
		ExpanderName:                  *expanderFlag,
		MaxEmptyBulkDelete:            *maxEmptyBulkDeleteFlag,
		MaxGracefulTerminationSec:     *maxGracefulTerminationFlag,
		MaxNodeProvisionTime:          *maxNodeProvisionTime,
		MaxNodesTotal:                 *maxNodesTotal,
		NodeGroups:                    nodeGroupsFlag,
		UnregisteredNodeRemovalTime:   *unregisteredNodeRemovalTime,
		ScaleDownDelay:                *scaleDownDelay,
		ScaleDownEnabled:              *scaleDownEnabled,
		ScaleDownTrialInterval:        *scaleDownTrialInterval,
		ScaleDownUnneededTime:         *scaleDownUnneededTime,
		ScaleDownUnreadyTime:          *scaleDownUnreadyTime,
		ScaleDownUtilizationThreshold: *scaleDownUtilizationThreshold,
		VerifyUnschedulablePods:       *verifyUnschedulablePods,
		WriteStatusConfigMap:          *writeStatusConfigMapFlag,
		BalanceSimilarNodeGroups:      *balanceSimilarNodeGroupsFlag,
		ConfigNamespace:               *namespace,
	}

	configFetcherOpts := dynamic.ConfigFetcherOptions{
		ConfigMapName: *configMapName,
		Namespace:     *namespace,
	}

	return core.AutoscalerOptions{
		AutoscalingOptions:   autoscalingOpts,
		ConfigFetcherOptions: configFetcherOpts,
	}
}

func createKubeClient() kube_client.Interface {
	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kubernetes url: %v", err)
	}

	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kubernetes client configuration: %v", err)
	}

	return kube_client.NewForConfigOrDie(kubeConfig)
}

func registerSignalHandlers(autoscaler core.Autoscaler) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	glog.Info("Registered cleanup signal handler")

	go func() {
		<-sigs
		glog.Info("Receieved signal, attempting cleanup")
		autoscaler.ExitCleanUp()
		glog.Info("Cleaned up, exiting...")
		glog.Flush()
		os.Exit(0)
	}()
}

func run(healthCheck *metrics.HealthCheck) {
	kubeClient := createKubeClient()
	kubeEventRecorder := kube_util.CreateEventRecorder(kubeClient)
	opts := createAutoscalerOptions()
	predicateCheckerStopChannel := make(chan struct{})
	predicateChecker, err := simulator.NewPredicateChecker(kubeClient, predicateCheckerStopChannel)
	if err != nil {
		glog.Fatalf("Failed to create predicate checker: %v", err)
	}
	listerRegistryStopChannel := make(chan struct{})
	listerRegistry := kube_util.NewListerRegistryWithDefaultListers(kubeClient, listerRegistryStopChannel)
	autoscaler, err := core.NewAutoscaler(opts, predicateChecker, kubeClient, kubeEventRecorder, listerRegistry)
	if err != nil {
		glog.Fatal("Failed to create autoscaler: %v", err)
	}
	autoscaler.CleanUp()
	registerSignalHandlers(autoscaler)
	healthCheck.StartMonitoring()

	for {
		select {
		case <-time.After(*scanInterval):
			{
				loopStart := time.Now()
				metrics.UpdateLastTime("main", loopStart)
				healthCheck.UpdateLastActivity(loopStart)

				err := autoscaler.RunOnce(loopStart)
				if err != nil && err.Type() != errors.TransientError {
					metrics.RegisterError(err)
				} else {
					healthCheck.UpdateLastSuccessfulRun(time.Now())
				}

				metrics.UpdateDuration("main", loopStart)
			}
		}
	}
}

func main() {
	leaderElection := defaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = true

	bindFlags(&leaderElection, pflag.CommandLine)
	flag.Var(&nodeGroupsFlag, "nodes", "sets min,max size and other configuration data for a node group in a format accepted by cloud provider."+
		"Can be used multiple times. Format: <min>:<max>:<other...>")
	kube_flag.InitFlags()

	healthCheck := metrics.NewHealthCheck(*maxInactivityTimeFlag, *maxFailingTimeFlag)

	glog.Infof("Cluster Autoscaler %s", ClusterAutoscalerVersion)

	correctEstimator := false
	for _, availableEstimator := range estimator.AvailableEstimators {
		if *estimatorFlag == availableEstimator {
			correctEstimator = true
		}
	}
	if !correctEstimator {
		glog.Fatalf("Unrecognized estimator: %v", *estimatorFlag)
	}

	go func() {
		http.Handle("/metrics", prometheus.Handler())
		http.Handle("/health-check", healthCheck)
		err := http.ListenAndServe(*address, nil)
		glog.Fatalf("Failed to start metrics: %v", err)
	}()

	if !leaderElection.LeaderElect {
		run(healthCheck)
	} else {
		id, err := os.Hostname()
		if err != nil {
			glog.Fatalf("Unable to get hostname: %v", err)
		}

		kubeClient := createKubeClient()

		// Validate that the client is ok.
		_, err = kubeClient.Core().Nodes().List(metav1.ListOptions{})
		if err != nil {
			glog.Fatalf("Failed to get nodes from apiserver: %v", err)
		}

		kube_leaderelection.RunOrDie(kube_leaderelection.LeaderElectionConfig{
			Lock: &resourcelock.EndpointsLock{
				EndpointsMeta: metav1.ObjectMeta{
					Namespace: *namespace,
					Name:      "cluster-autoscaler",
				},
				Client: kubeClient.Core(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity:      id,
					EventRecorder: kube_util.CreateEventRecorder(kubeClient),
				},
			},
			LeaseDuration: leaderElection.LeaseDuration.Duration,
			RenewDeadline: leaderElection.RenewDeadline.Duration,
			RetryPeriod:   leaderElection.RetryPeriod.Duration,
			Callbacks: kube_leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ <-chan struct{}) {
					// Since we are committing a suicide after losing
					// mastership, we can safely ignore the argument.
					run(healthCheck)
				},
				OnStoppedLeading: func() {
					glog.Fatalf("lost master")
				},
			},
		})
	}
}

func defaultLeaderElectionConfiguration() componentconfig.LeaderElectionConfiguration {
	return componentconfig.LeaderElectionConfiguration{
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.EndpointsResourceLock,
	}
}

func bindFlags(l *componentconfig.LeaderElectionConfiguration, fs *pflag.FlagSet) {
	fs.BoolVar(&l.LeaderElect, "leader-elect", l.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	fs.DurationVar(&l.LeaseDuration.Duration, "leader-elect-lease-duration", l.LeaseDuration.Duration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline.Duration, "leader-elect-renew-deadline", l.RenewDeadline.Duration, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod.Duration, "leader-elect-retry-period", l.RetryPeriod.Duration, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	fs.StringVar(&l.ResourceLock, "leader-elect-resource-lock", l.ResourceLock, ""+
		"The type of resource resource object that is used for locking during"+
		"leader election. Supported options are `endpoints` (default) and `configmap`.")
}

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)
