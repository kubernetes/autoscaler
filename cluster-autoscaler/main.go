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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	capacityclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/config/flags"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/kubernetes/pkg/features"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	capacitybuffer "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/controller"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	cbprocessor "k8s.io/autoscaler/cluster-autoscaler/processors/capacitybuffer"
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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/client-go/informers"
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
)

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
	// Get AutoscalingOptions from flags.
	autoscalingOptions := flags.AutoscalingOptions()

	kubeClient := kube_util.CreateKubeClient(autoscalingOptions.KubeClientOpts)

	// Informer transform to trim ManagedFields for memory efficiency.
	trim := func(obj interface{}) (interface{}, error) {
		if accessor, err := meta.Accessor(obj); err == nil {
			accessor.SetManagedFields(nil)
		}
		return obj, nil
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, informers.WithTransform(trim))

	fwHandle, err := framework.NewHandle(informerFactory, autoscalingOptions.SchedulerConfig, autoscalingOptions.DynamicResourceAllocationEnabled)
	if err != nil {
		return nil, nil, err
	}
	deleteOptions := options.NewNodeDeleteOptions(autoscalingOptions)
	drainabilityRules := rules.Default(deleteOptions)

	var snapshotStore clustersnapshot.ClusterSnapshotStore = store.NewDeltaSnapshotStore(autoscalingOptions.ClusterSnapshotParallelism)
	opts := core.AutoscalerOptions{
		AutoscalingOptions:   autoscalingOptions,
		FrameworkHandle:      fwHandle,
		ClusterSnapshot:      predicate.NewPredicateSnapshot(snapshotStore, fwHandle, autoscalingOptions.DynamicResourceAllocationEnabled),
		KubeClient:           kubeClient,
		InformerFactory:      informerFactory,
		DebuggingSnapshotter: debuggingSnapshotter,
		DeleteOptions:        deleteOptions,
		DrainabilityRules:    drainabilityRules,
		ScaleUpOrchestrator:  orchestrator.New(),
	}

	opts.Processors = ca_processors.DefaultProcessors(autoscalingOptions)
	opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(&autoscalingOptions.NodeInfoCacheExpireTime, autoscalingOptions.ForceDaemonSets)
	podListProcessor := podlistprocessor.NewDefaultPodListProcessor(scheduling.ScheduleAnywhere)

	var ProvisioningRequestInjector *provreq.ProvisioningRequestPodsInjector
	if autoscalingOptions.ProvisioningRequestEnabled {
		podListProcessor.AddProcessor(provreq.NewProvisioningRequestPodsFilter(provreq.NewDefautlEventManager()))

		restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
		client, err := provreqclient.NewProvisioningRequestClient(restConfig)
		if err != nil {
			return nil, nil, err
		}

		ProvisioningRequestInjector, err = provreq.NewProvisioningRequestPodsInjector(restConfig, opts.ProvisioningRequestInitialBackoffTime, opts.ProvisioningRequestMaxBackoffTime, opts.ProvisioningRequestMaxBackoffCacheSize, opts.CheckCapacityBatchProcessing, opts.CheckCapacityProcessorInstance)
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
		provreqProcesor := provreq.NewProvReqProcessor(client, opts.CheckCapacityProcessorInstance)
		opts.LoopStartNotifier = loopstart.NewObserversList([]loopstart.Observer{provreqProcesor})

		podListProcessor.AddProcessor(provreqProcesor)

		opts.Processors.ScaleUpEnforcer = provreq.NewProvisioningRequestScaleUpEnforcer()
	}

	var capacitybufferClient *capacityclient.CapacityBufferClient
	var capacitybufferClientError error
	if autoscalingOptions.CapacitybufferControllerEnabled {
		restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
		capacitybufferClient, capacitybufferClientError = capacityclient.NewCapacityBufferClientFromConfig(restConfig)
		if capacitybufferClientError == nil && capacitybufferClient != nil {
			nodeBufferController := capacitybuffer.NewDefaultBufferController(capacitybufferClient)
			go nodeBufferController.Run(make(chan struct{}))
		}
	}

	if autoscalingOptions.CapacitybufferPodInjectionEnabled {
		if capacitybufferClient == nil {
			restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
			capacitybufferClient, capacitybufferClientError = capacityclient.NewCapacityBufferClientFromConfig(restConfig)
		}
		if capacitybufferClientError == nil && capacitybufferClient != nil {
			bufferPodInjector := cbprocessor.NewCapacityBufferPodListProcessor(capacitybufferClient, []string{common.ActiveProvisioningStrategy})
			podListProcessor = pods.NewCombinedPodListProcessor([]pods.PodListProcessor{bufferPodInjector, podListProcessor})
			opts.Processors.ScaleUpStatusProcessor = status.NewCombinedScaleUpStatusProcessor([]status.ScaleUpStatusProcessor{cbprocessor.NewFakePodsScaleUpStatusProcessor(), opts.Processors.ScaleUpStatusProcessor})
		}
	}

	if autoscalingOptions.ProactiveScaleupEnabled {
		podInjectionBackoffRegistry := podinjectionbackoff.NewFakePodControllerRegistry()

		podInjectionPodListProcessor := podinjection.NewPodInjectionPodListProcessor(podInjectionBackoffRegistry)
		enforceInjectedPodsLimitProcessor := podinjection.NewEnforceInjectedPodsLimitProcessor(autoscalingOptions.PodInjectionLimit)

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
			opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewAsgTagResourceNodeInfoProvider(&autoscalingOptions.NodeInfoCacheExpireTime, autoscalingOptions.ForceDaemonSets)
		} else if autoscalingOptions.CloudProviderName == cloudprovider.GceProviderName {
			nodeInfoComparatorBuilder = nodegroupset.CreateGceNodeInfoComparator
			opts.Processors.TemplateNodeInfoProvider = nodeinfosprovider.NewAnnotationNodeInfoProvider(&autoscalingOptions.NodeInfoCacheExpireTime, autoscalingOptions.ForceDaemonSets)
		}
		nodeInfoComparator = nodeInfoComparatorBuilder(autoscalingOptions.BalancingExtraIgnoredLabels, autoscalingOptions.NodeGroupSetRatios)
	}

	opts.Processors.NodeGroupSetProcessor = &nodegroupset.BalancingNodeGroupSetProcessor{
		Comparator: nodeInfoComparator,
	}

	// These metrics should be published only once.
	metrics.UpdateCPULimitsCores(autoscalingOptions.MinCoresTotal, autoscalingOptions.MaxCoresTotal)
	metrics.UpdateMemoryLimitsBytes(autoscalingOptions.MinMemoryTotal, autoscalingOptions.MaxMemoryTotal)

	// Initialize metrics.
	metrics.InitMetrics()

	// Create autoscaler.
	autoscaler, err := core.NewAutoscaler(opts, informerFactory)
	if err != nil {
		return nil, nil, err
	}

	// Start informers. This must come after fully constructing the autoscaler because
	// additional informers might have been registered in the factory during NewAutoscaler.
	stop := make(chan struct{})
	informerFactory.Start(stop)

	klog.Info("Initializing resource informers, blocking until caches are synced")
	informersSynced := informerFactory.WaitForCacheSync(stop)
	for _, synced := range informersSynced {
		if !synced {
			return nil, nil, fmt.Errorf("unable to start and sync resource informers")
		}
	}

	podObserver := loop.StartPodObserver(context, kube_util.CreateKubeClient(autoscalingOptions.KubeClientOpts))

	// A ProvisioningRequestPodsInjector is used as provisioningRequestProcessingTimesGetter here to obtain the last time a
	// ProvisioningRequest was processed. This is because the ProvisioningRequestPodsInjector in addition to injecting pods
	// also marks the ProvisioningRequest as accepted or failed.
	trigger := loop.NewLoopTrigger(autoscaler, ProvisioningRequestInjector, podObserver, autoscalingOptions.ScanInterval)

	return autoscaler, trigger, nil
}

