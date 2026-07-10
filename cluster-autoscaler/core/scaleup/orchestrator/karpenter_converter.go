package orchestrator

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/karpenter"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"

	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	karpentercloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	karpenterscheduling "sigs.k8s.io/karpenter/pkg/scheduling"
	karpenterresources "sigs.k8s.io/karpenter/pkg/utils/resources"
)

const caIgnoreReservedValue = "ca-ignore-reserved-value"

const (
	fallbackCPUCoefficient = 1.0
	fallbackMemCoefficient = 0.125
	fallbackGPUCoefficient = 25.0
	fallbackSpotDiscount   = 0.3
)

type poolGroupKey struct {
	taints string
}

type itIdentity struct {
	itName    string
	resources string
	dsKeys    string
}

// DefaultKarpenterConverter implements karpenter.KarpenterConverter.
type DefaultKarpenterConverter struct {
	// PricingModel is used to calculate offering prices.
	PricingModel cloudprovider.PricingModel
	// IgnoredLabels contains additional labels to ignore when collapsing instance types.
	IgnoredLabels []string
}

// isIgnoredLabel returns true if a label key should be excluded from static custom label
// grouping when converting CA NodeGroups into Karpenter NodePools and InstanceTypes.
//
// WHY these labels receive special treatment:
// 1. Topology & Hostname Labels (kubernetes.io/hostname, topology.kubernetes.io/zone,
//    topology.kubernetes.io/region, and legacy beta variants):
//    Physical nodes and template nodes naturally differ by hostname and failure domain.
//    If treated as static custom labels distinguishing NodePools or InstanceTypes, identical
//    compute shapes in different zones would be fragmented into separate, isolated NodePools.
//    Instead, topology labels are stripped from static identity and modeled dynamically as
//    offering-level requirements on each InstanceType.
// 2. Capacity Type (karpenter.sh/capacity-type):
//    On-demand vs. spot distinction is modeled as a dynamic offering requirement on the
//    InstanceType rather than a static custom label that splits NodePools.
// 3. CA Internal Metadata (cluster-autoscaler.kubernetes.io/template-node):
//    Synthetic bookkeeping labels attached by CA to mock/template nodes are internal
//    annotations and must not pollute workload scheduling requirements or fragment pools.
// 4. User-Configured Balancing Ignored Labels (c.IgnoredLabels):
//    Provider-specific nodepool tags (e.g. cloud.google.com/gke-nodepool, eks.amazonaws.com/nodegroup)
//    passed via CA's --balancing-ignore-label flag are ignored to allow similar node groups
//    differing only in provider metadata to be collapsed into unified Karpenter offerings.
func (c *DefaultKarpenterConverter) isIgnoredLabel(key string) bool {
	if key == apiv1.LabelHostname ||
		key == apiv1.LabelTopologyZone ||
		key == apiv1.LabelTopologyRegion ||
		key == apiv1.LabelZoneFailureDomain ||
		key == apiv1.LabelZoneRegion ||
		key == apiv1.LabelZoneFailureDomainStable ||
		key == apiv1.LabelZoneRegionStable ||
		key == apiv1.LabelInstanceType ||
		key == apiv1.LabelInstanceTypeStable ||
		key == "beta.kubernetes.io/instance-type" ||
		key == "beta.kubernetes.io/arch" ||
		key == "beta.kubernetes.io/os" ||
		key == karpenterv1.CapacityTypeLabelKey ||
		key == "cluster-autoscaler.kubernetes.io/template-node" {
		return true
	}
	for _, ign := range c.IgnoredLabels {
		if key == ign {
			return true
		}
	}
	return false
}

// NewDefaultKarpenterConverter creates a new DefaultKarpenterConverter initialized with the given pricing model and ignored labels.
func NewDefaultKarpenterConverter(pricingModel cloudprovider.PricingModel, ignoredLabels []string) *DefaultKarpenterConverter {
	return &DefaultKarpenterConverter{
		PricingModel:  pricingModel,
		IgnoredLabels: append([]string{}, ignoredLabels...),
	}
}

