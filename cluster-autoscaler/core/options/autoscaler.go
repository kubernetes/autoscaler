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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	csinodeprovider "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/provider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	draprovider "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/provider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
)

// AutoscalerOptions is the whole set of options for configuring an autoscaler
type AutoscalerOptions struct {
	config.AutoscalingOptions
	KubeClient             kube_client.Interface
	InformerFactory        informers.SharedInformerFactory
	AutoscalingKubeClients *context.AutoscalingKubeClients
	CloudProvider          cloudprovider.CloudProvider
	FrameworkHandle        *framework.Handle
	ClusterSnapshot        clustersnapshot.ClusterSnapshot
	ExpanderStrategy       expander.Strategy
	EstimatorBuilder       estimator.EstimatorBuilder
	Processors             *ca_processors.AutoscalingProcessors
	LoopStartNotifier      *loopstart.ObserversList
	Backoff                backoff.Backoff
	DebuggingSnapshotter   debuggingsnapshot.DebuggingSnapshotter
	RemainingPdbTracker    pdb.RemainingPdbTracker
	ScaleUpOrchestrator    scaleup.Orchestrator
	DeleteOptions          options.NodeDeleteOptions
	DrainabilityRules      rules.Rules
	DraProvider            *draprovider.Provider
	QuotasTrackerOptions   resourcequotas.TrackerOptions
	CSIProvider            *csinodeprovider.Provider
}
