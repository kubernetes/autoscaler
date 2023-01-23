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

package gce

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	gce "google.golang.org/api/compute/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

// GceTemplateBuilder builds templates for GCE nodes.
type GceTemplateBuilder struct{}

// LocalSSDDiskSizeInGiB is the size of each local SSD in GiB
// (cf. https://cloud.google.com/compute/docs/disks/local-ssd)
const LocalSSDDiskSizeInGiB = 375

// These annotations are used internally only to store information in node temlate and use it later in CA, the actuall nodes won't have these annotations.
const (
	// LocalSsdCountAnnotation is the annotation for number of attached local SSDs to the node.
	LocalSsdCountAnnotation = "cluster-autoscaler/gce/local-ssd-count"
	// BootDiskTypeAnnotation is the annotation for boot disk type of the node.
	BootDiskTypeAnnotation = "cluster-autoscaler/gce/boot-disk-type"
	// BootDiskSizeAnnotation is the annotation for boot disk sise of the node/
	BootDiskSizeAnnotation = "cluster-autoscaler/gce/boot-disk-size"
	// EphemeralStorageLocalSsdAnnotation is the annotation for nodes where ephemeral storage is backed up by local SSDs.
	EphemeralStorageLocalSsdAnnotation = "cluster-autoscaler/gce/ephemeral-storage-local-ssd"
)

// TODO: This should be imported from sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/common/constants.go
// This key is applicable to both GCE and GKE
const gceCSITopologyKeyZone = "topology.gke.io/zone"

func (t *GceTemplateBuilder) getAcceleratorCount(accelerators []*gce.AcceleratorConfig) int64 {
	count := int64(0)
	for _, accelerator := range accelerators {
		if strings.HasPrefix(accelerator.AcceleratorType, "nvidia-") {
			count += accelerator.AcceleratorCount
		}
	}
	return count
}

// BuildCapacity builds a list of resource capacities given list of hardware.
func (t *GceTemplateBuilder) BuildCapacity(m MigOsInfo, cpu int64, mem int64, accelerators []*gce.AcceleratorConfig,
	ephemeralStorage int64, ephemeralStorageLocalSSDCount int64, pods *int64, r OsReservedCalculator, extendedResources apiv1.ResourceList) (apiv1.ResourceList, error) {
	capacity := apiv1.ResourceList{}
	if pods == nil {
		capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	} else {
		capacity[apiv1.ResourcePods] = *resource.NewQuantity(*pods, resource.DecimalSI)
	}

	capacity[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
	memTotal := mem - r.CalculateKernelReserved(m, mem)
	capacity[apiv1.ResourceMemory] = *resource.NewQuantity(memTotal, resource.DecimalSI)

	if accelerators != nil && len(accelerators) > 0 {
		capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(t.getAcceleratorCount(accelerators), resource.DecimalSI)
	}

	if ephemeralStorage > 0 {
		var storageTotal int64
		if ephemeralStorageLocalSSDCount > 0 {
			storageTotal = ephemeralStorage - EphemeralStorageOnLocalSSDFilesystemOverheadInBytes(ephemeralStorageLocalSSDCount, m.OsDistribution())
		} else {
			storageTotal = ephemeralStorage - r.CalculateOSReservedEphemeralStorage(m, ephemeralStorage)
		}
		capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(math.Max(float64(storageTotal), 0)), resource.DecimalSI)
	}

	for resourceName, quantity := range extendedResources {
		capacity[resourceName] = quantity
	}

	return capacity, nil
}

// BuildAllocatableFromKubeEnv builds node allocatable based on capacity of the node and
// value of kubeEnv.
// KubeEnv is a multi-line string containing entries in the form of
// <RESOURCE_NAME>:<string>. One of the resources it contains is a list of
// kubelet arguments from which we can extract the resources reserved by
// the kubelet for its operation. Allocated resources are capacity minus reserved.
// If we fail to extract the reserved resources from kubeEnv (e.g it is in a
// wrong format or does not contain kubelet arguments), we return an error.
func (t *GceTemplateBuilder) BuildAllocatableFromKubeEnv(capacity apiv1.ResourceList, kubeEnv string, evictionHard *EvictionHard) (apiv1.ResourceList, error) {
	kubeReserved, err := extractKubeReservedFromKubeEnv(kubeEnv)
	if err != nil {
		return nil, err
	}
	reserved, err := parseKubeReserved(kubeReserved)
	if err != nil {
		return nil, err
	}
	return t.CalculateAllocatable(capacity, reserved, evictionHard), nil
}