func run(healthCheck *metrics.HealthCheck, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) {
	autoscalingOpts := flags.AutoscalingOptions()

	metrics.RegisterAll(autoscalingOpts.EmitPerNodeGroupMetrics)
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
	if autoscalingOpts.FrequentLoopsEnabled {
		// We need to have two timestamps because the scaleUp activity alternates between processing ProvisioningRequests,
		// so we need to pass the older timestamp (previousRun) to trigger.Wait to run immediately if only one of the activities is productive.
		lastRun := time.Now()
		previousRun := time.Now()
		for {
			trigger.Wait(previousRun)
			previousRun, lastRun = lastRun, time.Now()
			loop.RunAutoscalerOnce(autoscaler, healthCheck, lastRun)
		}
	} else {
		for {
			time.Sleep(autoscalingOpts.ScanInterval)
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

	leaderElection := leaderElectionConfiguration()
	// Must be called before kube_flag.InitFlags() to ensure leader election flags are parsed and available.
	componentopts.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	logsapi.AddFlags(loggingConfig, pflag.CommandLine)
	featureGate.AddFlag(pflag.CommandLine)
	kube_flag.InitFlags()

	autoscalingOpts := flags.AutoscalingOptions()

	// If the DRA flag is passed, we need to set the DRA feature gate as well. The selection of scheduler plugins for the default
	// scheduling profile depends on feature gates, and the DRA plugin is only included if the DRA feature gate is enabled. The DRA
	// plugin itself also checks the DRA feature gate and doesn't do anything if it's not enabled.
	if autoscalingOpts.DynamicResourceAllocationEnabled && !featureGate.Enabled(features.DynamicResourceAllocation) {
		if err := featureGate.SetFromMap(map[string]bool{string(features.DynamicResourceAllocation): true}); err != nil {
			klog.Fatalf("couldn't enable the DRA feature gate: %v", err)
		}
	}

	logs.InitLogs()
	if err := logsapi.ValidateAndApply(loggingConfig, featureGate); err != nil {
		klog.Fatalf("Failed to validate and apply logging configuration: %v", err)
	}

	healthCheck := metrics.NewHealthCheck(autoscalingOpts.MaxInactivityTime, autoscalingOpts.MaxFailingTime)

	klog.V(1).Infof("Cluster Autoscaler %s", version.ClusterAutoscalerVersion)

	debuggingSnapshotter := debuggingsnapshot.NewDebuggingSnapshotter(autoscalingOpts.DebuggingSnapshotEnabled)

	go func() {
		pathRecorderMux := mux.NewPathRecorderMux("cluster-autoscaler")
		defaultMetricsHandler := legacyregistry.Handler().ServeHTTP
		pathRecorderMux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
			defaultMetricsHandler(w, req)
		})
		if autoscalingOpts.DebuggingSnapshotEnabled {
			pathRecorderMux.HandleFunc("/snapshotz", debuggingSnapshotter.ResponseHandler)
		}
		pathRecorderMux.HandleFunc("/health-check", healthCheck.ServeHTTP)
		if autoscalingOpts.EnableProfiling {
			routes.Profiling{}.Install(pathRecorderMux)
		}
		err := http.ListenAndServe(autoscalingOpts.Address, pathRecorderMux)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()

	if !leaderElection.LeaderElect {
		run(healthCheck, debuggingSnapshotter)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Unable to get hostname: %v", err)
		}

		kubeClient := kube_util.CreateKubeClient(autoscalingOpts.KubeClientOpts)

		// Validate that the client is ok.
		_, err = kubeClient.CoreV1().Nodes().List(ctx.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("Failed to get nodes from apiserver: %v", err)
		}

		lock, err := resourcelock.New(
			leaderElection.ResourceLock,
			autoscalingOpts.ConfigNamespace,
			leaderElection.ResourceName,
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity:      id,
				EventRecorder: kube_util.CreateEventRecorder(kubeClient, autoscalingOpts.RecordDuplicatedEvents),
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

func leaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:   true,
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
