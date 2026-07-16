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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/karpenter/pkg/operator/logging"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	volumeutil "sigs.k8s.io/karpenter/pkg/utils/volume"
)

// UnsupportedProvisioners is a map of volume plugins that are not supported. When a pod requests storage using a PersistentVolumeClaim (PVC)
// that uses a StorageClass with any of these unsupported provisioners, Karpenter will skip scheduling that pod.
var UnsupportedProvisioners = sets.New[string]()

// UnsupportedTopologyKeys is a set of topology keys that are not supported. When a StorageClass has AllowedTopologies
// containing any of these keys, Karpenter will skip scheduling pods that reference a PVC with that StorageClass, since nodes
// created by Karpenter will never satisfy the topology requirement.
var UnsupportedTopologyKeys = sets.New[string]()

func NewVolumeTopology(kubeClient client.Client) *VolumeTopology {
	return &VolumeTopology{kubeClient: kubeClient}
}

type VolumeTopology struct {
	kubeClient client.Client
}

// GetRequirements returns the volume topology requirements for the pod as a list of alternatives.
// Each alternative is a scheduling.Requirements representing one valid combination of volume topology constraints.
// When a volume has multiple allowed topology terms (OR'd NodeSelectorTerms or AllowedTopologies),
// each term becomes a separate alternative. For pods with multiple volumes, the cross product of all
// per-volume alternatives is computed.
//
// These requirements should be:
//   - Added to nodeRequirements (for NodeClaim topology filtering)
//   - NOT added to pod's NodeAffinity (to preserve correct TSC counting)
func (v *VolumeTopology) GetRequirements(ctx context.Context, pod *v1.Pod) ([]scheduling.Requirements, error) {
	// Start with a single empty alternative (matches everything)
	alternatives := []scheduling.Requirements{nil}

	for _, vol := range pod.Spec.Volumes {
		volAlts, err := v.getRequirements(ctx, pod, vol)
		if err != nil {
			return nil, err
		}
		if len(volAlts) == 0 {
			continue
		}

		alternatives = mergeVolumeRequirementAlternatives(alternatives, volAlts)
	}

	// If we still have just the initial empty alternative, there are no volume requirements
	if len(alternatives) == 1 && alternatives[0] == nil {
		return nil, nil
	}

	log.FromContext(ctx).
		WithValues("Pod", klog.KObj(pod), "alternatives", len(alternatives)).
		V(1).Info("getting requirements from pod volumes")
	return alternatives, nil
}

func mergeVolumeRequirementAlternatives(alternatives, volAlts []scheduling.Requirements) []scheduling.Requirements {
	mergedAlternatives := mergeCompatibleVolumeRequirementAlternatives(alternatives, volAlts)
	if len(mergedAlternatives) != 0 {
		return mergedAlternatives
	}

	// Prefer only compatible cross-product branches, but preserve the old merged result when every
	// branch is incompatible. Treating the all-pruned case as unschedulable requires separate
	// provisioning metrics and scheduling-decision handling.
	return mergeAllVolumeRequirementAlternatives(alternatives, volAlts)
}

func mergeCompatibleVolumeRequirementAlternatives(alternatives, volAlts []scheduling.Requirements) []scheduling.Requirements {
	var mergedAlternatives []scheduling.Requirements
	for _, existing := range alternatives {
		for _, volReq := range volAlts {
			if !volumeRequirementsCompatible(existing, volReq) {
				continue
			}
			mergedAlternatives = append(mergedAlternatives, mergeVolumeRequirements(existing, volReq))
		}
	}
	return mergedAlternatives
}

func mergeAllVolumeRequirementAlternatives(alternatives, volAlts []scheduling.Requirements) []scheduling.Requirements {
	var mergedAlternatives []scheduling.Requirements
	for _, existing := range alternatives {
		for _, volReq := range volAlts {
			mergedAlternatives = append(mergedAlternatives, mergeVolumeRequirements(existing, volReq))
		}
	}
	return mergedAlternatives
}

func volumeRequirementsCompatible(existing, volReq scheduling.Requirements) bool {
	if existing == nil || volReq == nil {
		return true
	}
	return existing.Intersects(volReq) == nil
}

func mergeVolumeRequirements(existing, volReq scheduling.Requirements) scheduling.Requirements {
	merged := scheduling.NewRequirements()
	if existing != nil {
		merged.Add(existing.Values()...)
	}
	merged.Add(volReq.Values()...)
	return merged
}

