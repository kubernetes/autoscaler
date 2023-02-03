/*
Copyright 2020 The Kubernetes Authors.

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

package clusterapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	cpuKey          = "capacity.cluster-autoscaler.kubernetes.io/cpu"
	memoryKey       = "capacity.cluster-autoscaler.kubernetes.io/memory"
	diskCapacityKey = "capacity.cluster-autoscaler.kubernetes.io/ephemeral-disk"
	gpuTypeKey      = "capacity.cluster-autoscaler.kubernetes.io/gpu-type"
	gpuCountKey     = "capacity.cluster-autoscaler.kubernetes.io/gpu-count"
	maxPodsKey      = "capacity.cluster-autoscaler.kubernetes.io/maxPods"
	taintsKey       = "capacity.cluster-autoscaler.kubernetes.io/taints"
	labelsKey       = "capacity.cluster-autoscaler.kubernetes.io/labels"
)

var (
	// clusterNameLabel is the label applied to objects(Machine, MachineSet, MachineDeployment)
	// to identify which cluster they are owned by. Because the label can be
	// affected by the CAPI_GROUP environment variable, it is initialized here.
	clusterNameLabel = getClusterNameLabel()

	// errMissingMinAnnotation is the error returned when a
	// machine set does not have an annotation keyed by
	// nodeGroupMinSizeAnnotationKey.
	errMissingMinAnnotation = errors.New("missing min annotation")

	// errMissingMaxAnnotation is the error returned when a
	// machine set does not have an annotation keyed by
	// nodeGroupMaxSizeAnnotationKey.
	errMissingMaxAnnotation = errors.New("missing max annotation")

	// errInvalidMinAnnotationValue is the error returned when a
	// machine set has a non-integral min annotation value.
	errInvalidMinAnnotation = errors.New("invalid min annotation")

	// errInvalidMaxAnnotationValue is the error returned when a
	// machine set has a non-integral max annotation value.
	errInvalidMaxAnnotation = errors.New("invalid max annotation")

	// machineDeleteAnnotationKey is the annotation used by cluster-api to indicate
	// that a machine should be deleted. Because this key can be affected by the
	// CAPI_GROUP env variable, it is initialized here.
	machineDeleteAnnotationKey = getMachineDeleteAnnotationKey()

	// machineAnnotationKey is the annotation used by the cluster-api on Node objects
	// to specify the name of the related Machine object. Because this can be affected
	// by the CAPI_GROUP env variable, it is initialized here.
	machineAnnotationKey = getMachineAnnotationKey()

	// nodeGroupMinSizeAnnotationKey and nodeGroupMaxSizeAnnotationKey are the keys
	// used in MachineSet and MachineDeployment annotations to specify the limits
	// for the node group. Because the keys can be affected by the CAPI_GROUP env
	// variable, they are initialized here.
	nodeGroupMinSizeAnnotationKey = getNodeGroupMinSizeAnnotationKey()
	nodeGroupMaxSizeAnnotationKey = getNodeGroupMaxSizeAnnotationKey()
	zeroQuantity                  = resource.MustParse("0")
)

type normalizedProviderID string

// minSize returns the minimum value encoded in the annotations keyed
// by nodeGroupMinSizeAnnotationKey. Returns errMissingMinAnnotation
// if the annotation doesn't exist or errInvalidMinAnnotation if the
// value is not of type int.
func minSize(annotations map[string]string) (int, error) {
	val, found := annotations[nodeGroupMinSizeAnnotationKey]
	if !found {
		return 0, errMissingMinAnnotation
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.Wrapf(err, "%s", errInvalidMinAnnotation)
	}
	return i, nil
}

// maxSize returns the maximum value encoded in the annotations keyed
// by nodeGroupMaxSizeAnnotationKey. Returns errMissingMaxAnnotation
// if the annotation doesn't exist or errInvalidMaxAnnotation if the
// value is not of type int.
func maxSize(annotations map[string]string) (int, error) {
	val, found := annotations[nodeGroupMaxSizeAnnotationKey]
	if !found {
		return 0, errMissingMaxAnnotation
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.Wrapf(err, "%s", errInvalidMaxAnnotation)
	}
	return i, nil
}

func parseScalingBounds(annotations map[string]string) (int, int, error) {
	minSize, err := minSize(annotations)
	if err != nil && err != errMissingMinAnnotation {
		return 0, 0, err
	}

	if minSize < 0 {
		return 0, 0, errInvalidMinAnnotation
	}

	maxSize, err := maxSize(annotations)
	if err != nil && err != errMissingMaxAnnotation {
		return 0, 0, err
	}

	if maxSize < 0 {
		return 0, 0, errInvalidMaxAnnotation
	}

	if maxSize < minSize {
		return 0, 0, errInvalidMaxAnnotation
	}

	return minSize, maxSize, nil
}

func getOwnerForKind(u *unstructured.Unstructured, kind string) *metav1.OwnerReference {
	for _, ref := range u.GetOwnerReferences() {
		if ref.Kind == kind && ref.Name != "" {
			return ref.DeepCopy()
		}
	}

	return nil
}

func machineOwnerRef(machine *unstructured.Unstructured) *metav1.OwnerReference {
	return getOwnerForKind(machine, machineSetKind)
}

func machineSetOwnerRef(machineSet *unstructured.Unstructured) *metav1.OwnerReference {
	return getOwnerForKind(machineSet, machineDeploymentKind)
}

func machineSetHasMachineDeploymentOwnerRef(machineSet *unstructured.Unstructured) bool {
	return machineSetOwnerRef(machineSet) != nil
}

// normalizedProviderString splits s on '/' returning everything after
// the last '/'.
func normalizedProviderString(s string) normalizedProviderID {
	split := strings.Split(s, "/")
	return normalizedProviderID(split[len(split)-1])
}

func scaleFromZeroAnnotationsEnabled(annotations map[string]string) bool {
	cpu := annotations[cpuKey]
	mem := annotations[memoryKey]

	if cpu != "" && mem != "" {
		return true
	}
	return false
}

func parseKey(annotations map[string]string, key string) (resource.Quantity, error) {
	if val, exists := annotations[key]; exists && val != "" {
		return resource.ParseQuantity(val)
	}
	return zeroQuantity.DeepCopy(), nil
}

func parseIntKey(annotations map[string]string, key string) (resource.Quantity, error) {
	if val, exists := annotations[key]; exists && val != "" {
		valInt, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return zeroQuantity.DeepCopy(), fmt.Errorf("value %q from annotation %q expected to be an integer: %v", val, key, err)
		}
		return *resource.NewQuantity(valInt, resource.DecimalSI), nil
	}
	return zeroQuantity.DeepCopy(), nil
}

func parseCPUCapacity(annotations map[string]string) (resource.Quantity, error) {
	return parseKey(annotations, cpuKey)
}

func parseMemoryCapacity(annotations map[string]string) (resource.Quantity, error) {
	return parseKey(annotations, memoryKey)
}

func parseEphemeralDiskCapacity(annotations map[string]string) (resource.Quantity, error) {
	return parseKey(annotations, diskCapacityKey)
}

func parseGPUCount(annotations map[string]string) (resource.Quantity, error) {
	return parseIntKey(annotations, gpuCountKey)
}

// The GPU type is not currently considered by the autoscaler when planning
// expansion, but most likely will be in the future. This method is being added
// in expectation of that arrival.
// see https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/utils/gpu/gpu.go
func parseGPUType(annotations map[string]string) string {
	if val, found := annotations[gpuTypeKey]; found {
		return val
	}
	return ""
}

func parseMaxPodsCapacity(annotations map[string]string) (resource.Quantity, error) {
	return parseIntKey(annotations, maxPodsKey)
}

func clusterNameFromResource(r *unstructured.Unstructured) string {
	// Use Spec.ClusterName if defined (only available on v1alpha3+ types)
	clusterName, found, err := unstructured.NestedString(r.Object, "spec", "clusterName")
	if err != nil {
		return ""
	}

	if found {
		return clusterName
	}

	// Fallback to value of clusterNameLabel
	if clusterName, ok := r.GetLabels()[clusterNameLabel]; ok {
		return clusterName
	}

	return ""
}

// getNodeGroupMinSizeAnnotationKey returns the key that is used for the
// node group minimum size annotation. This function is needed because the user can
// change the default group name by using the CAPI_GROUP environment variable.
func getNodeGroupMinSizeAnnotationKey() string {
	key := fmt.Sprintf("%s/cluster-api-autoscaler-node-group-min-size", getCAPIGroup())
	return key
}

// getNodeGroupMaxSizeAnnotationKey returns the key that is used for the
// node group maximum size annotation. This function is needed because the user can
// change the default group name by using the CAPI_GROUP environment variable.
func getNodeGroupMaxSizeAnnotationKey() string {
	key := fmt.Sprintf("%s/cluster-api-autoscaler-node-group-max-size", getCAPIGroup())
	return key
}

// getMachineDeleteAnnotationKey returns the key that is used by cluster-api for marking
// machines to be deleted. This function is needed because the user can change the default
// group name by using the CAPI_GROUP environment variable.
func getMachineDeleteAnnotationKey() string {
	key := fmt.Sprintf("%s/delete-machine", getCAPIGroup())
	return key
}

// getMachineAnnotationKey returns the key that is used by cluster-api for annotating
// nodes with their related machine objects. This function is needed because the user can change
// the default group name by using the CAPI_GROUP environment variable.
func getMachineAnnotationKey() string {
	key := fmt.Sprintf("%s/machine", getCAPIGroup())
	return key
}

// getClusterNameLabel returns the key that is used by cluster-api for labeling
// which cluster an object belongs to. This function is needed because the user can change
// the default group name by using the CAPI_GROUP environment variable.
func getClusterNameLabel() string {
	key := fmt.Sprintf("%s/cluster-name", getCAPIGroup())
	return key
}