// CalculateAllocatable computes allocatable resources subtracting kube reserved values
// and kubelet eviction memory buffer from corresponding capacity.
func (t *GceTemplateBuilder) CalculateAllocatable(capacity apiv1.ResourceList, kubeReserved apiv1.ResourceList, evictionHard *EvictionHard) apiv1.ResourceList {
	allocatable := apiv1.ResourceList{}
	for key, value := range capacity {
		quantity := value.DeepCopy()
		if reservedQuantity, found := kubeReserved[key]; found {
			quantity.Sub(reservedQuantity)
		}
		if key == apiv1.ResourceMemory {
			quantity = *resource.NewQuantity(quantity.Value()-GetKubeletEvictionHardForMemory(evictionHard), resource.BinarySI)
		}
		if key == apiv1.ResourceEphemeralStorage {
			quantity = *resource.NewQuantity(quantity.Value()-int64(GetKubeletEvictionHardForEphemeralStorage(value.Value(), evictionHard)), resource.BinarySI)
		}
		allocatable[key] = quantity
	}
	return allocatable
}

func getKubeEnvValueFromTemplateMetadata(template *gce.InstanceTemplate) (string, error) {
	if template.Properties.Metadata == nil {
		return "", fmt.Errorf("instance template %s has no metadata", template.Name)
	}
	for _, item := range template.Properties.Metadata.Items {
		if item.Key == "kube-env" {
			if item.Value == nil {
				return "", fmt.Errorf("no kube-env content in metadata")
			}
			return *item.Value, nil
		}
	}
	return "", nil
}

// MigOsInfo return os detailes information that stored in template.
func (t *GceTemplateBuilder) MigOsInfo(migId string, template *gce.InstanceTemplate) (MigOsInfo, error) {
	kubeEnvValue, err := getKubeEnvValueFromTemplateMetadata(template)
	if err != nil {
		return nil, fmt.Errorf("could not obtain kube-env from template metadata; %v", err)
	}
	os := extractOperatingSystemFromKubeEnv(kubeEnvValue)
	if os == OperatingSystemUnknown {
		return nil, fmt.Errorf("could not obtain os from kube-env from template metadata")
	}

	osDistribution := extractOperatingSystemDistributionFromKubeEnv(kubeEnvValue)
	if osDistribution == OperatingSystemDistributionUnknown {
		return nil, fmt.Errorf("could not obtain os-distribution from kube-env from template metadata")
	}
	arch, err := extractSystemArchitectureFromKubeEnv(kubeEnvValue)
	if err != nil {
		arch = DefaultArch
		klog.Errorf("Couldn't extract architecture from kube-env for MIG %q, falling back to %q. Error: %v", migId, arch, err)
	}
	return NewMigOsInfo(os, osDistribution, arch), nil
}

