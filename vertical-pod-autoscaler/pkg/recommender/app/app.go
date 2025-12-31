/*
Copyright The Kubernetes Authors.

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

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/klog/v2"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	input_metrics "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	metrics_resources "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/resources"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// aggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
	aggregateContainerStateGCInterval               = 1 * time.Hour
	scaleCacheEntryLifetime           time.Duration = time.Hour
	scaleCacheEntryFreshnessTime      time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor       float64       = 1.
	scaleCacheLoopPeriod                            = 7 * time.Second
	defaultResyncPeriod               time.Duration = 10 * time.Minute
)

// RecommenderApp represents the recommender application
type RecommenderApp struct {
	config *recommender_config.RecommenderConfig
}

// NewRecommenderApp creates a new RecommenderApp with the given configuration
func NewRecommenderApp(config *recommender_config.RecommenderConfig) (*RecommenderApp, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Load bearer token from file if specified
	if config.PrometheusBearerTokenFile != "" {
		fileContent, err := os.ReadFile(config.PrometheusBearerTokenFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read bearer token file %s: %w", config.PrometheusBearerTokenFile, err)
		}
		config.PrometheusBearerToken = strings.TrimSpace(string(fileContent))
	}

	return &RecommenderApp{
		config: config,
	}, nil
}

// Run starts the recommender with the given context
func (app *RecommenderApp) Run(ctx context.Context, leaderElection componentbaseconfig.LeaderElectionConfiguration) error {
	stopCh := make(chan struct{})
	// Close stopCh when context is canceled to signal all goroutines to stop
	go func() {
		<-ctx.Done()
		close(stopCh)
	}()

	healthCheck := metrics.NewHealthCheck(app.config.MetricsFetcherInterval * 5)
	metrics_recommender.Register()
	metrics_quality.Register()
	metrics_resources.Register()
	server.Initialize(&app.config.CommonFlags.EnableProfiling, healthCheck, &app.config.Address)

	if !leaderElection.LeaderElect {
		return app.run(ctx, healthCheck)
	} else {
		id, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("unable to get hostname: %w", err)
		}

		id = id + "_" + string(uuid.NewUUID())

		config := common.CreateKubeConfigOrDie(app.config.CommonFlags.KubeConfig, float32(app.config.CommonFlags.KubeApiQps), int(app.config.CommonFlags.KubeApiBurst))
		kubeClient := kube_client.NewForConfigOrDie(config)

		lock, err := resourcelock.New(
			leaderElection.ResourceLock,
			leaderElection.ResourceNamespace,
			leaderElection.ResourceName,
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity: id,
			},
		)
		if err != nil {
			return fmt.Errorf("unable to create leader election lock: %w", err)
		}

		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ context.Context) {
					if err := app.run(ctx, healthCheck); err != nil {
						klog.Fatalf("Error running recommender: %v", err)
					}
				},
				OnStoppedLeading: func() {
					klog.Fatal("lost master")
				},
			},
		})
	}

	return nil
}

func (app *RecommenderApp) run(ctx context.Context, healthCheck *metrics.HealthCheck) error {
	// Create a stop channel that will be used to signal shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	config := common.CreateKubeConfigOrDie(app.config.CommonFlags.KubeConfig, float32(app.config.CommonFlags.KubeApiQps), int(app.config.CommonFlags.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(config)
	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(app.config.CommonFlags.VpaObjectNamespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(ctx, kubeClient, app.config.CommonFlags.VpaObjectNamespace, stopCh)

	factory.Start(stopCh)
	informerMap := factory.WaitForCacheSync(stopCh)
	for kind, synced := range informerMap {
		if !synced {
			return fmt.Errorf("could not sync cache for the %s informer", kind.String())
		}
	}

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(
		app.config.MemoryAggregationInterval,
		app.config.MemoryAggregationIntervalCount,
		app.config.MemoryHistogramDecayHalfLife,
		app.config.CpuHistogramDecayHalfLife,
		app.config.OOMBumpUpRatio,
		app.config.OOMMinBumpUp,
	))

	useCheckpoints := app.config.Storage != "prometheus"

	var postProcessors []routines.RecommendationPostProcessor
	if app.config.PostProcessorCPUasInteger {
		postProcessors = append(postProcessors, &routines.IntegerCPUPostProcessor{})
	}

	globalMaxAllowed := app.initGlobalMaxAllowed()
	// CappingPostProcessor, should always come in the last position for post-processing
	postProcessors = append(postProcessors, routines.NewCappingRecommendationProcessor(globalMaxAllowed))

	var source input_metrics.PodMetricsLister
	if app.config.UseExternalMetrics {
		resourceMetrics := map[apiv1.ResourceName]string{}
		if app.config.ExternalCpuMetric != "" {
			resourceMetrics[apiv1.ResourceCPU] = app.config.ExternalCpuMetric
		}
		if app.config.ExternalMemoryMetric != "" {
			resourceMetrics[apiv1.ResourceMemory] = app.config.ExternalMemoryMetric
		}
		externalClientOptions := &input_metrics.ExternalClientOptions{
			ResourceMetrics:    resourceMetrics,
			ContainerNameLabel: app.config.CtrNameLabel,
		}
		klog.V(1).InfoS("Using External Metrics", "options", externalClientOptions)
		source = input_metrics.NewExternalClient(config, clusterState, *externalClientOptions)
	} else {
		klog.V(1).InfoS("Using Metrics Server")
		source = input_metrics.NewPodMetricsesSource(resourceclient.NewForConfigOrDie(config))
	}

	ignoredNamespaces := strings.Split(app.config.CommonFlags.IgnoredVpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, app.config.CommonFlags.VpaObjectNamespace, "default-metrics-client"),
		VpaCheckpointClient: vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		VpaLister:           vpa_api_util.NewVpasLister(vpa_clientset.NewForConfigOrDie(config), stopCh, app.config.CommonFlags.VpaObjectNamespace),
		VpaCheckpointLister: vpa_api_util.NewVpaCheckpointLister(vpa_clientset.NewForConfigOrDie(config), stopCh, app.config.CommonFlags.VpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(config, kubeClient, factory),
		MemorySaveMode:      app.config.MemorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     app.config.RecommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
		VpaObjectNamespace:  app.config.CommonFlags.VpaObjectNamespace,
	}.Make()
	controllerFetcher.Start(ctx, scaleCacheLoopPeriod)

	recommender := routines.RecommenderFactory{
		ClusterState:                 clusterState,
		ClusterStateFeeder:           clusterStateFeeder,
		ControllerFetcher:            controllerFetcher,
		CheckpointWriter:             checkpoint.NewCheckpointWriter(clusterState, vpa_clientset.NewForConfigOrDie(config).AutoscalingV1()),
		VpaClient:                    vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		PodResourceRecommender:       logic.CreatePodResourceRecommender(),
		RecommendationPostProcessors: postProcessors,
		CheckpointsGCInterval:        app.config.CheckpointsGCInterval,
		UseCheckpoints:               useCheckpoints,
		UpdateWorkerCount:            app.config.UpdateWorkerCount,
	}.Make()

	promQueryTimeout, err := time.ParseDuration(app.config.QueryTimeout)
	if err != nil {
		return fmt.Errorf("could not parse --prometheus-query-timeout as a time.Duration: %w", err)
	}

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints(ctx)
	} else {
		historyConfig := history.PrometheusHistoryProviderConfig{
			Address:                app.config.PrometheusAddress,
			Insecure:               app.config.PrometheusInsecure,
			QueryTimeout:           promQueryTimeout,
			HistoryLength:          app.config.HistoryLength,
			HistoryResolution:      app.config.HistoryResolution,
			PodLabelPrefix:         app.config.PodLabelPrefix,
			PodLabelsMetricName:    app.config.PodLabelsMetricName,
			PodNamespaceLabel:      app.config.PodNamespaceLabel,
			PodNameLabel:           app.config.PodNameLabel,
			CtrNamespaceLabel:      app.config.CtrNamespaceLabel,
			CtrPodNameLabel:        app.config.CtrPodNameLabel,
			CtrNameLabel:           app.config.CtrNameLabel,
			CadvisorMetricsJobName: app.config.PrometheusJobName,
			Namespace:              app.config.CommonFlags.VpaObjectNamespace,
			Authentication: history.PrometheusCredentials{
				BearerToken: app.config.PrometheusBearerToken,
				Username:    app.config.Username,
				Password:    app.config.Password,
			},
		}
		provider, err := history.NewPrometheusHistoryProvider(historyConfig)
		if err != nil {
			return fmt.Errorf("could not initialize history provider: %w", err)
		}
		recommender.GetClusterStateFeeder().InitFromHistoryProvider(provider)
	}

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	ticker := time.NewTicker(app.config.MetricsFetcherInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			recommender.RunOnce()
			healthCheck.UpdateLastActivity()
		}
	}
}

func (app *RecommenderApp) initGlobalMaxAllowed() apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	if !app.config.MaxAllowedCPU.IsZero() {
		result[apiv1.ResourceCPU] = app.config.MaxAllowedCPU.Quantity
	}
	if !app.config.MaxAllowedMemory.IsZero() {
		result[apiv1.ResourceMemory] = app.config.MaxAllowedMemory.Quantity
	}
	return result
}

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

// DefaultLeaderElectionConfiguration returns the default leader election configuration
func DefaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.LeasesResourceLock,
		// This was changed from "vpa-recommender" to avoid conflicts with managed VPA deployments.
		ResourceName:      "vpa-recommender-lease",
		ResourceNamespace: metav1.NamespaceSystem,
	}
}
