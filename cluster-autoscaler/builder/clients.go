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

package builder

import (
	"context"
	"fmt"
	"time"

	provreqclientset "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/clientset/versioned"
	provreqinformers "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/client/informers/externalversions"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/besteffortatomic"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	provreqorchestrator "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// buildProvisioningRequest instantiates and configures all components required for
// Provisioning Request support (clients, informers, orchestrators, and processors)
// and wires them into the main Autoscaler options and pod list processors.
func (b *AutoscalerBuilder) buildProvisioningRequest(
	ctx context.Context,
	autoscalingOptions config.AutoscalingOptions,
	opts *coreoptions.AutoscalerOptions,
	podListProcessor *pods.CombinedPodListProcessor,
) (*provreq.ProvisioningRequestPodsInjector, error) {
	podListProcessor.AddProcessor(provreq.NewProvisioningRequestPodsFilter(provreq.NewDefautlEventManager()))

	var prClient provreqclientset.Interface
	if b.prClient != nil {
		prClient = b.prClient
	} else {
		var err error
		restConfig := kube_util.GetKubeConfig(autoscalingOptions.KubeClientOpts)
		prClient, err = provreqclientset.NewForConfig(restConfig)
		if err != nil {
			return nil, err
		}
	}

	prFactory := provreqinformers.NewSharedInformerFactory(prClient, 1*time.Hour)
	provReqLister := prFactory.Autoscaling().V1().ProvisioningRequests().Lister()

	podTemplLister := b.informerFactory.Core().V1().PodTemplates().Lister()

	client := provreqclient.NewProvisioningRequestClient(prClient, provReqLister, podTemplLister)

	prFactory.Start(ctx.Done())
	klog.Info("Waiting for Provisioning Request cache to sync...")
	synced := prFactory.WaitForCacheSync(ctx.Done())
	for _, ok := range synced {
		if !ok {
			return nil, fmt.Errorf("failed to sync Provisioning Request informers")
		}
	}
	klog.V(2).Info("Successful initial Provisioning Request sync")

	injector := provreq.NewProvisioningRequestPodsInjector(client, opts.ProvisioningRequestInitialBackoffTime, opts.ProvisioningRequestMaxBackoffTime, opts.ProvisioningRequestMaxBackoffCacheSize, opts.CheckCapacityBatchProcessing, opts.CheckCapacityProcessorInstance)
	podListProcessor.AddProcessor(injector)

	var provisioningRequestPodsInjector *provreq.ProvisioningRequestPodsInjector
	if autoscalingOptions.CheckCapacityBatchProcessing {
		klog.Infof("Batch processing for check capacity requests is enabled. Passing provisioning request injector to check capacity processor.")
		provisioningRequestPodsInjector = injector
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

	return injector, nil
}