// BuildNodeFromTemplate builds node from provided GCE template.
func (t *GceTemplateBuilder) BuildNodeFromTemplate(mig Mig, migOsInfo MigOsInfo, template *gce.InstanceTemplate, cpu int64, mem int64, pods *int64, reserved OsReservedCalculator) (*apiv1.Node, error) {

	if template.Properties == nil {
		return nil, fmt.Errorf("instance template %s has no properties", template.Name)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-template-%d", template.Name, rand.Int63())

	kubeEnvValue, err := getKubeEnvValueFromTemplateMetadata(template)
	if err != nil {
		return nil, fmt.Errorf("could not obtain kube-env from template metadata; %v", err)
	}

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	addBootDiskAnnotations(&node, template.Properties)
	var ephemeralStorage int64 = -1
	if !isBootDiskEphemeralStorageWithInstanceTemplateDisabled(kubeEnvValue) {
		// ephemeral storage is backed up by boot disk
		ephemeralStorage, err = getBootDiskEphemeralStorageFromInstanceTemplateProperties(template.Properties)
	} else {
		// ephemeral storage is backed up by local ssd
		addAnnotation(&node, EphemeralStorageLocalSsdAnnotation, strconv.FormatBool(true))
	}

	localSsdCount, err := getLocalSsdCount(template.Properties)
	if localSsdCount > 0 {
		addAnnotation(&node, LocalSsdCountAnnotation, strconv.FormatInt(localSsdCount, 10))
	}
	ephemeralStorageLocalSsdCount := ephemeralStorageLocalSSDCount(kubeEnvValue)
	if err == nil && ephemeralStorageLocalSsdCount > 0 {
		ephemeralStorage, err = getEphemeralStorageOnLocalSsd(localSsdCount, ephemeralStorageLocalSsdCount)
	}
	if err != nil {
		return nil, fmt.Errorf("could not fetch ephemeral storage from instance template: %v", err)
	}

	extendedResources, err := extractExtendedResourcesFromKubeEnv(kubeEnvValue)
	if err != nil {
		// External Resources are optional and should not break the template creation
		klog.Errorf("could not fetch extended resources from instance template: %v", err)
	}

	capacity, err := t.BuildCapacity(migOsInfo, cpu, mem, template.Properties.GuestAccelerators, ephemeralStorage, ephemeralStorageLocalSsdCount, pods, reserved, extendedResources)
	if err != nil {
		return nil, err
	}

	node.Status = apiv1.NodeStatus{
		Capacity: capacity,
	}
	var nodeAllocatable apiv1.ResourceList

	if kubeEnvValue != "" {
		// Extract labels
		kubeEnvLabels, err := extractLabelsFromKubeEnv(kubeEnvValue)
		if err != nil {
			return nil, err
		}
		node.Labels = cloudprovider.JoinStringMaps(node.Labels, kubeEnvLabels)

		// Extract taints
		kubeEnvTaints, err := extractTaintsFromKubeEnv(kubeEnvValue)
		if err != nil {
			return nil, err
		}
		node.Spec.Taints = append(node.Spec.Taints, kubeEnvTaints...)

		// Extract Eviction Hard
		evictionHardFromKubeEnv, err := extractEvictionHardFromKubeEnv(kubeEnvValue)
		if err != nil || len(evictionHardFromKubeEnv) == 0 {
			klog.Warning("unable to get evictionHardFromKubeEnv values, continuing without it.")
		}
		evictionHard := ParseEvictionHardOrGetDefault(evictionHardFromKubeEnv)

		if allocatable, err := t.BuildAllocatableFromKubeEnv(node.Status.Capacity, kubeEnvValue, evictionHard); err == nil {
			nodeAllocatable = allocatable
		}
	}

	if nodeAllocatable == nil {
		klog.Warningf("could not extract kube-reserved from kubeEnv for mig %q, setting allocatable to capacity.", mig.GceRef().Name)
		node.Status.Allocatable = node.Status.Capacity
	} else {
		node.Status.Allocatable = nodeAllocatable
	}
	// GenericLabels
	labels, err := BuildGenericLabels(mig.GceRef(), template.Properties.MachineType, nodeName, migOsInfo.Os(), migOsInfo.Arch())
	if err != nil {
		return nil, err
	}
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, labels)

	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func ephemeralStorageLocalSSDCount(kubeEnvValue string) int64 {
	v, found, err := extractAutoscalerVarFromKubeEnv(kubeEnvValue, "ephemeral_storage_local_ssd_count")
	if err != nil {
		klog.Warningf("cannot extract ephemeral_storage_local_ssd_count from kube-env, default to 0: %v", err)
		return 0
	}

	if !found {
		return 0
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		klog.Warningf("cannot parse ephemeral_storage_local_ssd_count value, default to 0: %v", err)
		return 0
	}

	return int64(n)
}

func getLocalSsdCount(instanceProperties *gce.InstanceProperties) (int64, error) {
	if instanceProperties.Disks == nil {
		return 0, fmt.Errorf("instance properties disks is nil")
	}
	var count int64
	for _, disk := range instanceProperties.Disks {
		if disk != nil && disk.InitializeParams != nil {
			if disk.Type == "SCRATCH" && disk.InitializeParams.DiskType == "local-ssd" {
				count++
			}
		}
	}
	return count, nil
}

func getEphemeralStorageOnLocalSsd(localSsdCount, ephemeralStorageLocalSsdCount int64) (int64, error) {
	if localSsdCount < ephemeralStorageLocalSsdCount {
		return 0, fmt.Errorf("actual local SSD count is lower than ephemeral_storage_local_ssd_count")
	}
	return ephemeralStorageLocalSsdCount * LocalSSDDiskSizeInGiB * units.GiB, nil
}

// isBootDiskEphemeralStorageWithInstanceTemplateDisabled will allow bypassing Disk Size of Boot Disk from being
// picked up from Instance Template and used as Ephemeral Storage, in case other type of storage are used
// as ephemeral storage
func isBootDiskEphemeralStorageWithInstanceTemplateDisabled(kubeEnvValue string) bool {
	v, found, err := extractAutoscalerVarFromKubeEnv(kubeEnvValue, "BLOCK_EPH_STORAGE_BOOT_DISK")
	if err == nil && found && v == "true" {
		return true
	}
	return false
}

func getBootDiskEphemeralStorageFromInstanceTemplateProperties(instanceProperties *gce.InstanceProperties) (ephemeralStorage int64, err error) {
	if instanceProperties.Disks == nil {
		return 0, fmt.Errorf("unable to get ephemeral storage because instance properties disks is nil")
	}

	for _, disk := range instanceProperties.Disks {
		if disk != nil && disk.InitializeParams != nil {
			if disk.Boot {
				return disk.InitializeParams.DiskSizeGb * units.GiB, nil
			}
		}
	}

	return 0, fmt.Errorf("unable to get ephemeral storage, either no attached disks or no disk with boot=true")
}

// BuildGenericLabels builds basic labels that should be present on every GCE node,
// including hostname, zone etc.
func BuildGenericLabels(ref GceRef, machineType string, nodeName string, os OperatingSystem, arch SystemArchitecture) (map[string]string, error) {
	result := make(map[string]string)

	if os == OperatingSystemUnknown {
		return nil, fmt.Errorf("unknown operating system passed")
	}

	// TODO: extract it somehow
	result[apiv1.LabelArchStable] = arch.Name()
	result[apiv1.LabelOSStable] = string(os)

	result[apiv1.LabelInstanceTypeStable] = machineType
	ix := strings.LastIndex(ref.Zone, "-")
	if ix == -1 {
		return nil, fmt.Errorf("unexpected zone: %s", ref.Zone)
	}
	result[apiv1.LabelTopologyRegion] = ref.Zone[:ix]
	result[apiv1.LabelTopologyZone] = ref.Zone
	result[gceCSITopologyKeyZone] = ref.Zone
	result[apiv1.LabelHostname] = nodeName
	return result, nil
}

func parseKubeReserved(kubeReserved string) (apiv1.ResourceList, error) {
	resourcesMap, err := parseKeyValueListToMap(kubeReserved)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kube-reserved from kube-env: %q", err)
	}
	reservedResources := apiv1.ResourceList{}
	for name, quantity := range resourcesMap {
		switch apiv1.ResourceName(name) {
		case apiv1.ResourceCPU, apiv1.ResourceMemory, apiv1.ResourceEphemeralStorage:
			if q, err := resource.ParseQuantity(quantity); err == nil && q.Sign() >= 0 {
				reservedResources[apiv1.ResourceName(name)] = q
			}
		default:
			klog.Warningf("ignoring resource from kube-reserved: %q", name)
		}
	}
	return reservedResources, nil
}

