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
		AvailabilityDomain: n.Labels[apiv1.LabelZoneFailureDomain],
		CompartmentID:      n.Annotations[ociAnnotationCompartmentID],
		InstanceID:         n.Annotations[ociInstanceIDAnnotation],
		PoolID:             n.Annotations[ociInstancePoolIDAnnotation],
		PrivateIPAddress:   getNodeInternalAddress(n),
		PublicIPAddress:    getNodeExternalAddress(n),
		Shape:              n.Labels[apiv1.LabelInstanceType],
	}, nil
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
