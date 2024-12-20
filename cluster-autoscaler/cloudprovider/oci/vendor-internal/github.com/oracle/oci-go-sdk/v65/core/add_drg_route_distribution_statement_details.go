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

// AddDrgRouteDistributionStatementDetails Details used to add a route distribution statement.
type AddDrgRouteDistributionStatementDetails struct {

	// The action is applied only if all of the match criteria is met.
	MatchCriteria []DrgRouteDistributionMatchCriteria `mandatory:"true" json:"matchCriteria"`

	// Accept: import/export the route "as is"
	Action AddDrgRouteDistributionStatementDetailsActionEnum `mandatory:"true" json:"action"`

	// This field is used to specify the priority of each statement in a route distribution.
	// The priority will be represented as a number between 0 and 65535 where a lower number
	// indicates a higher priority. When a route is processed, statements are applied in the order
	// defined by their priority. The first matching rule dictates the action that will be taken
	// on the route.
	Priority *int `mandatory:"true" json:"priority"`
}

func (m AddDrgRouteDistributionStatementDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AddDrgRouteDistributionStatementDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingAddDrgRouteDistributionStatementDetailsActionEnum(string(m.Action)); !ok && m.Action != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Action: %s. Supported values are: %s.", m.Action, strings.Join(GetAddDrgRouteDistributionStatementDetailsActionEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *AddDrgRouteDistributionStatementDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		MatchCriteria []drgroutedistributionmatchcriteria               `json:"matchCriteria"`
		Action        AddDrgRouteDistributionStatementDetailsActionEnum `json:"action"`
		Priority      *int                                              `json:"priority"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.MatchCriteria = make([]DrgRouteDistributionMatchCriteria, len(model.MatchCriteria))
	for i, n := range model.MatchCriteria {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.MatchCriteria[i] = nn.(DrgRouteDistributionMatchCriteria)
		} else {
			m.MatchCriteria[i] = nil
		}
	}
	m.Action = model.Action

	m.Priority = model.Priority

	return
}

// AddDrgRouteDistributionStatementDetailsActionEnum Enum with underlying type: string
type AddDrgRouteDistributionStatementDetailsActionEnum string

// Set of constants representing the allowable values for AddDrgRouteDistributionStatementDetailsActionEnum
const (
	AddDrgRouteDistributionStatementDetailsActionAccept AddDrgRouteDistributionStatementDetailsActionEnum = "ACCEPT"
)

var mappingAddDrgRouteDistributionStatementDetailsActionEnum = map[string]AddDrgRouteDistributionStatementDetailsActionEnum{
	"ACCEPT": AddDrgRouteDistributionStatementDetailsActionAccept,
}

var mappingAddDrgRouteDistributionStatementDetailsActionEnumLowerCase = map[string]AddDrgRouteDistributionStatementDetailsActionEnum{
	"accept": AddDrgRouteDistributionStatementDetailsActionAccept,
}

// GetAddDrgRouteDistributionStatementDetailsActionEnumValues Enumerates the set of values for AddDrgRouteDistributionStatementDetailsActionEnum
func GetAddDrgRouteDistributionStatementDetailsActionEnumValues() []AddDrgRouteDistributionStatementDetailsActionEnum {
	values := make([]AddDrgRouteDistributionStatementDetailsActionEnum, 0)
	for _, v := range mappingAddDrgRouteDistributionStatementDetailsActionEnum {
		values = append(values, v)
	}
	return values
}

// GetAddDrgRouteDistributionStatementDetailsActionEnumStringValues Enumerates the set of values in String for AddDrgRouteDistributionStatementDetailsActionEnum
func GetAddDrgRouteDistributionStatementDetailsActionEnumStringValues() []string {
	return []string{
		"ACCEPT",
	}
}

// GetMappingAddDrgRouteDistributionStatementDetailsActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddDrgRouteDistributionStatementDetailsActionEnum(val string) (AddDrgRouteDistributionStatementDetailsActionEnum, bool) {
	enum, ok := mappingAddDrgRouteDistributionStatementDetailsActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