// Convert translates CA NodeGroups into Karpenter primitives and returns a stateless ConversionResult.
func (c *DefaultKarpenterConverter) Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) (*karpenter.ConversionResult, error) {
	poolITToNodeGroupsMap := make(map[string]map[string][]cloudprovider.NodeGroup)
	ngToPoolMap := make(map[string]string)
	offeringMapForIT := make(map[string]*karpentercloudprovider.Offering)

	customLabelKeys := c.collectCustomLabelKeys(nodeGroups, nodeInfos)
	poolGroups := c.groupNodeGroupsByTaints(nodeGroups, nodeInfos)

	var nodePools []*karpenterv1.NodePool
	itMap := make(map[string][]*karpentercloudprovider.InstanceType)

	var poolKeys []poolGroupKey
	for k := range poolGroups {
		poolKeys = append(poolKeys, k)
	}
	sort.Slice(poolKeys, func(i, j int) bool {
		return poolKeys[i].taints < poolKeys[j].taints
	})

	for _, key := range poolKeys {
		groups := poolGroups[key]
		hasher := fnv.New64a()
		hasher.Write([]byte(key.taints))
		hashStr := fmt.Sprintf("%016x", hasher.Sum64())
		nodePoolName := fmt.Sprintf("pool-%s", hashStr)

		poolITToNodeGroupsMap[nodePoolName] = make(map[string][]cloudprovider.NodeGroup)
		for _, ng := range groups {
			ngToPoolMap[ng.Id()] = nodePoolName
		}

		collapsedGroups := c.collapseGroupsByStaticIdentity(groups, nodeInfos, customLabelKeys)

		var keys []itIdentity
		for k := range collapsedGroups {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].itName != keys[j].itName {
				return keys[i].itName < keys[j].itName
			}
			if keys[i].resources != keys[j].resources {
				return keys[i].resources < keys[j].resources
			}
			return keys[i].dsKeys < keys[j].dsKeys
		})

		var instanceTypes []*karpentercloudprovider.InstanceType
		for _, id := range keys {
			ngs := collapsedGroups[id]
			sort.Slice(ngs, func(i, j int) bool {
				return ngs[i].Id() < ngs[j].Id()
			})
			physicalITName := id.itName
			itName := physicalITName

			poolITToNodeGroupsMap[nodePoolName][itName] = append(poolITToNodeGroupsMap[nodePoolName][itName], ngs...)
			it := c.CreateInstanceType(itName, physicalITName, ngs, nodeInfos, nodePoolName, customLabelKeys, offeringMapForIT)
			if it != nil {
				instanceTypes = append(instanceTypes, it)
			}
		}

		// Pre-compute minimum offering price for each InstanceType pointer
		minPrices := make(map[*karpentercloudprovider.InstanceType]float64, len(instanceTypes))
		for _, it := range instanceTypes {
			minPrice := math.MaxFloat64
			for _, off := range it.Offerings {
				if off.Price < minPrice {
					minPrice = off.Price
				}
			}
			minPrices[it] = minPrice
		}

		// Sort instance types by price (cheapest first), breaking ties deterministically by unique pointer/name
		sort.Slice(instanceTypes, func(i, j int) bool {
			iPrice := minPrices[instanceTypes[i]]
			jPrice := minPrices[instanceTypes[j]]
			if iPrice != jPrice {
				return iPrice < jPrice
			}
			return instanceTypes[i].Name < instanceTypes[j].Name
		})

		var taints []apiv1.Taint
		if len(groups) > 0 {
			if ni := nodeInfos[groups[0].Id()]; ni != nil && ni.Node() != nil {
				taints = ni.Node().Spec.Taints
			}
		}

		nodePool := &karpenterv1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodePoolName,
				UID:  types.UID(nodePoolName + "-uid"),
			},
			Spec: karpenterv1.NodePoolSpec{
				Template: karpenterv1.NodeClaimTemplate{
					Spec: karpenterv1.NodeClaimTemplateSpec{
						NodeClassRef: &karpenterv1.NodeClassReference{
							Group: "ca.prototype",
							Kind:  "PrototypeNodeClass",
							Name:  "default",
						},
						Taints: taints,
					},
				},
			},
		}

		nodePool.Spec.Template.Spec.Requirements = c.buildNodePoolRequirements(instanceTypes)

		nodePools = append(nodePools, nodePool)
		itMap[nodePoolName] = instanceTypes
	}

	sort.Slice(nodePools, func(i, j int) bool {
		return nodePools[i].Name < nodePools[j].Name
	})

	return &karpenter.ConversionResult{
		NodePools:          nodePools,
		InstanceTypes:      itMap,
		OfferingMap:        offeringMapForIT,
		PoolITToNodeGroups: poolITToNodeGroupsMap,
		NodeGroupToPool:    ngToPoolMap,
	}, nil
}

