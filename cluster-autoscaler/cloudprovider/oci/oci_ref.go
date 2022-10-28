/*
Copyright 2021 Oracle and/or its affiliates.

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

package oci

import (
	apiv1 "k8s.io/api/core/v1"
	"strings"
)

// OciRef contains s reference to some entity in OCI world.
type OciRef struct {
	AvailabilityDomain string
	Name               string
	CompartmentID      string
	InstanceID         string
	PoolID             string
	PrivateIPAddress   string
	PublicIPAddress    string
	Shape              string
}

func nodeToOciRef(n *apiv1.Node) (OciRef, error) {

	return OciRef{
		Name:               n.ObjectMeta.Name,
		AvailabilityDomain: getNodeAZ(n),
		CompartmentID:      n.Annotations[ociAnnotationCompartmentID],
		InstanceID:         getNodeInstanceID(n),
		PoolID:             getNodeInstancePoolID(n),
		PrivateIPAddress:   getNodeInternalAddress(n),
		PublicIPAddress:    getNodeExternalAddress(n),
		Shape:              getNodeShape(n),
	}, nil
}

// getNodeShape returns the shape of the node instance if set as a label or an empty string if is not found.
func getNodeShape(node *apiv1.Node) string {
	// First check for the deprecated label
	if shape, ok := node.Labels[apiv1.LabelInstanceType]; ok {
		return shape
	} else if shape, ok := node.Labels[apiv1.LabelInstanceTypeStable]; ok {
		return shape
	}
	return ""
}

// getNodeAZ returns the availability domain/zone of the node instance if set as a label or an empty string if is not found.
func getNodeAZ(node *apiv1.Node) string {
	// First check for the deprecated label
	if az, ok := node.Labels[apiv1.LabelZoneFailureDomain]; ok {
		return az
	} else if az, ok := node.Labels[apiv1.LabelTopologyZone]; ok {
		return az
	}
	return ""
}

// getNodeInternalAddress returns the first private address of the node and an empty string if none are found.
func getNodeInternalAddress(node *apiv1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

// getNodeExternalAddress returns the first public address of the node and an empty string if none are found.
func getNodeExternalAddress(node *apiv1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeExternalIP {
			return address.Address
		}
	}
	return ""
}

// getNodeInstancePoolID returns the instance pool ID if set as a label or annotation or an empty string if is not found.
func getNodeInstancePoolID(node *apiv1.Node) string {

	// Handle unfilled instance placeholder (instances that have yet to be created)
	if strings.Contains(node.Name, instanceIDUnfulfilled) {
		instIndex := strings.LastIndex(node.Name, "-")
		return strings.Replace(node.Name[:instIndex], instanceIDUnfulfilled, "", 1)
	}

	poolIDPrefixLabel, _ := node.Labels[instancePoolIDLabelPrefix]
	poolIDSuffixLabel, _ := node.Labels[instancePoolIDLabelSuffix]

	if poolIDPrefixLabel != "" && poolIDSuffixLabel != "" {
		return poolIDPrefixLabel + "." + poolIDSuffixLabel
	}

	poolIDAnnotation, _ := node.Annotations[ociInstancePoolIDAnnotation]
	return poolIDAnnotation
}

// getNodeInstanceID returns the instance ID if set as a label or annotation or an empty string if is not found.
func getNodeInstanceID(node *apiv1.Node) string {

	// Handle unfilled instance placeholder (instances that have yet to be created)
	if strings.Contains(node.Name, instanceIDUnfulfilled) {
		return node.Name
	}

	instancePrefixLabel, _ := node.Labels[instanceIDLabelPrefix]
	instanceSuffixLabel, _ := node.Labels[instanceIDLabelSuffix]

	if instancePrefixLabel != "" && instanceSuffixLabel != "" {
		return instancePrefixLabel + "." + instanceSuffixLabel
	}

	instanceIDAnnotation, _ := node.Annotations[ociInstanceIDAnnotation]
	return instanceIDAnnotation
}