// GetLabelsFromTemplate returns labels from instance template
func GetLabelsFromTemplate(template *gce.InstanceTemplate) (map[string]string, error) {
	kubeEnv, err := getKubeEnvValueFromTemplateMetadata(template)
	if err != nil {
		return nil, err
	}
	return extractLabelsFromKubeEnv(kubeEnv)
}

func extractLabelsFromKubeEnv(kubeEnv string) (map[string]string, error) {
	// In v1.10+, labels are only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	labels, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "node_labels")
	if err != nil {
		klog.Errorf("error while trying to extract node_labels from AUTOSCALER_ENV_VARS: %v", err)
	}
	if !found {
		labels, err = extractFromKubeEnv(kubeEnv, "NODE_LABELS")
		if err != nil {
			return nil, err
		}
	}
	return parseKeyValueListToMap(labels)
}

// GetTaintsFromTemplate returns labels from instance template
func GetTaintsFromTemplate(template *gce.InstanceTemplate) ([]apiv1.Taint, error) {
	kubeEnv, err := getKubeEnvValueFromTemplateMetadata(template)
	if err != nil {
		return nil, err
	}
	return extractTaintsFromKubeEnv(kubeEnv)
}

func extractTaintsFromKubeEnv(kubeEnv string) ([]apiv1.Taint, error) {
	// In v1.10+, taints are only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	taints, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "node_taints")
	if err != nil {
		klog.Errorf("error while trying to extract node_taints from AUTOSCALER_ENV_VARS: %v", err)
	}
	if !found {
		taints, err = extractFromKubeEnv(kubeEnv, "NODE_TAINTS")
		if err != nil {
			return nil, err
		}
	}
	taintMap, err := parseKeyValueListToMap(taints)
	if err != nil {
		return nil, err
	}
	return buildTaints(taintMap)
}