func (v *VolumeTopology) getRequirements(ctx context.Context, pod *v1.Pod, volume v1.Volume) ([]scheduling.Requirements, error) {
	pvc, err := volumeutil.GetPersistentVolumeClaim(ctx, v.kubeClient, pod, volume)
	if err != nil {
		return nil, fmt.Errorf("discovering persistent volume claim, %w", err)
	}
	// Not all volume types have PVCs, e.g. emptyDir, hostPath, etc.
	if pvc == nil {
		return nil, nil
	}

	// Persistent Volume Requirements
	if pvc.Spec.VolumeName != "" {
		requirements, err := v.getPersistentVolumeRequirements(ctx, pod, pvc.Spec.VolumeName)
		if err != nil {
			return nil, fmt.Errorf("getting existing requirements, %w", err)
		}
		return requirements, nil
	}
	// Storage Class Requirements
	if sc := lo.FromPtr(pvc.Spec.StorageClassName); sc != "" {
		requirements, err := v.getStorageClassRequirements(ctx, sc)
		if err != nil {
			return nil, err
		}
		return requirements, nil
	}
	return nil, nil
}

func (v *VolumeTopology) getStorageClassRequirements(ctx context.Context, storageClassName string) ([]scheduling.Requirements, error) {
	storageClass := &storagev1.StorageClass{}
	if err := v.kubeClient.Get(ctx, types.NamespacedName{Name: storageClassName}, storageClass); err != nil {
		return nil, serrors.Wrap(fmt.Errorf("getting storage class, %w", err), "StorageClass", klog.KRef("", storageClassName))
	}
	var alternatives []scheduling.Requirements
	// Each TopologySelectorTerm is OR'd — each becomes a separate alternative
	for _, topology := range storageClass.AllowedTopologies {
		var requirements []v1.NodeSelectorRequirement
		for _, requirement := range topology.MatchLabelExpressions {
			requirements = append(requirements, v1.NodeSelectorRequirement{Key: requirement.Key, Operator: v1.NodeSelectorOpIn, Values: requirement.Values})
		}
		if len(requirements) > 0 {
			alternatives = append(alternatives, scheduling.NewNodeSelectorRequirements(requirements...))
		}
	}
	return alternatives, nil
}

func (v *VolumeTopology) getPersistentVolumeRequirements(ctx context.Context, pod *v1.Pod, volumeName string) ([]scheduling.Requirements, error) {
	pv := &v1.PersistentVolume{}
	if err := v.kubeClient.Get(ctx, types.NamespacedName{Name: volumeName, Namespace: pod.Namespace}, pv); err != nil {
		return nil, serrors.Wrap(fmt.Errorf("getting persistent volume, %w", err), "PersistentVolume", klog.KRef("", volumeName))
	}
	if pv.Spec.NodeAffinity == nil || pv.Spec.NodeAffinity.Required == nil {
		return nil, nil
	}

	var alternatives []scheduling.Requirements
	// Each NodeSelectorTerm is OR'd — each becomes a separate alternative
	for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		requirements := term.MatchExpressions
		// If we are using a Local volume or a HostPath volume, then we should ignore the Hostname affinity
		// on it because re-scheduling this pod to a new node means not using the same Hostname affinity that we currently have
		if pv.Spec.Local != nil || pv.Spec.HostPath != nil {
			requirements = lo.Reject(requirements, func(req v1.NodeSelectorRequirement, _ int) bool {
				return req.Key == v1.LabelHostname
			})
			if len(term.MatchExpressions) > 0 && len(requirements) == 0 {
				// Preserve hostname-only terms as unconstrained alternatives, since hostname affinity
				// is intentionally ignored for Local and HostPath volumes.
				alternatives = append(alternatives, scheduling.NewRequirements())
				continue
			}
		}
		if len(requirements) > 0 {
			alternatives = append(alternatives, scheduling.NewNodeSelectorRequirements(requirements...))
		}
	}
	return alternatives, nil
}

