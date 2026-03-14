/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools/consts"
	"k8s.io/klog/v2"

	"strings"
)

// OciRef contains s reference to some entity in OCI world.
type OciRef struct {
	AvailabilityDomain string
	Name               string
	CompartmentID      string
	InstanceID         string
	NodePoolID         string
	InstancePoolID     string
	PrivateIPAddress   string
	PublicIPAddress    string
	Shape              string
	IsNodeSelfManaged  bool
}

// NodeToOciRef converts a node object into an oci reference
func NodeToOciRef(n *apiv1.Node) (OciRef, error) {

	return OciRef{
		Name:               n.ObjectMeta.Name,
		AvailabilityDomain: getNodeAZ(n),
		CompartmentID:      n.Annotations[consts.OciAnnotationCompartmentID],
		InstanceID:         getNodeInstanceID(n),
		NodePoolID:         n.Annotations["oci.oraclecloud.com/node-pool-id"],
		InstancePoolID:     getNodeInstancePoolID(n),
		PrivateIPAddress:   getNodeInternalAddress(n),
		PublicIPAddress:    getNodeExternalAddress(n),
		Shape:              getNodeShape(n),
		IsNodeSelfManaged:  n.Labels["oci.oraclecloud.com/node.info.byon"] == "true",
	}, nil
}

// getNodeShape returns the shape of the node instance if set as a label or an empty string if is not found.
func getNodeShape(node *apiv1.Node) string {
	// First check for the deprecated label
	if shape, ok := node.Labels[apiv1.LabelInstanceType]; ok {
		klog.V(5).Infof("Extracting node shape %s from label %s", shape, apiv1.LabelInstanceType)
		return shape
	} else if shape, ok := node.Labels[apiv1.LabelInstanceTypeStable]; ok {
		klog.V(5).Infof("Extracting node shape %s from label %s", shape, apiv1.LabelInstanceTypeStable)
		return shape
	}
	return ""
}

// getNodeAZ returns the availability domain/zone of the node instance if set as a label or an empty string if is not found.
func getNodeAZ(node *apiv1.Node) string {
	// First check for the deprecated label
	if az, ok := node.Labels[apiv1.LabelZoneFailureDomain]; ok {
		klog.V(5).Infof("Extracting availability domain %s from label %s", az, apiv1.LabelZoneFailureDomain)
		return az
	} else if az, ok := node.Labels[apiv1.LabelTopologyZone]; ok {
		klog.V(5).Infof("Extracting availability domain %s from label %s", az, apiv1.LabelTopologyZone)
		return az
	}
	return ""
}

// getNodeInternalAddress returns the first private address of the node and an empty string if none are found.
func getNodeInternalAddress(node *apiv1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeInternalIP {
			klog.V(5).Infof("Extracting node internal IP %s from node %s", address.Address, node.Name)
			return address.Address
		}
	}
	return ""
}

// getNodeExternalAddress returns the first public address of the node and an empty string if none are found.
func getNodeExternalAddress(node *apiv1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeExternalIP {
			klog.V(5).Infof("Extracting node external IP %s from node %s", address.Address, node.Name)
			return address.Address
		}
	}
	return ""
}

// getNodeInstancePoolID returns the instance pool ID if set as a label or annotation or an empty string if is not found.
func getNodeInstancePoolID(node *apiv1.Node) string {

	// Handle unfilled instance placeholder (instances that have yet to be created)
	if strings.Contains(node.Name, consts.InstanceIDUnfulfilled) {
		instIndex := strings.LastIndex(node.Name, "-")
		return strings.Replace(node.Name[:instIndex], consts.InstanceIDUnfulfilled, "", 1)
	}

	poolIDPrefixLabel, _ := node.Labels[consts.InstancePoolIDLabelPrefix]
	poolIDSuffixLabel, _ := node.Labels[consts.InstancePoolIDLabelSuffix]

	if poolIDPrefixLabel != "" && poolIDSuffixLabel != "" {
		klog.V(5).Infof("Extracting instance-pool %s from labels %s + %s", poolIDPrefixLabel+"."+poolIDSuffixLabel, consts.InstancePoolIDLabelPrefix, consts.InstancePoolIDLabelSuffix)
		return poolIDPrefixLabel + "." + poolIDSuffixLabel
	}

	poolIDAnnotation, _ := node.Annotations[consts.OciInstancePoolIDAnnotation]
	klog.V(5).Infof("Extracting instance-pool %s from annotation %s", poolIDAnnotation, consts.OciInstanceIDAnnotation)
	return poolIDAnnotation
}

// getNodeInstanceID returns the instance ID if set as a label or annotation or an empty string if is not found.
func getNodeInstanceID(node *apiv1.Node) string {
	providerID := strings.TrimPrefix(node.Spec.ProviderID, "oci://")
	if len(providerID) != 0 {
		klog.V(5).Infof("Extracting instance-id %s from .spec.providerID", providerID)
		return providerID
	}

	// Handle unfilled instance placeholder (instances that have yet to be created)
	if strings.Contains(node.Name, consts.InstanceIDUnfulfilled) {
		return node.Name
	}

	instancePrefixLabel, _ := node.Labels[consts.InstanceIDLabelPrefix]
	instanceSuffixLabel, _ := node.Labels[consts.InstanceIDLabelSuffix]

	if instancePrefixLabel != "" && instanceSuffixLabel != "" {
		klog.V(5).Infof("Extracting instance-id %s from labels %s + %s", instancePrefixLabel+"."+instanceSuffixLabel, consts.InstanceIDLabelPrefix, consts.InstanceIDLabelSuffix)
		return instancePrefixLabel + "." + instanceSuffixLabel
	}

	instanceIDAnnotation, _ := node.Annotations[consts.OciInstanceIDAnnotation]
	klog.V(5).Infof("Extracting instance-id %s from annotation %s", instanceIDAnnotation, consts.OciInstanceIDAnnotation)
	return instanceIDAnnotation
}