func extractKubeReservedFromKubeEnv(kubeEnv string) (string, error) {
	// In v1.10+, kube-reserved is only exposed for the autoscaler via AUTOSCALER_ENV_VARS
	// see kubernetes/kubernetes#61119. We try AUTOSCALER_ENV_VARS first, then
	// fall back to the old way.
	kubeReserved, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "kube_reserved")
	if err != nil {
		klog.Errorf("error while trying to extract kube_reserved from AUTOSCALER_ENV_VARS: %v", err)
	}
	if !found {
		kubeletArgs, err := extractFromKubeEnv(kubeEnv, "KUBELET_TEST_ARGS")
		if err != nil {
			return "", err
		}
		resourcesRegexp := regexp.MustCompile(`--kube-reserved=([^ ]+)`)

		matches := resourcesRegexp.FindStringSubmatch(kubeletArgs)
		if len(matches) > 1 {
			return matches[1], nil
		}
		return "", fmt.Errorf("kube-reserved not in kubelet args in kube-env: %q", kubeletArgs)
	}
	return kubeReserved, nil
}

func extractExtendedResourcesFromKubeEnv(kubeEnvValue string) (apiv1.ResourceList, error) {
	extendedResourcesAsString, found, err := extractAutoscalerVarFromKubeEnv(kubeEnvValue, "extended_resources")
	if err != nil {
		klog.Warning("error while obtaining extended_resources from AUTOSCALER_ENV_VARS; %v", err)
		return nil, err
	}

	if !found {
		return apiv1.ResourceList{}, nil
	}

	extendedResourcesMap, err := parseKeyValueListToMap(extendedResourcesAsString)
	if err != nil {
		return apiv1.ResourceList{}, err
	}

	extendedResources := apiv1.ResourceList{}
	for name, quantity := range extendedResourcesMap {
		if q, err := resource.ParseQuantity(quantity); err == nil && q.Sign() >= 0 {
			extendedResources[apiv1.ResourceName(name)] = q
		} else if err != nil {
			klog.Warning("ignoring invalid value in extended_resources defined in AUTOSCALER_ENV_VARS; %v", err)
		}
	}
	return extendedResources, nil
}

// OperatingSystem denotes operating system used by nodes coming from node group
type OperatingSystem string

const (
	// OperatingSystemUnknown is used if operating system is unknown
	OperatingSystemUnknown OperatingSystem = ""
	// OperatingSystemLinux is used if operating system is Linux
	OperatingSystemLinux OperatingSystem = "linux"
	// OperatingSystemWindows is used if operating system is Windows
	OperatingSystemWindows OperatingSystem = "windows"

	// OperatingSystemDefault defines which operating system will be assumed if not explicitly passed via AUTOSCALER_ENV_VARS
	OperatingSystemDefault = OperatingSystemLinux
)

