/*
Copyright 2021 The Kubernetes Authors.

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

/*
  This provides support for local-data persistent volumes, by two means:
  * Removes volumes using "local-data" from the internal pods copy, for the
    duration of the current autoscaler RunOnce loop. local-data volumes (as
    any volume using a no-provisioner storage class) breaks the VolumeBinding
    predicate used during Scheduler Framework's evaluations.
  * Injects a custom resource request to pods having "local-data" volumes.
    Our autoscaler fork is placing the same resource on NodeInfo templates
    built from ASG/MIG/VMSS when they offer nodes with local-data storage.
    Those virtual NodeInfos are used when the autoscaler evaluates upscale
    candidates. Injecting those requests on pods allows us to upscale only
    nodes having local data, and to know those nodes can host a single pod
    requesting a local-data volume (because allocatable qty = request = 1).

  Caveats:
  * That's obviously not upstreamable
  * With that resource req, none of the existing real nodes can be considered
    by autoscaler as schedulable for pods requesting local-data volumes: the
    "storageclass/local-data" resource is only available on virtual nodes built
    from asg templates (so, during upscale simulations and for upcoming nodes),
    but not at "filter out pods schedulables on existing real nodes" phase.
    Hence the need for an other patch: once those nodes just became ready but
    the pod is not scheduled on them yet (eg. when their local data volume
    isn't built or bound yet): the autoscaler wouldn't consider those fresh
    nodes (now evaluated by using nodeInfos built from real/live nodes)
    suitable for their pendind pods anymore, since now the nodes don't have
    the requested custom resource. Which would lead to spurious re-upscales
    during the "node became Ready -> pod now scheduled to that node" phase.
  * Using that hack forces the use nodeinfos built from asg templates, rather
    than from real world nodes (as op. to upstream behaviour). Which we do
    also for other reasons anyway (scale from zero + balance similar).
*/

package pods

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/common"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
)

type transformLocalData struct {
	pvcLister   v1lister.PersistentVolumeClaimLister
	stopChannel chan struct{}
}

// NewTransformLocalData instantiate a transformLocalData processor
func NewTransformLocalData() *transformLocalData {
	return &transformLocalData{
		stopChannel: make(chan struct{}),
	}
}

// CleanUp shuts down the pv lister
func (p *transformLocalData) CleanUp() {
	close(p.stopChannel)
}

// Process replace volumes to local-data pv by our custom resource
func (p *transformLocalData) Process(ctx *context.AutoscalingContext, pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	if p.pvcLister == nil {
		p.pvcLister = NewPersistentVolumeClaimLister(ctx.ClientSet, p.stopChannel)
	}

	for _, po := range pods {
		var volumes []apiv1.Volume
		for _, vol := range po.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				volumes = append(volumes, vol)
				continue
			}
			pvc, err := p.pvcLister.PersistentVolumeClaims(po.Namespace).Get(vol.PersistentVolumeClaim.ClaimName)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					klog.Warningf("failed to fetch pvc for %s/%s: %v", po.GetNamespace(), po.GetName(), err)
				}
				volumes = append(volumes, vol)
				continue
			}
			if *pvc.Spec.StorageClassName != "local-data" {
				volumes = append(volumes, vol)
				continue
			}

			if len(po.Spec.Containers[0].Resources.Requests) == 0 {
				po.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{}
			}
			if len(po.Spec.Containers[0].Resources.Limits) == 0 {
				po.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{}
			}

			po.Spec.Containers[0].Resources.Requests[common.DatadogLocalDataResource] = common.DatadogLocalDataQuantity.DeepCopy()
			po.Spec.Containers[0].Resources.Limits[common.DatadogLocalDataResource] = common.DatadogLocalDataQuantity.DeepCopy()
		}
		po.Spec.Volumes = volumes
	}

	return pods, nil
}

// NewPersistentVolumeClaimLister builds a persistentvolumeclaim lister.
func NewPersistentVolumeClaimLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.PersistentVolumeClaimLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "persistentvolumeclaims", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.PersistentVolumeClaim{}, time.Hour)
	lister := v1lister.NewPersistentVolumeClaimLister(store)
	go reflector.Run(stopchannel)
	return lister
}