func (c *DefaultKarpenterConverter) collectCustomLabelKeys(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) sets.Set[string] {
	customLabelKeys := sets.New[string]()
	for _, ng := range nodeGroups {
		ni, found := nodeInfos[ng.Id()]
		if !found || ni == nil || ni.Node() == nil {
			continue
		}
		for k := range ni.Node().Labels {
			if c.isIgnoredLabel(k) {
				continue
			}
			customLabelKeys.Insert(k)
		}
	}
	return customLabelKeys
}

func (c *DefaultKarpenterConverter) groupNodeGroupsByTaints(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) map[poolGroupKey][]cloudprovider.NodeGroup {
	poolGroups := make(map[poolGroupKey][]cloudprovider.NodeGroup)
	for _, ng := range nodeGroups {
		ni, found := nodeInfos[ng.Id()]
		if !found || ni == nil || ni.Node() == nil {
			continue
		}
		node := ni.Node()
		key := poolGroupKey{
			taints: karpenter.SerializeTaints(node.Spec.Taints),
		}
		poolGroups[key] = append(poolGroups[key], ng)
	}
	return poolGroups
}

func (c *DefaultKarpenterConverter) collapseGroupsByStaticIdentity(groups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo, customLabelKeys sets.Set[string]) map[itIdentity][]cloudprovider.NodeGroup {
	collapsedGroups := make(map[itIdentity][]cloudprovider.NodeGroup)
	for _, ng := range groups {
		ni := nodeInfos[ng.Id()]
		if ni == nil || ni.Node() == nil {
			continue
		}
		node := ni.Node()
		id := itIdentity{
			itName:    karpenter.InstanceTypeNameFromLabels(node.Labels, ng.Id()),
			resources: serializeResources(node.Status.Capacity),
			dsKeys:    getCompatibleDSKeys(ng, nodeInfos),
		}
		collapsedGroups[id] = append(collapsedGroups[id], ng)
	}
	return collapsedGroups
}

func (c *DefaultKarpenterConverter) serializeCustomLabels(labels map[string]string, customLabelKeys sets.Set[string]) string {
	var parts []string
	for _, k := range sets.List(customLabelKeys) {
		if v, ok := labels[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return strings.Join(parts, ",")
}

func (c *DefaultKarpenterConverter) buildNodePoolRequirements(instanceTypes []*karpentercloudprovider.InstanceType) []karpenterv1.NodeSelectorRequirementWithMinValues {
	combinedValues := make(map[string]sets.Set[string])
	for _, it := range instanceTypes {
		for _, req := range it.Requirements.Values() {
			if karpenterv1.RestrictedLabels.Has(req.Key) {
				continue
			}
			if _, ok := combinedValues[req.Key]; !ok {
				combinedValues[req.Key] = sets.New[string]()
			}
			combinedValues[req.Key].Insert(req.Values()...)
		}
		for _, off := range it.Offerings {
			for _, req := range off.Requirements.Values() {
				if karpenterv1.RestrictedLabels.Has(req.Key) {
					continue
				}
				if _, ok := combinedValues[req.Key]; !ok {
					combinedValues[req.Key] = sets.New[string]()
				}
				combinedValues[req.Key].Insert(req.Values()...)
			}
		}
	}
	combinedRequirements := karpenterscheduling.NewRequirements()
	for _, key := range sets.List(sets.KeySet(combinedValues)) {
		values := combinedValues[key]
		valList := sets.List(values)
		if len(valList) == 0 {
			continue
		}
		if karpenterv1.WellKnownLabels.Has(key) {
			combinedRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpIn, valList...))
		} else {
			combinedRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpNotIn, caIgnoreReservedValue))
		}
	}
	return combinedRequirements.NodeSelectorRequirements()
}