func extractOperatingSystemFromKubeEnv(kubeEnv string) OperatingSystem {
	osValue, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "os")
	if err != nil {
		klog.Errorf("error while obtaining os from AUTOSCALER_ENV_VARS; %v", err)
		return OperatingSystemUnknown
	}

	if !found {
		klog.Warningf("no os defined in AUTOSCALER_ENV_VARS; using default %v", OperatingSystemDefault)
		return OperatingSystemDefault
	}

	switch osValue {
	case string(OperatingSystemLinux):
		return OperatingSystemLinux
	case string(OperatingSystemWindows):
		return OperatingSystemWindows
	default:
		klog.Errorf("unexpected os=%v passed via AUTOSCALER_ENV_VARS", osValue)
		return OperatingSystemUnknown
	}
}

// OperatingSystemImage denotes  image of the operating system used by nodes coming from node group
type OperatingSystemImage string

const (
	// OperatingSystemImageUnknown is used if operating distribution system is unknown
	OperatingSystemImageUnknown OperatingSystemImage = ""
	// OperatingSystemImageUbuntu is used if operating distribution system is Ubuntu
	OperatingSystemImageUbuntu OperatingSystemImage = "ubuntu"
	// OperatingSystemImageWindowsLTSC is used if operating distribution system is Windows LTSC
	OperatingSystemImageWindowsLTSC OperatingSystemImage = "windows_ltsc"
	// OperatingSystemImageWindowsSAC is used if operating distribution system is Windows SAC
	OperatingSystemImageWindowsSAC OperatingSystemImage = "windows_sac"
	// OperatingSystemImageCOS is used if operating distribution system is COS
	OperatingSystemImageCOS OperatingSystemImage = "cos"
	// OperatingSystemImageCOSContainerd is used if operating distribution system is COS Containerd
	OperatingSystemImageCOSContainerd OperatingSystemImage = "cos_containerd"
	// OperatingSystemImageUbuntuContainerd is used if operating distribution system is Ubuntu Containerd
	OperatingSystemImageUbuntuContainerd OperatingSystemImage = "ubuntu_containerd"
	// OperatingSystemImageWindowsLTSCContainerd is used if operating distribution system is Windows LTSC Containerd
	OperatingSystemImageWindowsLTSCContainerd OperatingSystemImage = "windows_ltsc_containerd"
	// OperatingSystemImageWindowsSACContainerd is used if operating distribution system is Windows SAC Containerd
	OperatingSystemImageWindowsSACContainerd OperatingSystemImage = "windows_sac_containerd"

	// OperatingSystemImageDefault defines which operating system will be assumed as default.
	OperatingSystemImageDefault = OperatingSystemImageCOSContainerd
)

// OperatingSystemDistribution denotes  distribution of the operating system used by nodes coming from node group
type OperatingSystemDistribution string

const (
	// OperatingSystemDistributionUnknown is used if operating distribution system is unknown
	OperatingSystemDistributionUnknown OperatingSystemDistribution = ""
	// OperatingSystemDistributionUbuntu is used if operating distribution system is Ubuntu
	OperatingSystemDistributionUbuntu OperatingSystemDistribution = "ubuntu"
	// OperatingSystemDistributionWindowsLTSC is used if operating distribution system is Windows LTSC
	OperatingSystemDistributionWindowsLTSC OperatingSystemDistribution = "windows_ltsc"
	// OperatingSystemDistributionWindowsSAC is used if operating distribution system is Windows SAC
	OperatingSystemDistributionWindowsSAC OperatingSystemDistribution = "windows_sac"
	// OperatingSystemDistributionCOS is used if operating distribution system is COS
	OperatingSystemDistributionCOS OperatingSystemDistribution = "cos"

	// OperatingSystemDistributionDefault defines which operating system will be assumed if not explicitly passed via AUTOSCALER_ENV_VARS
	OperatingSystemDistributionDefault = OperatingSystemDistributionCOS
)

func extractOperatingSystemDistributionFromImageType(imageType string) OperatingSystemDistribution {
	switch imageType {
	case string(OperatingSystemImageUbuntu), string(OperatingSystemImageUbuntuContainerd):
		return OperatingSystemDistributionUbuntu
	case string(OperatingSystemImageWindowsLTSC), string(OperatingSystemImageWindowsLTSCContainerd):
		return OperatingSystemDistributionWindowsLTSC
	case string(OperatingSystemImageWindowsSAC), string(OperatingSystemImageWindowsSACContainerd):
		return OperatingSystemDistributionWindowsSAC
	case string(OperatingSystemImageCOS), string(OperatingSystemImageCOSContainerd):
		return OperatingSystemDistributionCOS
	default:
		return OperatingSystemDistributionUnknown
	}
}

