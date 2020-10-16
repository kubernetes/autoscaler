/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Request Object
type ListResizeFlavorsRequest struct {
	InstanceUuid     *string                          `json:"instance_uuid,omitempty"`
	Limit            *int32                           `json:"limit,omitempty"`
	Marker           *string                          `json:"marker,omitempty"`
	SortDir          *ListResizeFlavorsRequestSortDir `json:"sort_dir,omitempty"`
	SortKey          *ListResizeFlavorsRequestSortKey `json:"sort_key,omitempty"`
	SourceFlavorId   *string                          `json:"source_flavor_id,omitempty"`
	SourceFlavorName *string                          `json:"source_flavor_name,omitempty"`
}

func (o ListResizeFlavorsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListResizeFlavorsRequest", string(data)}, " ")
}

type ListResizeFlavorsRequestSortDir struct {
	value string
}

type ListResizeFlavorsRequestSortDirEnum struct {
	ASC  ListResizeFlavorsRequestSortDir
	DESC ListResizeFlavorsRequestSortDir
}

func GetListResizeFlavorsRequestSortDirEnum() ListResizeFlavorsRequestSortDirEnum {
	return ListResizeFlavorsRequestSortDirEnum{
		ASC: ListResizeFlavorsRequestSortDir{
			value: "asc",
		},
		DESC: ListResizeFlavorsRequestSortDir{
			value: "desc",
		},
	}
}

func (c ListResizeFlavorsRequestSortDir) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListResizeFlavorsRequestSortDir) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

type ListResizeFlavorsRequestSortKey struct {
	value string
}

type ListResizeFlavorsRequestSortKeyEnum struct {
	FLAVORID  ListResizeFlavorsRequestSortKey
	SORT_KEY  ListResizeFlavorsRequestSortKey
	NAME      ListResizeFlavorsRequestSortKey
	MEMORY_MB ListResizeFlavorsRequestSortKey
	VCPUS     ListResizeFlavorsRequestSortKey
	ROOT_GB   ListResizeFlavorsRequestSortKey
}

func GetListResizeFlavorsRequestSortKeyEnum() ListResizeFlavorsRequestSortKeyEnum {
	return ListResizeFlavorsRequestSortKeyEnum{
		FLAVORID: ListResizeFlavorsRequestSortKey{
			value: "flavorid",
		},
		SORT_KEY: ListResizeFlavorsRequestSortKey{
			value: "sort_key",
		},
		NAME: ListResizeFlavorsRequestSortKey{
			value: "name",
		},
		MEMORY_MB: ListResizeFlavorsRequestSortKey{
			value: "memory_mb",
		},
		VCPUS: ListResizeFlavorsRequestSortKey{
			value: "vcpus",
		},
		ROOT_GB: ListResizeFlavorsRequestSortKey{
			value: "root_gb",
		},
	}
}

func (c ListResizeFlavorsRequestSortKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListResizeFlavorsRequestSortKey) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
