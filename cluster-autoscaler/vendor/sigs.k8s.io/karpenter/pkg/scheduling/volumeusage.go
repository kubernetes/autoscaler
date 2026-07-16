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

package scheduling

import (
	"context"
	"fmt"

	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	csitranslation "k8s.io/csi-translation-lib"
	"k8s.io/csi-translation-lib/plugins"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	volumeutil "sigs.k8s.io/karpenter/pkg/utils/volume"
)

//go:generate go tool -modfile=../../go.tools.mod controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="."

// translator is a CSI Translator that translates in-tree plugin names to their out-of-tree CSI driver names
var translator = csitranslation.New()

// +k8s:deepcopy-gen=true
type Volumes map[string]sets.Set[string]

func (u Volumes) Add(provisioner string, pvcID string) {
	existing, ok := u[provisioner]
	if !ok {
		existing = sets.New[string]()
		u[provisioner] = existing
	}
	existing.Insert(pvcID)
}

func (u Volumes) Union(vol Volumes) Volumes {
	cp := Volumes{}
	for k, v := range u {
		cp[k] = sets.New(sets.List(v)...)
	}
	for k, v := range vol {
		existing, ok := cp[k]
		if !ok {
			existing = sets.New[string]()
			cp[k] = existing
		}
		existing.Insert(sets.List(v)...)
	}
	return cp
}

func (u Volumes) Insert(volumes Volumes) {
	for k, v := range volumes {
		existing, ok := u[k]
		if !ok {
			existing = sets.New[string]()
			u[k] = existing
		}
		existing.Insert(sets.List(v)...)
	}
}

//nolint:gocyclo
func GetVolumes(ctx context.Context, kubeClient client.Client, pod *v1.Pod) (Volumes, error) {
	podPVCs := Volumes{}
	for _, volume := range pod.Spec.Volumes {
		pvc, err := volumeutil.GetPersistentVolumeClaim(ctx, kubeClient, pod, volume)
		// If the PVC is not found it was manually deleted and its finalizer removed. We should ignore this volume when
		// computing limits, otherwise Karpenter may never be able to update its cluster state.
		if err != nil {
			if errors.IsNotFound(err) {
				log.FromContext(ctx).WithValues("Pod", klog.KObj(pod), "volume", volume.Name).Error(err, "failed tracking CSI volume limits for volume")
				continue
			}
			return nil, fmt.Errorf("failed updating volume limits, %w", err)
		}
		// Not all volume types have PVCs, e.g. emptyDir, hostPath, etc.
		if pvc == nil {
			continue
		}
		driverName, err := ResolveDriver(ctx, kubeClient, pod, volume.Name, pvc, lo.FromPtr(pvc.Spec.StorageClassName))
		if err != nil {
			return nil, err
		}
		// might be a non-CSI driver, something we don't currently handle
		if driverName != "" {
			podPVCs.Add(driverName, client.ObjectKeyFromObject(pvc).String())
		}
	}
	return podPVCs, nil
}

