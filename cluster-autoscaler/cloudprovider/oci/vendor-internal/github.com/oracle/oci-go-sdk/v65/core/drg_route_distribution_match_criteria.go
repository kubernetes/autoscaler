// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DrgRouteDistributionMatchCriteria The match criteria in a route distribution statement. The match criteria outlines which routes
// should be imported or exported.
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
	case "MATCH_ALL":
		mm := DrgAttachmentMatchAllDrgRouteDistributionMatchCriteria{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for DrgRouteDistributionMatchCriteria: %s.", m.MatchType)
		return *m, nil
	}
}

func (m drgroutedistributionmatchcriteria) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m drgroutedistributionmatchcriteria) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgRouteDistributionMatchCriteriaMatchTypeEnum Enum with underlying type: string
type DrgRouteDistributionMatchCriteriaMatchTypeEnum string

// Set of constants representing the allowable values for DrgRouteDistributionMatchCriteriaMatchTypeEnum
const (
	DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentType DrgRouteDistributionMatchCriteriaMatchTypeEnum = "DRG_ATTACHMENT_TYPE"
	DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentId   DrgRouteDistributionMatchCriteriaMatchTypeEnum = "DRG_ATTACHMENT_ID"
	DrgRouteDistributionMatchCriteriaMatchTypeMatchAll          DrgRouteDistributionMatchCriteriaMatchTypeEnum = "MATCH_ALL"
)

var mappingDrgRouteDistributionMatchCriteriaMatchTypeEnum = map[string]DrgRouteDistributionMatchCriteriaMatchTypeEnum{
	"DRG_ATTACHMENT_TYPE": DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentType,
	"DRG_ATTACHMENT_ID":   DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentId,
	"MATCH_ALL":           DrgRouteDistributionMatchCriteriaMatchTypeMatchAll,
}

var mappingDrgRouteDistributionMatchCriteriaMatchTypeEnumLowerCase = map[string]DrgRouteDistributionMatchCriteriaMatchTypeEnum{
	"drg_attachment_type": DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentType,
	"drg_attachment_id":   DrgRouteDistributionMatchCriteriaMatchTypeDrgAttachmentId,
	"match_all":           DrgRouteDistributionMatchCriteriaMatchTypeMatchAll,
}

// GetDrgRouteDistributionMatchCriteriaMatchTypeEnumValues Enumerates the set of values for DrgRouteDistributionMatchCriteriaMatchTypeEnum
func GetDrgRouteDistributionMatchCriteriaMatchTypeEnumValues() []DrgRouteDistributionMatchCriteriaMatchTypeEnum {
	values := make([]DrgRouteDistributionMatchCriteriaMatchTypeEnum, 0)
	for _, v := range mappingDrgRouteDistributionMatchCriteriaMatchTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteDistributionMatchCriteriaMatchTypeEnumStringValues Enumerates the set of values in String for DrgRouteDistributionMatchCriteriaMatchTypeEnum
func GetDrgRouteDistributionMatchCriteriaMatchTypeEnumStringValues() []string {
	return []string{
		"DRG_ATTACHMENT_TYPE",
		"DRG_ATTACHMENT_ID",
		"MATCH_ALL",
	}
}

// GetMappingDrgRouteDistributionMatchCriteriaMatchTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteDistributionMatchCriteriaMatchTypeEnum(val string) (DrgRouteDistributionMatchCriteriaMatchTypeEnum, bool) {
	enum, ok := mappingDrgRouteDistributionMatchCriteriaMatchTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
