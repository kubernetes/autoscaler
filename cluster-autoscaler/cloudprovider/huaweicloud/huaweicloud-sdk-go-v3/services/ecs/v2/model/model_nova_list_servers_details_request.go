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
type NovaListServersDetailsRequest struct {
	ChangesSince        *string                               `json:"changes-since,omitempty"`
	Flavor              *string                               `json:"flavor,omitempty"`
	Image               *string                               `json:"image,omitempty"`
	Ip                  *string                               `json:"ip,omitempty"`
	Limit               *int32                                `json:"limit,omitempty"`
	Marker              *string                               `json:"marker,omitempty"`
	Name                *string                               `json:"name,omitempty"`
	NotTags             *string                               `json:"not-tags,omitempty"`
	ReservationId       *string                               `json:"reservation_id,omitempty"`
	SortKey             *NovaListServersDetailsRequestSortKey `json:"sort_key,omitempty"`
	Status              *NovaListServersDetailsRequestStatus  `json:"status,omitempty"`
	Tags                *string                               `json:"tags,omitempty"`
	OpenStackAPIVersion *string                               `json:"OpenStack-API-Version,omitempty"`
}

func (o NovaListServersDetailsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaListServersDetailsRequest", string(data)}, " ")
}

type NovaListServersDetailsRequestSortKey struct {
	value string
}

type NovaListServersDetailsRequestSortKeyEnum struct {
	CREATED_AT        NovaListServersDetailsRequestSortKey
	AVAILABILITY_ZONE NovaListServersDetailsRequestSortKey
	DISPLAY_NAME      NovaListServersDetailsRequestSortKey
	HOST              NovaListServersDetailsRequestSortKey
	INSTANCE_TYPE_ID  NovaListServersDetailsRequestSortKey
	KEY_NAME          NovaListServersDetailsRequestSortKey
	PROJECT_ID        NovaListServersDetailsRequestSortKey
	USER_ID           NovaListServersDetailsRequestSortKey
	UPDATED_AT        NovaListServersDetailsRequestSortKey
	UUID              NovaListServersDetailsRequestSortKey
	VM_STATE          NovaListServersDetailsRequestSortKey
}

func GetNovaListServersDetailsRequestSortKeyEnum() NovaListServersDetailsRequestSortKeyEnum {
	return NovaListServersDetailsRequestSortKeyEnum{
		CREATED_AT: NovaListServersDetailsRequestSortKey{
			value: "created_at",
		},
		AVAILABILITY_ZONE: NovaListServersDetailsRequestSortKey{
			value: "availability_zone",
		},
		DISPLAY_NAME: NovaListServersDetailsRequestSortKey{
			value: "display_name",
		},
		HOST: NovaListServersDetailsRequestSortKey{
			value: "host",
		},
		INSTANCE_TYPE_ID: NovaListServersDetailsRequestSortKey{
			value: "instance_type_id",
		},
		KEY_NAME: NovaListServersDetailsRequestSortKey{
			value: "key_name",
		},
		PROJECT_ID: NovaListServersDetailsRequestSortKey{
			value: "project_id",
		},
		USER_ID: NovaListServersDetailsRequestSortKey{
			value: "user_id",
		},
		UPDATED_AT: NovaListServersDetailsRequestSortKey{
			value: "updated_at",
		},
		UUID: NovaListServersDetailsRequestSortKey{
			value: "uuid",
		},
		VM_STATE: NovaListServersDetailsRequestSortKey{
			value: "vm_state",
		},
	}
}

func (c NovaListServersDetailsRequestSortKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NovaListServersDetailsRequestSortKey) UnmarshalJSON(b []byte) error {
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

type NovaListServersDetailsRequestStatus struct {
	value string
}

type NovaListServersDetailsRequestStatusEnum struct {
	ACTIVE            NovaListServersDetailsRequestStatus
	BUILD             NovaListServersDetailsRequestStatus
	DELETED           NovaListServersDetailsRequestStatus
	ERROR             NovaListServersDetailsRequestStatus
	HARD_REBOOT       NovaListServersDetailsRequestStatus
	MIGRATING         NovaListServersDetailsRequestStatus
	REBOOT            NovaListServersDetailsRequestStatus
	RESIZE            NovaListServersDetailsRequestStatus
	REVERT_RESIZE     NovaListServersDetailsRequestStatus
	SHELVED           NovaListServersDetailsRequestStatus
	SHELVED_OFFLOADED NovaListServersDetailsRequestStatus
	SHUTOFF           NovaListServersDetailsRequestStatus
	UNKNOWN           NovaListServersDetailsRequestStatus
	VERIFY_RESIZE     NovaListServersDetailsRequestStatus
}

func GetNovaListServersDetailsRequestStatusEnum() NovaListServersDetailsRequestStatusEnum {
	return NovaListServersDetailsRequestStatusEnum{
		ACTIVE: NovaListServersDetailsRequestStatus{
			value: "ACTIVE",
		},
		BUILD: NovaListServersDetailsRequestStatus{
			value: "BUILD",
		},
		DELETED: NovaListServersDetailsRequestStatus{
			value: "DELETED",
		},
		ERROR: NovaListServersDetailsRequestStatus{
			value: "ERROR",
		},
		HARD_REBOOT: NovaListServersDetailsRequestStatus{
			value: "HARD_REBOOT",
		},
		MIGRATING: NovaListServersDetailsRequestStatus{
			value: "MIGRATING",
		},
		REBOOT: NovaListServersDetailsRequestStatus{
			value: "REBOOT",
		},
		RESIZE: NovaListServersDetailsRequestStatus{
			value: "RESIZE",
		},
		REVERT_RESIZE: NovaListServersDetailsRequestStatus{
			value: "REVERT_RESIZE",
		},
		SHELVED: NovaListServersDetailsRequestStatus{
			value: "SHELVED",
		},
		SHELVED_OFFLOADED: NovaListServersDetailsRequestStatus{
			value: "SHELVED_OFFLOADED",
		},
		SHUTOFF: NovaListServersDetailsRequestStatus{
			value: "SHUTOFF",
		},
		UNKNOWN: NovaListServersDetailsRequestStatus{
			value: "UNKNOWN",
		},
		VERIFY_RESIZE: NovaListServersDetailsRequestStatus{
			value: "VERIFY_RESIZE",
		},
	}
}

func (c NovaListServersDetailsRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NovaListServersDetailsRequestStatus) UnmarshalJSON(b []byte) error {
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
