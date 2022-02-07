// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// DrgRouteDistributionMatchCriteria The match criteria in a route distribution statement. The match criteria outlines which routes
// should be imported or exported. Leaving the match criteria empty implies match ALL.
type DrgRouteDistributionMatchCriteria interface {
}

type drgroutedistributionmatchcriteria struct {
	JsonData  []byte
	MatchType string `json:"matchType"`
}

// UnmarshalJSON unmarshals json
func (m *drgroutedistributionmatchcriteria) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerdrgroutedistributionmatchcriteria drgroutedistributionmatchcriteria
	s := struct {
		Model Unmarshalerdrgroutedistributionmatchcriteria
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.MatchType = s.Model.MatchType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *drgroutedistributionmatchcriteria) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.MatchType {
	case "DRG_ATTACHMENT_ID":
		mm := DrgAttachmentIdDrgRouteDistributionMatchCriteria{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "DRG_ATTACHMENT_TYPE":
		mm := DrgAttachmentTypeDrgRouteDistributionMatchCriteria{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

func (m drgroutedistributionmatchcriteria) String() string {
	return common.PointerString(m)
}

// DrgRouteDistributionMatchCriteriaMatchTypeEnum Enum with underlying type: string
type DrgRouteDistributionMatchCriteriaMatchTypeEnum string

// Set of constants representing the allowable values for DrgRouteDistributionMatchCriteriaMatchTypeEnum
const (
	DrgRouteDistributionMatchCriteriaMatchTypeType DrgRouteDistributionMatchCriteriaMatchTypeEnum = "DRG_ATTACHMENT_TYPE"
	DrgRouteDistributionMatchCriteriaMatchTypeId   DrgRouteDistributionMatchCriteriaMatchTypeEnum = "DRG_ATTACHMENT_ID"
)

var mappingDrgRouteDistributionMatchCriteriaMatchType = map[string]DrgRouteDistributionMatchCriteriaMatchTypeEnum{
	"DRG_ATTACHMENT_TYPE": DrgRouteDistributionMatchCriteriaMatchTypeType,
	"DRG_ATTACHMENT_ID":   DrgRouteDistributionMatchCriteriaMatchTypeId,
}

// GetDrgRouteDistributionMatchCriteriaMatchTypeEnumValues Enumerates the set of values for DrgRouteDistributionMatchCriteriaMatchTypeEnum
func GetDrgRouteDistributionMatchCriteriaMatchTypeEnumValues() []DrgRouteDistributionMatchCriteriaMatchTypeEnum {
	values := make([]DrgRouteDistributionMatchCriteriaMatchTypeEnum, 0)
	for _, v := range mappingDrgRouteDistributionMatchCriteriaMatchType {
		values = append(values, v)
	}
	return values
}