// ResolveDriver resolves the storage driver name in the following order:
//  1. If the PV associated with the pod volume is using CSI.driver in its spec, then use that name
//  2. If the StorageClass associated with the PV has a Provisioner
func ResolveDriver(ctx context.Context, kubeClient client.Client, pod *v1.Pod, volumeName string, pvc *v1.PersistentVolumeClaim, storageClassName string) (string, error) {
	// We can track the volume usage by the CSI Driver name which is pulled from the storage class for dynamic
	// volumes, or if it's bound/static we can pull the volume name
	if pvc.Spec.VolumeName != "" {
		driverName, err := driverFromVolume(ctx, kubeClient, pvc.Spec.VolumeName)
		if err != nil {
			return "", err
		}
		if driverName != "" {
			return driverName, nil
		}
		// The PVC is bound, but not to a volume managed by a CSI driver or a known in-tree equivalent. This PVC can be ignored for the purposes of volume limit tracking.
		return "", nil
	}

	// This can occur in two scenarios:
	//  1. The storage class was explicitly set to "" to disable dynamic provisioning
	//  2. The storage class was not set but the cluster doesn't have a default storage class
	// In either of these cases, a PV must have been previously bound to the PVC and has since been removed. We can
	// ignore this PVC while computing limits and continue.
	if storageClassName == "" {
		log.FromContext(ctx).WithValues("volume", volumeName, "Pod", klog.KObj(pod), "PersistentVolumeClaim", klog.KObj(pvc)).V(1).Info("failed tracking CSI volume limits for volume with unbound PVC, no storage class specified")
		return "", nil
	}

	driverName, err := driverFromSC(ctx, kubeClient, storageClassName)
	if err != nil {
		// There are two scenarios where a StorageClass may be defined but not found:
		//  1. The StorageClass was manually deleted and the finalizer removed
		//  2. The StorageClass never existed and was used to bind the PVC to an existing PV, but that PV was removed
		// In either of these cases, we should ignore the PVC while computing limits and continue.
		if errors.IsNotFound(err) {
			log.FromContext(ctx).WithValues("volume", volumeName, "Pod", klog.KObj(pod), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName)).V(1).Info("failed tracking CSI volume limits for volume with unbound PVC", "error", err)
			return "", nil
		}
		return "", err
	}
	return driverName, nil
}

// driverFromSC resolves the storage driver name by getting the Provisioner name from the StorageClass
func driverFromSC(ctx context.Context, kubeClient client.Client, storageClassName string) (string, error) {
	var sc storagev1.StorageClass
	if err := kubeClient.Get(ctx, client.ObjectKey{Name: storageClassName}, &sc); err != nil {
		return "", err
	}
	// Check if the provisioner name is an in-tree plugin name
	if csiName, err := translator.GetCSINameFromInTreeName(sc.Provisioner); err == nil {
		return csiName, nil
	}
	return sc.Provisioner, nil
}

// driverFromVolume resolves the storage driver name by getting the CSI spec from inside the PersistentVolume
func driverFromVolume(ctx context.Context, kubeClient client.Client, volumeName string) (string, error) {
	var pv v1.PersistentVolume
	if err := kubeClient.Get(ctx, client.ObjectKey{Name: volumeName}, &pv); err != nil {
		return "", err
	}
	if pv.Spec.CSI != nil {
		return pv.Spec.CSI.Driver, nil
	} else if pv.Spec.AWSElasticBlockStore != nil {
		return plugins.AWSEBSDriverName, nil
	}
	return "", nil
}

// VolumeUsage tracks volume limits on a per node basis.  The number of volumes that can be mounted varies by instance
// type. We need to be aware and track the mounted volume usage to inform our awareness of which pods can schedule to
// which nodes.
// +k8s:deepcopy-gen=true
type VolumeUsage struct {
	volumes    Volumes
	podVolumes map[types.NamespacedName]Volumes
	limits     map[string]int
}

func NewVolumeUsage() *VolumeUsage {
	return &VolumeUsage{
		volumes:    Volumes{},
		podVolumes: map[types.NamespacedName]Volumes{},
		limits:     map[string]int{},
	}
}

func (v *VolumeUsage) ExceedsLimits(vols Volumes) error {
	for k, volumes := range v.volumes.Union(vols) {
		if limit, hasLimit := v.limits[k]; hasLimit && len(volumes) > limit {
			return serrors.Wrap(fmt.Errorf("would exceed volume limit"), "provisioner", k, "volume-count", len(volumes), "volume-limit", limit)
		}
	}
	return nil
}

func (v *VolumeUsage) AddLimit(storageDriver string, value int) {
	v.limits[storageDriver] = value
}

func (v *VolumeUsage) Add(pod *v1.Pod, volumes Volumes) {
	v.podVolumes[client.ObjectKeyFromObject(pod)] = volumes
	v.volumes = v.volumes.Union(volumes)
}

func (v *VolumeUsage) DeletePod(key types.NamespacedName) {
	delete(v.podVolumes, key)
	// volume names could be duplicated, so we re-create our volumes
	v.volumes = Volumes{}
	for _, c := range v.podVolumes {
		v.volumes.Insert(c)
	}
}