// SystemArchitecture denotes distribution of the System Architecture used by nodes coming from node group
type SystemArchitecture string

const (
	// UnknownArch is used if the Architecture is Unknown
	UnknownArch SystemArchitecture = ""
	// Amd64 is used if the Architecture is x86_64
	Amd64 SystemArchitecture = "amd64"
	// Arm64 is used if the Architecture is ARM
	Arm64 SystemArchitecture = "arm64"
	// DefaultArch is used if the Architecture is used as a fallback if not passed by AUTOSCALER_ENV_VARS
	DefaultArch SystemArchitecture = Amd64
)

// Name returns the string value for SystemArchitecture
func (s SystemArchitecture) Name() string {
	return string(s)
}

func extractSystemArchitectureFromKubeEnv(kubeEnv string) (SystemArchitecture, error) {
	archName, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "arch")
	if err != nil {
		return UnknownArch, fmt.Errorf("error while obtaining arch from AUTOSCALER_ENV_VARS: %v", err)
	}
	if !found {
		return UnknownArch, fmt.Errorf("no arch defined in AUTOSCALER_ENV_VARS")
	}
	arch := ToSystemArchitecture(archName)
	if arch == UnknownArch {
		return UnknownArch, fmt.Errorf("unknown arch %q defined in AUTOSCALER_ENV_VARS", archName)
	}
	return arch, nil
}

// ToSystemArchitecture parses a string to SystemArchitecture. Returns UnknownArch if the string doesn't represent a
// valid architecture.
func ToSystemArchitecture(arch string) SystemArchitecture {
	switch arch {
	case string(Arm64):
		return Arm64
	case string(Amd64):
		return Amd64
	default:
		return UnknownArch
	}
}

func extractOperatingSystemDistributionFromKubeEnv(kubeEnv string) OperatingSystemDistribution {
	osDistributionValue, found, err := extractAutoscalerVarFromKubeEnv(kubeEnv, "os_distribution")
	if err != nil {
		klog.Errorf("error while obtaining os from AUTOSCALER_ENV_VARS; %v", err)
		return OperatingSystemDistributionUnknown
	}

	if !found {
		klog.Warningf("no os-distribution defined in AUTOSCALER_ENV_VARS; using default %v", OperatingSystemDistributionDefault)
		return OperatingSystemDistributionDefault
	}

	switch osDistributionValue {

	case string(OperatingSystemDistributionUbuntu):
		return OperatingSystemDistributionUbuntu
	case string(OperatingSystemDistributionWindowsLTSC):
		return OperatingSystemDistributionWindowsLTSC
	case string(OperatingSystemDistributionWindowsSAC):
		return OperatingSystemDistributionWindowsSAC
	case string(OperatingSystemDistributionCOS):
		return OperatingSystemDistributionCOS
	// Deprecated
	case "cos_containerd":
		klog.Warning("cos_containerd os distribution is deprecated")
		return OperatingSystemDistributionCOS
	// Deprecated
	case "ubuntu_containerd":
		klog.Warning("ubuntu_containerd os distribution is deprecated")
		return OperatingSystemDistributionUbuntu
	default:
		klog.Errorf("unexpected os-distribution=%v passed via AUTOSCALER_ENV_VARS", osDistributionValue)
		return OperatingSystemDistributionUnknown
	}
}

func getFloat64Option(options map[string]string, templateName, name string) (float64, bool) {
	raw, ok := options[name]
	if !ok {
		return 0, false
	}

	option, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		klog.Warningf("failed to convert autoscaling_options option %q (value %q) for MIG %q to float: %v", name, raw, templateName, err)
		return 0, false
	}

	return option, true
}

func getDurationOption(options map[string]string, templateName, name string) (time.Duration, bool) {
	raw, ok := options[name]
	if !ok {
		return 0, false
	}

	option, err := time.ParseDuration(raw)
	if err != nil {
		klog.Warningf("failed to convert autoscaling_options option %q (value %q) for MIG %q to duration: %v", name, raw, templateName, err)
		return 0, false
	}

	return option, true
}

