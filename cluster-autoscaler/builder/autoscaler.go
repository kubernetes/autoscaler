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

package builder

import (
	"context"
	"fmt"
	capacityclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	capacitybuffer "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/controller"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/autoscaler/cluster-autoscaler/core/podlistprocessor"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	cbprocessor "k8s.io/autoscaler/cluster-autoscaler/processors/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/podinjection"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/emptycandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates/previouscandidates"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	provreqorchestrator "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// AutoscalerBuilder is the builder object for creating a Cluster Autoscaler instance.
type AutoscalerBuilder struct {
	options config.AutoscalingOptions

	kubeClient      kubernetes.Interface
	listerRegistry  kube_util.ListerRegistry
	podObserver     *loop.UnschedulablePodObserver
	cloudProvider   cloudprovider.CloudProvider
	informerFactory informers.SharedInformerFactory
}

// New creates a builder with default options.
func New(opts config.AutoscalingOptions) *AutoscalerBuilder {
	return &AutoscalerBuilder{
		options: opts,
	}
}

// WithKubeClient allows injecting a FakeK8s client.
func (b *AutoscalerBuilder) WithKubeClient(client kubernetes.Interface) *AutoscalerBuilder {
	b.kubeClient = client
	return b
}

// WithListerRegistry allows injecting a fake ListerRegistry.
func (b *AutoscalerBuilder) WithListerRegistry(registry kube_util.ListerRegistry) *AutoscalerBuilder {
	b.listerRegistry = registry
	return b
}

// WithPodObserver allows injecting a pod observer.
func (b *AutoscalerBuilder) WithPodObserver(podObserver *loop.UnschedulablePodObserver) *AutoscalerBuilder {
	b.podObserver = podObserver
	return b
}

// WithCloudProvider allows injecting a cloud provider.
func (b *AutoscalerBuilder) WithCloudProvider(cloudProvider cloudprovider.CloudProvider) *AutoscalerBuilder {
	b.cloudProvider = cloudProvider
	return b
}

// WithInformerFactory allows injecting a shared informer factory.
func (b *AutoscalerBuilder) WithInformerFactory(f informers.SharedInformerFactory) *AutoscalerBuilder {
	b.informerFactory = f
	return b
}

// Build constructs the Autoscaler based on the provided configuration.
func (b *AutoscalerBuilder) Build(ctx context.Context, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) (core.Autoscaler, *loop.LoopTrigger, error) {
	// Get AutoscalingOptions from flags.
	autoscalingOptions := b.options

	if b.kubeClient == nil {
		return nil, nil, fmt.Errorf("kubeClient is missing: ensure WithKubeClient() is called")
	}
	if b.informerFactory == nil {
		return nil, nil, fmt.Errorf("informerFactory is missing: ensure WithInformerFactory() is called")
	}

	fwHandle, err := framework.NewHandle(b.informerFactory, autoscalingOptions.SchedulerConfig, autoscalingOptions.DynamicResourceAllocationEnabled, autoscalingOptions.CSINodeAwareSchedulingEnabled)
	if err != nil {
		return nil, nil, err
	}
	deleteOptions := options.NewNodeDeleteOptions(autoscalingOptions)
	drainabilityRules := rules.Default(deleteOptions)

	var snapshotStore clustersnapshot.ClusterSnapshotStore = store.NewDeltaSnapshotStore(autoscalingOptions.ClusterSnapshotParallelism)
	opts := coreoptions.AutoscalerOptions{
		AutoscalingOptions:   autoscalingOptions,
		FrameworkHandle:      fwHandle,
		ClusterSnapshot:      predicate.NewPredicateSnapshot(snapshotStore, fwHandle, autoscalingOptions.DynamicResourceAllocationEnabled, autoscalingOptions.PredicateParallelism, autoscalingOptions.CSINodeAwareSchedulingEnabled),
		KubeClient:           b.kubeClient,
		InformerFactory:      b.informerFactory,
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
			go nodeBufferController.Run(ctx.Done())
		}
	}

	if autoscalingOptions.CapacitybufferPodInjectionEnabled {
		if capacitybufferClient == nil {
			restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
			capacitybufferClient, capacitybufferClientError = capacityclient.NewCapacityBufferClientFromConfig(restConfig)
		}
		if capacitybufferClientError == nil && capacitybufferClient != nil {
			buffersPodsRegistry := cbprocessor.NewDefaultCapacityBuffersFakePodsRegistry()
			bufferPodInjector := cbprocessor.NewCapacityBufferPodListProcessor(
				capacitybufferClient,
				[]string{common.ActiveProvisioningStrategy},
				buffersPodsRegistry, true)
			podListProcessor = pods.NewCombinedPodListProcessor([]pods.PodListProcessor{bufferPodInjector, podListProcessor})
			opts.Processors.ScaleUpStatusProcessor = status.NewCombinedScaleUpStatusProcessor([]status.ScaleUpStatusProcessor{
				cbprocessor.NewFakePodsScaleUpStatusProcessor(buffersPodsRegistry), opts.Processors.ScaleUpStatusProcessor})
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

	// These metrics should be published only once.
	metrics.UpdateCPULimitsCores(autoscalingOptions.MinCoresTotal, autoscalingOptions.MaxCoresTotal)
	metrics.UpdateMemoryLimitsBytes(autoscalingOptions.MinMemoryTotal, autoscalingOptions.MaxMemoryTotal)

	// Initialize metrics.
	metrics.InitMetrics()

	if b.listerRegistry != nil {
		autoscalingKubeClients := ca_context.NewAutoscalingKubeClients(opts.AutoscalingOptions, opts.KubeClient, opts.InformerFactory)
		autoscalingKubeClients.ListerRegistry = b.listerRegistry
		opts.AutoscalingKubeClients = autoscalingKubeClients
	}

	if b.cloudProvider != nil {
		opts.CloudProvider = b.cloudProvider
	}

	// Create autoscaler.
	autoscaler, err := core.NewAutoscaler(opts, b.informerFactory)
	if err != nil {
		return nil, nil, err
	}

	b.informerFactory.Start(ctx.Done())

	klog.Info("Waiting for caches to sync...")
	synced := b.informerFactory.WaitForCacheSync(ctx.Done())
	for _, ok := range synced {
		if !ok {
			return nil, nil, fmt.Errorf("failed to sync informer caches")
		}
	}

	if b.podObserver == nil {
		b.podObserver = loop.StartPodObserver(ctx, b.kubeClient)
	}

	// A ProvisioningRequestPodsInjector is used as provisioningRequestProcessingTimesGetter here to obtain the last time a
	// ProvisioningRequest was processed. This is because the ProvisioningRequestPodsInjector in addition to injecting pods
	// also marks the ProvisioningRequest as accepted or failed.
	trigger := loop.NewLoopTrigger(autoscaler, ProvisioningRequestInjector, b.podObserver, autoscalingOptions.ScanInterval)
	return autoscaler, trigger, nil
}
