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
	"os"
	"strconv"
	"strings"
	"sync"

	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	cpuKey                              = "capacity.cluster-autoscaler.kubernetes.io/cpu"
	memoryKey                           = "capacity.cluster-autoscaler.kubernetes.io/memory"
	diskCapacityKey                     = "capacity.cluster-autoscaler.kubernetes.io/ephemeral-disk"
	gpuTypeKey                          = "capacity.cluster-autoscaler.kubernetes.io/gpu-type"
	gpuCountKey                         = "capacity.cluster-autoscaler.kubernetes.io/gpu-count"
	maxPodsKey                          = "capacity.cluster-autoscaler.kubernetes.io/maxPods"
	taintsKey                           = "capacity.cluster-autoscaler.kubernetes.io/taints"
	labelsKey                           = "capacity.cluster-autoscaler.kubernetes.io/labels"
	draDriverKey                        = "capacity.cluster-autoscaler.kubernetes.io/dra-driver"
	machineDeploymentRevisionAnnotation = "machinedeployment.clusters.x-k8s.io/revision"
	machineDeploymentNameLabel          = "cluster.x-k8s.io/deployment-name"
	// UnknownArch is used if the Architecture is Unknown
	UnknownArch SystemArchitecture = ""
	// Amd64 is used if the Architecture is x86_64
	Amd64 SystemArchitecture = "amd64"
	// Arm64 is used if the Architecture is ARM64
	Arm64 SystemArchitecture = "arm64"
	// Ppc64le is used if the Architecture is ppc64le
	Ppc64le SystemArchitecture = "ppc64le"
	// S390x is used if the Architecture is s390x
	S390x SystemArchitecture = "s390x"
	// DefaultArch should be used as a fallback if not passed by the environment via the --scale-up-from-zero-default-arch
	DefaultArch = Amd64
	// scaleUpFromZeroDefaultEnvVar is the name of the env var for the default architecture
	scaleUpFromZeroDefaultArchEnvVar = "CAPI_SCALE_ZERO_DEFAULT_ARCH"
	// GpuDeviceType is used if DRA device is GPU
	GpuDeviceType = "gpu"

	// Cluster API constants, copied from cluster-api/api/core/v1beta1/machine_types.go
	// nodeRoleLabelPrefix is one of the CAPI managed Node label prefixes.
	nodeRoleLabelPrefix = "node-role.kubernetes.io"
	// nodeRestrictionLabelDomain is one of the CAPI managed Node label domains.
	nodeRestrictionLabelDomain = "node-restriction.kubernetes.io"
	// managedNodeLabelDomain is one of the CAPI managed Node label domains.
	managedNodeLabelDomain = "node.cluster.x-k8s.io"
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

	// clusterNameAnnotationKey is the annotation used by cluster-api for annotating nodes
	// with their cluster name.
	clusterNameAnnotationKey = getClusterNameAnnotationKey()

	// clusterNamespaceAnnotationKey is the annotation used by cluster-api for annotating nodes
	// with their cluster namespace.
	clusterNamespaceAnnotationKey = getClusterNamespaceAnnotationKey()

	// nodeGroupMinSizeAnnotationKey and nodeGroupMaxSizeAnnotationKey are the keys
	// used in MachineSet and MachineDeployment annotations to specify the limits
	// for the node group. Because the keys can be affected by the CAPI_GROUP env
	// variable, they are initialized here.
	nodeGroupMinSizeAnnotationKey = getNodeGroupMinSizeAnnotationKey()
	nodeGroupMaxSizeAnnotationKey = getNodeGroupMaxSizeAnnotationKey()
	zeroQuantity                  = resource.MustParse("0")

	nodeGroupAutoscalingOptionsKeyPrefix = getNodeGroupAutoscalingOptionsKeyPrefix()

	systemArchitecture *SystemArchitecture
	once               sync.Once
)

type normalizedProviderID string

// SystemArchitecture represents a CPU architecture (e.g., amd64, arm64, ppc64le, s390x).
// It is used to determine the default architecture to use when building the nodes templates for scaling up from zero
// by some cloud providers. This code is the same as the GCE implementation at
// https://github.com/kubernetes/autoscaler/blob/3852f352d96b8763292a9122163c1152dfedec55/cluster-autoscaler/cloudprovider/gce/templates.go#L611-L657
// which is kept to allow for a smooth transition to this package, once the GCE team is ready to use it.
type SystemArchitecture string

// Name returns the string value for SystemArchitecture
func (s SystemArchitecture) Name() string {
	return string(s)
}

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