func extractAutoscalingOptionsFromKubeEnv(kubeEnvValue string) (map[string]string, error) {
	optionsAsString, found, err := extractAutoscalerVarFromKubeEnv(kubeEnvValue, "autoscaling_options")
	if err != nil {
		klog.Warningf("error while obtaining autoscaling_options from AUTOSCALER_ENV_VARS: %v", err)
		return nil, err
	}

	if !found {
		klog.V(5).Info("no autoscaling_options defined in AUTOSCALER_ENV_VARS")
		return make(map[string]string), nil
	}

	return parseKeyValueListToMap(optionsAsString)
}

func extractEvictionHardFromKubeEnv(kubeEnvValue string) (map[string]string, error) {
	evictionHardAsString, found, err := extractAutoscalerVarFromKubeEnv(kubeEnvValue, "evictionHard")
	if err != nil {
		klog.Warning("error while obtaining eviction-hard from AUTOSCALER_ENV_VARS; %v", err)
		return nil, err
	}

	if !found {
		klog.Warning("no evictionHard defined in AUTOSCALER_ENV_VARS;")
		return make(map[string]string), nil
	}

	return parseKeyValueListToMap(evictionHardAsString)
}

func extractAutoscalerVarFromKubeEnv(kubeEnv, name string) (value string, found bool, err error) {
	const autoscalerVars = "AUTOSCALER_ENV_VARS"
	autoscalerVals, err := extractFromKubeEnv(kubeEnv, autoscalerVars)
	if err != nil {
		return "", false, err
	}

	if strings.Trim(autoscalerVals, " ") == "" {
		// empty or not present AUTOSCALER_ENV_VARS
		return "", false, nil
	}

	for _, val := range strings.Split(autoscalerVals, ";") {
		val = strings.Trim(val, " ")
		items := strings.SplitN(val, "=", 2)
		if len(items) != 2 {
			return "", false, fmt.Errorf("malformed autoscaler var: %s", val)
		}
		if strings.Trim(items[0], " ") == name {
			return strings.Trim(items[1], " \"'"), true, nil
		}
	}
	klog.V(5).Infof("var %s not found in %s: %v", name, autoscalerVars, autoscalerVals)
	return "", false, nil
}

func extractFromKubeEnv(kubeEnv, resource string) (string, error) {
	kubeEnvMap := make(map[string]string)
	err := yaml.Unmarshal([]byte(kubeEnv), &kubeEnvMap)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling kubeEnv: %v", err)
	}
	return kubeEnvMap[resource], nil
}

func parseKeyValueListToMap(kvList string) (map[string]string, error) {
	result := make(map[string]string)
	if len(kvList) == 0 {
		return result, nil
	}
	for _, keyValue := range strings.Split(kvList, ",") {
		kvItems := strings.SplitN(keyValue, "=", 2)
		if len(kvItems) != 2 {
			return nil, fmt.Errorf("error while parsing key-value list, val: %s", keyValue)
		}
		result[kvItems[0]] = kvItems[1]
	}
	return result, nil
}

func buildTaints(kubeEnvTaints map[string]string) ([]apiv1.Taint, error) {
	taints := make([]apiv1.Taint, 0)
	for key, value := range kubeEnvTaints {
		values := strings.SplitN(value, ":", 2)
		if len(values) != 2 {
			return nil, fmt.Errorf("error while parsing node taint value and effect: %s", value)
		}
		taints = append(taints, apiv1.Taint{
			Key:    key,
			Value:  values[0],
			Effect: apiv1.TaintEffect(values[1]),
		})
	}
	return taints, nil
}

func addAnnotation(node *apiv1.Node, key, value string) {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[key] = value
}

func addBootDiskAnnotations(node *apiv1.Node, instanceProperties *gce.InstanceProperties) {
	if instanceProperties.Disks == nil {
		return
	}
	for _, disk := range instanceProperties.Disks {
		if disk != nil && disk.InitializeParams != nil {
			if disk.Boot {
				addAnnotation(node, BootDiskSizeAnnotation, strconv.FormatInt(disk.InitializeParams.DiskSizeGb, 10))
				addAnnotation(node, BootDiskTypeAnnotation, disk.InitializeParams.DiskType)
			}
		}
	}
}