func serializeResources(res apiv1.ResourceList) string {
	if res == nil {
		return "c:0,m:0"
	}
	return fmt.Sprintf("c:%d,m:%d", res.Cpu().MilliValue(), res.Memory().Value())
}

func (c *DefaultKarpenterConverter) CreateInstanceType(name string, physicalITName string, ngs []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo, nodePoolName string, customLabelKeys sets.Set[string], offeringMap map[string]*karpentercloudprovider.Offering) *karpentercloudprovider.InstanceType {
	if len(ngs) == 0 {
		return nil
	}
	representativeNg := ngs[0]
	representativeNi := nodeInfos[representativeNg.Id()]
	if representativeNi == nil || representativeNi.Node() == nil {
		return nil
	}
	node := representativeNi.Node()

	// Collect homogeneous labels across all node groups in this IT
	homogeneousLabels := make(map[string]string)
	if len(ngs) > 0 {
		firstNi := nodeInfos[ngs[0].Id()]
		if firstNi != nil && firstNi.Node() != nil {
			for k, v := range firstNi.Node().Labels {
				if c.isIgnoredLabel(k) {
					continue
				}
				homogeneousLabels[k] = v
			}
		}
		for _, ng := range ngs[1:] {
			ni := nodeInfos[ng.Id()]
			if ni == nil || ni.Node() == nil {
				continue
			}
			for k, v := range homogeneousLabels {
				if val, ok := ni.Node().Labels[k]; !ok || val != v {
					delete(homogeneousLabels, k)
				}
			}
		}
	}

	labels := make(map[string]string)
	for k, v := range homogeneousLabels {
		labels[k] = v
	}
	if _, ok := labels[apiv1.LabelArchStable]; !ok {
		if val, betaOk := node.Labels["beta.kubernetes.io/arch"]; betaOk {
			labels[apiv1.LabelArchStable] = val
		} else {
			labels[apiv1.LabelArchStable] = karpenterv1.ArchitectureAmd64
		}
	}
	if _, ok := labels[apiv1.LabelOSStable]; !ok {
		if val, betaOk := node.Labels["beta.kubernetes.io/os"]; betaOk {
			labels[apiv1.LabelOSStable] = val
		} else {
			labels[apiv1.LabelOSStable] = "linux"
		}
	}
	labels[apiv1.LabelInstanceTypeStable] = name
	labels[karpenterv1.NodePoolLabelKey] = nodePoolName
	labels["ca.prototype/prototypenodeclass"] = "default"

	requirements := karpenterscheduling.NewLabelRequirements(labels)

	var offerings karpentercloudprovider.Offerings
	for _, ng := range ngs {
		ni := nodeInfos[ng.Id()]
		if ni == nil || ni.Node() == nil {
			continue
		}
		n := ni.Node()
		zone := n.Labels[apiv1.LabelTopologyZone]
		if zone == "" {
			if betaZone, ok := n.Labels[apiv1.LabelZoneFailureDomain]; ok && betaZone != "" {
				zone = betaZone
			} else {
				zone = "default-zone"
			}
		}
		capType := n.Labels[karpenterv1.CapacityTypeLabelKey]
		if capType == "" {
			capType = karpenterv1.CapacityTypeOnDemand
		}

		offeringRequirements := karpenterscheduling.NewRequirements(
			karpenterscheduling.NewRequirement(karpenterv1.CapacityTypeLabelKey, apiv1.NodeSelectorOpIn, capType),
		)
		if zone != "" && zone != "default-zone" {
			offeringRequirements.Add(karpenterscheduling.NewRequirement(apiv1.LabelTopologyZone, apiv1.NodeSelectorOpIn, zone))
		}
		for _, key := range sets.List(customLabelKeys) {
			if val, ok := n.Labels[key]; ok {
				offeringRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpIn, val))
			} else {
				offeringRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpDoesNotExist))
			}
		}

		var cpu, memGiB, gpu float64
		if n.Status.Capacity != nil {
			cpu = float64(n.Status.Capacity.Cpu().MilliValue()) / 1000.0
			memGiB = float64(n.Status.Capacity.Memory().Value()) / (1024.0 * 1024.0 * 1024.0)
			for k, v := range n.Status.Capacity {
				if strings.Contains(string(k), "gpu") {
					gpu += float64(v.Value())
				}
			}
		}

		price := (fallbackCPUCoefficient * cpu) + (fallbackMemCoefficient * memGiB) + (fallbackGPUCoefficient * gpu)
		if capType == karpenterv1.CapacityTypeSpot {
			price = price * fallbackSpotDiscount
		}
		if c.PricingModel != nil {
			if p, err := c.PricingModel.NodePrice(n, time.Now(), time.Now().Add(time.Hour)); err == nil {
				price = p
			}
		}

		offering := &karpentercloudprovider.Offering{
			Available:    true,
			Price:        price,
			Requirements: offeringRequirements,
		}
		offerings = append(offerings, offering)
		offeringMap[ng.Id()] = offering
	}

	var capacity apiv1.ResourceList
	if node.Status.Capacity != nil {
		capacity = node.Status.Capacity.DeepCopy()
	} else {
		capacity = make(apiv1.ResourceList)
	}

	if representativeNi.CSINode != nil && representativeNi.CSINode.Spec.Drivers != nil {
		for _, driver := range representativeNi.CSINode.Spec.Drivers {
			if driver.Allocatable != nil && driver.Allocatable.Count != nil {
				capacity[apiv1.ResourceName(driver.Name)] = *resource.NewQuantity(int64(*driver.Allocatable.Count), resource.DecimalSI)
			}
		}
	}

	var kubeReserved apiv1.ResourceList
	if node.Status.Capacity != nil && node.Status.Allocatable != nil {
		kubeReserved = karpenterresources.Subtract(node.Status.Capacity, node.Status.Allocatable)
	} else {
		kubeReserved = apiv1.ResourceList{}
	}

	return &karpentercloudprovider.InstanceType{
		Name:         name,
		Requirements: requirements,
		Offerings:    offerings,
		Capacity:     capacity,
		Overhead: &karpentercloudprovider.InstanceTypeOverhead{
			KubeReserved:   kubeReserved,
			SystemReserved: apiv1.ResourceList{},
		},
	}
}

func getCompatibleDSKeys(ng cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) string {
	ni, found := nodeInfos[ng.Id()]
	if !found || ni == nil {
		return ""
	}
	var dsKeys []string
	for _, pi := range ni.Pods() {
		if pi.Pod != nil && podutils.IsDaemonSetPod(pi.Pod) {
			dsName := pi.Pod.Name
			if controllerRef := metav1.GetControllerOf(pi.Pod); controllerRef != nil {
				dsName = controllerRef.Name
			}
			dsKey := fmt.Sprintf("%s/%s", pi.Pod.Namespace, dsName)
			dsKeys = append(dsKeys, dsKey)
		}
	}
	sort.Strings(dsKeys)
	return strings.Join(dsKeys, ",")
}