func autoscalingOptions(annotations map[string]string) map[string]string {
	options := map[string]string{}
	for k, v := range annotations {
		if !strings.HasPrefix(k, nodeGroupAutoscalingOptionsKeyPrefix) {
			continue
		}
		resourceName := strings.Split(k, nodeGroupAutoscalingOptionsKeyPrefix)
		if len(resourceName) < 2 || resourceName[1] == "" || v == "" {
			continue
		}
		options[resourceName[1]] = strings.ToLower(v)
	}
	return options
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
	if u != nil {
		for _, ref := range u.GetOwnerReferences() {
			if ref.Kind == kind && ref.Name != "" {
				return ref.DeepCopy()
			}
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
	if strings.HasPrefix(s, "azure://") && strings.Contains(s, "virtualMachineScaleSets") {
		return normalizedProviderID(s)
	}
	split := strings.Split(s, "/")
	return normalizedProviderID(split[len(split)-1])
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

func parseDRADriver(annotations map[string]string) string {
	if val, found := annotations[draDriverKey]; found {
		return val
	}
	return ""
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

// getNodeGroupAutoscalingOptionsKeyPrefix returns the key that is used for autoscaling options
// per node group which override autoscaler default options.
func getNodeGroupAutoscalingOptionsKeyPrefix() string {
	key := fmt.Sprintf("%s/autoscaling-options-", getCAPIGroup())
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

// getClusterNameAnnotationKey returns the key that is used by cluster-api for annotating nodes
// with their cluster name.
func getClusterNameAnnotationKey() string {
	key := fmt.Sprintf("%s/cluster-name", getCAPIGroup())
	return key
}

// getClusterNamespaceAnnotationKey returns the key that is used by cluster-api for annotating nodes
// with their cluster namespace.
func getClusterNamespaceAnnotationKey() string {
	key := fmt.Sprintf("%s/cluster-namespace", getCAPIGroup())
	return key
}

// getClusterNameLabel returns the key that is used by cluster-api for labeling
// which cluster an object belongs to. This function is needed because the user can change
// the default group name by using the CAPI_GROUP environment variable.
func getClusterNameLabel() string {
	key := fmt.Sprintf("%s/cluster-name", getCAPIGroup())
	return key
}

// SystemArchitectureFromString parses a string to SystemArchitecture. Returns UnknownArch if the string doesn't represent a
// valid architecture.
func SystemArchitectureFromString(arch string) SystemArchitecture {
	switch arch {
	case string(Arm64):
		return Arm64
	case string(Amd64):
		return Amd64
	case string(Ppc64le):
		return Ppc64le
	case string(S390x):
		return S390x
	default:
		return UnknownArch
	}
}

// GetDefaultScaleFromZeroArchitecture returns the SystemArchitecture from the environment variable
// CAPI_SCALE_ZERO_DEFAULT_ARCH or DefaultArch if the variable is set to an invalid value.
func GetDefaultScaleFromZeroArchitecture() SystemArchitecture {
	once.Do(func() {
		archStr := os.Getenv(scaleUpFromZeroDefaultArchEnvVar)
		arch := SystemArchitectureFromString(archStr)
		klog.V(5).Infof("the default scale from zero architecture value is set to %s (%s)", archStr, arch.Name())
		if arch == UnknownArch {
			arch = DefaultArch
			klog.Errorf("Unrecognized architecture '%s', falling back to %s",
				scaleUpFromZeroDefaultArchEnvVar, DefaultArch.Name())
		}
		systemArchitecture = &arch
	})
	return *systemArchitecture
}

// getManagedNodeLabelsFromLabels returns a map of labels that will be propagated
// to nodes based on the Cluster API metadata propagation rules.
func getManagedNodeLabelsFromLabels(labels map[string]string) map[string]string {
	// TODO elmiko, add a user configuration to inject a string with their `--additional-sync-machine-labels` string.
	// ref: https://cluster-api.sigs.k8s.io/reference/api/metadata-propagation#machine
	managedLabels := map[string]string{}
	for key, value := range labels {
		if isManagedLabel(key) {
			managedLabels[key] = value
		}

	}

	return managedLabels
}

func isManagedLabel(key string) bool {
	dnsSubdomainOrName := strings.Split(key, "/")[0]
	if dnsSubdomainOrName == nodeRoleLabelPrefix {
		return true
	}
	if dnsSubdomainOrName == nodeRestrictionLabelDomain || strings.HasSuffix(dnsSubdomainOrName, "."+nodeRestrictionLabelDomain) {
		return true
	}
	if dnsSubdomainOrName == managedNodeLabelDomain || strings.HasSuffix(dnsSubdomainOrName, "."+managedNodeLabelDomain) {
		return true
	}
	return false
}
