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

// DrgRouteDistributionStatement A single statement within a route distribution. All match criteria in a statement must be met
// for the action to take place.
type DrgRouteDistributionStatement struct {

	// The action is applied only if all of the match criteria is met.
	// If there are no match criteria in a statement, any input is considered a match and the action is applied.
	MatchCriteria []DrgRouteDistributionMatchCriteria `mandatory:"true" json:"matchCriteria"`

	// `ACCEPT` indicates the route should be imported or exported as-is.
	Action DrgRouteDistributionStatementActionEnum `mandatory:"true" json:"action"`

	// This field specifies the priority of each statement in a route distribution.
	// Priorities must be unique within a particular route distribution.
	// The priority will be represented as a number between 0 and 65535 where a lower number
	// indicates a higher priority. When a route is processed, statements are applied in the order
	// defined by their priority. The first matching rule dictates the action that will be taken
	// on the route.
	Priority *int `mandatory:"true" json:"priority"`

	// The Oracle-assigned ID of the route distribution statement.
	Id *string `mandatory:"true" json:"id"`
}

func (m DrgRouteDistributionStatement) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgRouteDistributionStatement) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgRouteDistributionStatementActionEnum(string(m.Action)); !ok && m.Action != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Action: %s. Supported values are: %s.", m.Action, strings.Join(GetDrgRouteDistributionStatementActionEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *DrgRouteDistributionStatement) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		MatchCriteria []drgroutedistributionmatchcriteria     `json:"matchCriteria"`
		Action        DrgRouteDistributionStatementActionEnum `json:"action"`
		Priority      *int                                    `json:"priority"`
		Id            *string                                 `json:"id"`
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

	m.Id = model.Id

	return
}

// DrgRouteDistributionStatementActionEnum Enum with underlying type: string
type DrgRouteDistributionStatementActionEnum string

// Set of constants representing the allowable values for DrgRouteDistributionStatementActionEnum
const (
	DrgRouteDistributionStatementActionAccept DrgRouteDistributionStatementActionEnum = "ACCEPT"
)

var mappingDrgRouteDistributionStatementActionEnum = map[string]DrgRouteDistributionStatementActionEnum{
	"ACCEPT": DrgRouteDistributionStatementActionAccept,
}

var mappingDrgRouteDistributionStatementActionEnumLowerCase = map[string]DrgRouteDistributionStatementActionEnum{
	"accept": DrgRouteDistributionStatementActionAccept,
}

// GetDrgRouteDistributionStatementActionEnumValues Enumerates the set of values for DrgRouteDistributionStatementActionEnum
func GetDrgRouteDistributionStatementActionEnumValues() []DrgRouteDistributionStatementActionEnum {
	values := make([]DrgRouteDistributionStatementActionEnum, 0)
	for _, v := range mappingDrgRouteDistributionStatementActionEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteDistributionStatementActionEnumStringValues Enumerates the set of values in String for DrgRouteDistributionStatementActionEnum
func GetDrgRouteDistributionStatementActionEnumStringValues() []string {
	return []string{
		"ACCEPT",
	}
}

// GetMappingDrgRouteDistributionStatementActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteDistributionStatementActionEnum(val string) (DrgRouteDistributionStatementActionEnum, bool) {
	enum, ok := mappingDrgRouteDistributionStatementActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
