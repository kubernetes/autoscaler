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
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	autoscalerbuilder "k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/flags"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/client-go/informers"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
	"k8s.io/kubernetes/pkg/features"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cqv1alpha1.AddToScheme(scheme))
	// TODO: add other CRDs
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

func run(healthCheck *metrics.HealthCheck, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) {
	autoscalingOpts := flags.AutoscalingOptions()

	metrics.RegisterAll(autoscalingOpts.EmitPerNodeGroupMetrics)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	restConfig := kube_util.GetKubeConfig(autoscalingOpts.KubeClientOpts)
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			DefaultTransform: cache.TransformStripManagedFields(),
		},
		// TODO: migrate leader election, metrics, healthcheck, pprof servers to Manager
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	if err != nil {
		klog.Fatalf("Failed to create manager: %v", err)
	}

	autoscaler, trigger := mustBuildAutoscaler(ctx, autoscalingOpts, debuggingSnapshotter, mgr)

	// Register signal handlers for graceful shutdown.
	// TODO: replace with ctrl.SetupSignalHandlers() and handle graceful shutdown with context
	registerSignalHandlers(autoscaler)

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	// Start components running in background.
	if err := autoscaler.Start(); err != nil {
		klog.Fatalf("Failed to autoscaler background components: %v", err)
	}

	err = mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		// Autoscale ad infinitum.
		if autoscalingOpts.FrequentLoopsEnabled {
			// We need to have two timestamps because the scaleUp activity alternates between processing ProvisioningRequests,
			// so we need to pass the older timestamp (previousRun) to trigger.Wait to run immediately if only one of the activities is productive.
			lastRun := time.Now()
			previousRun := time.Now()
			for {
				select {
				case <-ctx.Done():
					// TODO: handle graceful shutdown with context
					return nil
				default:
					trigger.Wait(previousRun)
					previousRun, lastRun = lastRun, time.Now()
					loop.RunAutoscalerOnce(autoscaler, healthCheck, lastRun)
				}
			}
		} else {
			for {
				select {
				case <-ctx.Done():
					// TODO: handle graceful shutdown with context
					return nil
				case <-time.After(autoscalingOpts.ScanInterval):
					loop.RunAutoscalerOnce(autoscaler, healthCheck, time.Now())
				}
			}
		}
	}))
	if err != nil {
		klog.Fatalf("Failed to add runnable to manager: %v", err)
	}

	if err := mgr.Start(ctx); err != nil {
		klog.Fatalf("Manager exited with error: %v", err)
	}
}

func mustBuildAutoscaler(ctx context.Context, opts config.AutoscalingOptions, debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter, mgr manager.Manager) (core.Autoscaler, *loop.LoopTrigger) {
	kubeClient := kube_util.CreateKubeClient(opts.KubeClientOpts)

	// Informer transform to trim ManagedFields for memory efficiency.
	trim := func(obj interface{}) (interface{}, error) {
		if accessor, err := meta.Accessor(obj); err == nil {
			accessor.SetManagedFields(nil)
		}
		return obj, nil
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, informers.WithTransform(trim))

	autoscaler, trigger, err := autoscalerbuilder.New(opts).
		WithDebuggingSnapshotter(debuggingSnapshotter).
		WithManager(mgr).
		WithKubeClient(kubeClient).
		WithInformerFactory(informerFactory).
		Build(ctx)

	if err != nil {
		klog.Fatalf("Failed to create autoscaler: %v", err)
	}
	return autoscaler, trigger
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

	// The DRA feature controls whether the DRA scheduler plugin is selected in scheduler framework. The local DRA flag controls whether
	// DRA logic is enabled in Cluster Autoscaler. The 2 values should be in sync - enabling DRA logic in CA without selecting the DRA scheduler
	// plugin doesn't actually do anything, and selecting the DRA scheduler plugin without enabling DRA logic in CA means the plugin is not set up
	// correctly and can panic.
	if autoscalingOpts.DynamicResourceAllocationEnabled != featureGate.Enabled(features.DynamicResourceAllocation) {
		if err := featureGate.SetFromMap(map[string]bool{string(features.DynamicResourceAllocation): autoscalingOpts.DynamicResourceAllocationEnabled}); err != nil {
			klog.Fatalf("couldn't set the DRA feature gate to %v: %v", autoscalingOpts.DynamicResourceAllocationEnabled, err)
		}
	}

	logs.InitLogs()

	opts, err := flags.ComputeLoggingOptions(pflag.CommandLine)
	if err != nil {
		klog.Fatalf("Failed to configure logging: %v", err)
	}

	if err := logsapi.ValidateAndApplyWithOptions(loggingConfig, opts, featureGate); err != nil {
		klog.Fatalf("Failed to validate and apply logging configuration: %v", err)
	}
	ctrl.SetLogger(klog.NewKlogr())

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
		_, err = kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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
				EventRecorder: kube_util.CreateEventRecorder(context.TODO(), kubeClient, autoscalingOpts.RecordDuplicatedEvents),
			},
		)
		if err != nil {
			klog.Fatalf("Unable to create leader election lock: %v", err)
		}

		leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ context.Context) {
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