// ValidatePersistentVolumeClaims returns an error if the pod doesn't appear to be valid with respect to
// PVCs (e.g. the PVC is not found or references an unknown storage class).
// nolint:gocyclo
func (v *VolumeTopology) ValidatePersistentVolumeClaims(ctx context.Context, pod *v1.Pod) error {
	for _, vol := range pod.Spec.Volumes {
		pvc, err := volumeutil.GetPersistentVolumeClaim(ctx, v.kubeClient, pod, vol)
		if err != nil {
			return err
		}
		// Not all volume types have PVCs, e.g. emptyDir, hostPath, etc.
		if pvc == nil {
			continue
		}
		// Handle cases specifically rejected by kube-scheduler
		// https://github.com/kubernetes/kubernetes/blob/56f6358c11b78e8e3d39e8cd8ff016ff7c70c56b/pkg/scheduler/framework/plugins/volumebinding/volume_binding.go#L333
		if !pvc.DeletionTimestamp.IsZero() {
			return serrors.Wrap(fmt.Errorf("persistentvolumeclaim is being deleted"), "PersistentVolumeClaim", klog.KObj(pvc))
		}
		if pvc.Status.Phase == v1.ClaimLost {
			return serrors.Wrap(fmt.Errorf("persistentvolumeclaim bound to non-existent persistentvolume"), "PersistentVolumeClaim", klog.KObj(pvc), "PersistentVolume", klog.KRef("", pvc.Spec.VolumeName))
		}
		storageClassName := lo.FromPtr(pvc.Spec.StorageClassName)
		if pvc.Spec.VolumeName != "" {
			if err = v.validateVolume(ctx, pvc.Spec.VolumeName); err != nil {
				return serrors.Wrap(fmt.Errorf("failed to validate pvc, %w", err), "PersistentVolumeClaim", klog.KObj(pvc), "PersistentVolume", klog.KRef("", pvc.Spec.VolumeName), "StorageClass", klog.KRef("", storageClassName))
			}
			// kube-scheduler treats PVCs that have a volumeName as Immediate volumes
			// Any PVC that does not contain the "pv.kubernetes.io/bind-completed" annotation is not considered bound
			// https://github.com/kubernetes/kubernetes/blob/ecf2c52f756461cfb7ffd5469975ecd635e5feeb/pkg/scheduler/framework/plugins/volumebinding/binder.go#L770
			if _, ok := pvc.Annotations[volume.AnnBindCompleted]; !ok {
				return serrors.Wrap(fmt.Errorf("pvc is considered unbound because it does not contain annotation"), "annotation", volume.AnnBindCompleted, "PersistentVolumeClaim", klog.KRef("", pvc.Name), "PersistentVolume", klog.KRef("", pvc.Spec.VolumeName))
			}
		} else {
			// PVC is unbound, we can't schedule unless the pod defines a valid storage class
			if storageClassName == "" {
				return serrors.Wrap(fmt.Errorf("unbound pvc must define a storage class"), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName))
			}
			storageClass := &storagev1.StorageClass{}
			if err = v.kubeClient.Get(ctx, types.NamespacedName{Name: storageClassName}, storageClass); err != nil {
				return serrors.Wrap(fmt.Errorf("failed to validate pvc, failed to get storage class, %w", err), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName))
			}
			// Ignore pods than have unbound pvc for volumeBindingMode immediate
			if lo.FromPtr(storageClass.VolumeBindingMode) == storagev1.VolumeBindingImmediate {
				return serrors.Wrap(fmt.Errorf("failed to validate pvc, pvc with immediate volume binding mode must be bound"), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName))
			}
			// Reject pods whose StorageClass has AllowedTopologies with unsupported topology keys,
			// since Karpenter-created nodes will never have matching labels for these keys.
			for _, term := range storageClass.AllowedTopologies {
				for _, expr := range term.MatchLabelExpressions {
					if UnsupportedTopologyKeys.Has(expr.Key) {
						return serrors.Wrap(fmt.Errorf("failed to validate pvc, storage class uses unsupported topology key"), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName), "TopologyKey", expr.Key)
					}
				}
			}
		}
		// Finally, validate that the driver is in the set of supported drivers
		driver, err := scheduling.ResolveDriver(log.IntoContext(ctx, logging.NopLogger), v.kubeClient, pod, vol.Name, pvc, lo.FromPtr(pvc.Spec.StorageClassName))
		if err != nil {
			return serrors.Wrap(fmt.Errorf("failed to validate pvc, %w", err), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName))
		}
		if UnsupportedProvisioners.Has(driver) {
			return serrors.Wrap(fmt.Errorf("failed to validate pvc, provisioner is not supported"), "PersistentVolumeClaim", klog.KObj(pvc), "StorageClass", klog.KRef("", storageClassName), "Provisioner", driver)
		}
	}
	return nil
}

func (v *VolumeTopology) validateVolume(ctx context.Context, volumeName string) error {
	// we have a volume name, so ensure that it exists
	if volumeName != "" {
		pv := &v1.PersistentVolume{}
		if err := v.kubeClient.Get(ctx, types.NamespacedName{Name: volumeName}, pv); err != nil {
			return err
		}
		if !pv.DeletionTimestamp.IsZero() {
			return fmt.Errorf("persistentvolume is being deleted")
		}
	}
	return nil
}
