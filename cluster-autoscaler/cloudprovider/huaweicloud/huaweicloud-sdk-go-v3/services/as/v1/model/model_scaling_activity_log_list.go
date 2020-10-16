/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// 伸缩活动日志列表。
type ScalingActivityLogList struct {
	// 伸缩活动状态：SUCCESS：成功。FAIL：失败。DOING：伸缩过程中。
	Status *ScalingActivityLogListStatus `json:"status,omitempty"`
	// 伸缩活动触发时间，遵循UTC时间。
	StartTime *sdktime.SdkTime `json:"start_time,omitempty"`
	// 伸缩活动结束时间，遵循UTC时间。
	EndTime *sdktime.SdkTime `json:"end_time,omitempty"`
	// 伸缩活动日志ID。
	Id *string `json:"id,omitempty"`
	// 完成伸缩活动且只被移出弹性伸缩组的云服务器名称列表，云服务器名之间以逗号分隔。
	InstanceRemovedList *string `json:"instance_removed_list,omitempty"`
	// 完成伸缩活动且被移出弹性伸缩组并删除的云服务器名称列表，云服务器名之间以逗号分隔。
	InstanceDeletedList *string `json:"instance_deleted_list,omitempty"`
	// 完成伸缩活动且被加入弹性伸缩组的云服务器名称列表，云服务器名之间以逗号分割。
	InstanceAddedList *string `json:"instance_added_list,omitempty"`
	// 伸缩活动中变化（增加或减少）的云服务器数量。
	ScalingValue *string `json:"scaling_value,omitempty"`
	// 伸缩活动的描述信息。
	Description *string `json:"description,omitempty"`
	// 伸缩组当前instance值。
	InstanceValue *int32 `json:"instance_value,omitempty"`
	// 伸缩活动最终desire值。
	DesireValue *int32 `json:"desire_value,omitempty"`
}

func (o ScalingActivityLogList) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScalingActivityLogList", string(data)}, " ")
}

type ScalingActivityLogListStatus struct {
	value string
}

type ScalingActivityLogListStatusEnum struct {
	SUCCESS ScalingActivityLogListStatus
	FAIL    ScalingActivityLogListStatus
	DING    ScalingActivityLogListStatus
}

func GetScalingActivityLogListStatusEnum() ScalingActivityLogListStatusEnum {
	return ScalingActivityLogListStatusEnum{
		SUCCESS: ScalingActivityLogListStatus{
			value: "SUCCESS",
		},
		FAIL: ScalingActivityLogListStatus{
			value: "FAIL",
		},
		DING: ScalingActivityLogListStatus{
			value: "DING",
		},
	}
}

func (c ScalingActivityLogListStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScalingActivityLogListStatus) UnmarshalJSON(b []byte) error {
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
