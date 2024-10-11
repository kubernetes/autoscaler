// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// NodePoolOptions Options for creating or updating node pools.
type NodePoolOptions struct {

	// Available Kubernetes versions.
	KubernetesVersions []string `mandatory:"false" json:"kubernetesVersions"`

	// Available shapes for nodes.
	Shapes []string `mandatory:"false" json:"shapes"`

	// Deprecated. See sources.
	// When creating a node pool using the `CreateNodePoolDetails` object, only image names contained in this
	// property can be passed to the `nodeImageName` property.
	Images []string `mandatory:"false" json:"images"`

	// Available source of the node.
	Sources []NodeSourceOption `mandatory:"false" json:"sources"`
}

func (m NodePoolOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NodePoolOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *NodePoolOptions) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		KubernetesVersions []string           `json:"kubernetesVersions"`
		Shapes             []string           `json:"shapes"`
		Images             []string           `json:"images"`
		Sources            []nodesourceoption `json:"sources"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.KubernetesVersions = make([]string, len(model.KubernetesVersions))
	copy(m.KubernetesVersions, model.KubernetesVersions)
	m.Shapes = make([]string, len(model.Shapes))
	copy(m.Shapes, model.Shapes)
	m.Images = make([]string, len(model.Images))
	copy(m.Images, model.Images)
	m.Sources = make([]NodeSourceOption, len(model.Sources))
	for i, n := range model.Sources {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.Sources[i] = nn.(NodeSourceOption)
		} else {
			m.Sources[i] = nil
		}
	}
	return
}
