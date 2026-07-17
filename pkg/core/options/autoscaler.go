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

package options

import (
	"sigs.k8s.io/cluster-autoscaler/pkg/capacitybuffer/fakepods"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/clusterstate/scaleupfailures"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/pdb"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaleup"
	"sigs.k8s.io/cluster-autoscaler/pkg/debuggingsnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/estimator"
	"sigs.k8s.io/cluster-autoscaler/pkg/expander"
	"sigs.k8s.io/cluster-autoscaler/pkg/observers/loopstart"
	ca_processors "sigs.k8s.io/cluster-autoscaler/pkg/processors"
	"sigs.k8s.io/cluster-autoscaler/pkg/resourcequotas"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	csinodeprovider "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/provider"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules"
	draprovider "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/provider"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/options"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/backoff"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AutoscalerOptions is the whole set of options for configuring an autoscaler
type AutoscalerOptions struct {
	config.AutoscalingOptions
	KubeClient                 kube_client.Interface
	InformerFactory            informers.SharedInformerFactory
	AutoscalingKubeClients     *context.AutoscalingKubeClients
	CloudProvider              cloudprovider.CloudProvider
	FrameworkHandle            *framework.Handle
	ClusterSnapshot            clustersnapshot.ClusterSnapshot
	ExpanderStrategy           expander.Strategy
	EstimatorBuilder           estimator.EstimatorBuilder
	Processors                 *ca_processors.AutoscalingProcessors
	LoopStartObservers         []loopstart.Observer
	Backoff                    backoff.Backoff
	ScaleUpFailuresRegistry    *scaleupfailures.Registry
	DebuggingSnapshotter       debuggingsnapshot.DebuggingSnapshotter
	RemainingPdbTracker        pdb.RemainingPdbTracker
	ScaleUpOrchestrator        scaleup.Orchestrator
	DeleteOptions              options.NodeDeleteOptions
	DrainabilityRules          rules.Rules
	DraProvider                *draprovider.Provider
	QuotasTrackerOptions       resourcequotas.TrackerOptions
	MinQuotasTrackerOptions    resourcequotas.TrackerOptions
	CSIProvider                *csinodeprovider.Provider
	KubeClientNew              client.Client
	KubeCache                  cache.Cache
	CapacityBufferPodsRegistry *fakepods.Registry
}
