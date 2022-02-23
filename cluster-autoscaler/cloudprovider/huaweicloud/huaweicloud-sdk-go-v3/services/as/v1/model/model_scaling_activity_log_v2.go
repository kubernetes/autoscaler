package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

// 伸缩活动日志列表。
type ScalingActivityLogV2 struct {
	// 伸缩活动状态：SUCCESS：成功。FAIL：失败。DOING：伸缩过程中。

	Status *ScalingActivityLogV2Status `json:"status,omitempty"`
	// 伸缩活动触发时间，遵循UTC时间。

	StartTime *sdktime.SdkTime `json:"start_time,omitempty"`
	// 伸缩活动结束时间，遵循UTC时间。

	EndTime *sdktime.SdkTime `json:"end_time,omitempty"`
	// 伸缩活动日志ID。

	Id *string `json:"id,omitempty"`
	// 完成伸缩活动且只被移出弹性伸缩组的云服务器名称列表，云服务信息之间以逗号分隔。

	InstanceRemovedList *[]ScalingInstance `json:"instance_removed_list,omitempty"`
	// 完成伸缩活动且被移出弹性伸缩组并删除的云服务器名称列表，云服务器信息之间以逗号分隔。

	InstanceDeletedList *[]ScalingInstance `json:"instance_deleted_list,omitempty"`
	// 完成伸缩活动且被加入弹性伸缩组的云服务器名称列表，云服务器信息之间以逗号分割。

	InstanceAddedList *[]ScalingInstance `json:"instance_added_list,omitempty"`
	// 弹性伸缩组中伸缩活动失败的云服务器列表。

	InstanceFailedList *[]ScalingInstance `json:"instance_failed_list,omitempty"`
	// 完成伸缩活动且被转入/移出备用状态的云服务器列表

	InstanceStandbyList *[]ScalingInstance `json:"instance_standby_list,omitempty"`
	// 伸缩活动中变化（增加或减少）的云服务器数量。

	ScalingValue *string `json:"scaling_value,omitempty"`
	// 伸缩活动的描述信息。

	Description *string `json:"description,omitempty"`
	// 伸缩组当前instance值。

	InstanceValue *int32 `json:"instance_value,omitempty"`
	// 伸缩活动最终desire值。

	DesireValue *int32 `json:"desire_value,omitempty"`
	// 绑定成功的负载均衡器列表。

	LbBindSuccessList *[]ModifyLb `json:"lb_bind_success_list,omitempty"`
	// 绑定失败的负载均衡器列表。

	LbBindFailedList *[]ModifyLb `json:"lb_bind_failed_list,omitempty"`
	// 解绑成功的负载均衡器列表。

	LbUnbindSuccessList *[]ModifyLb `json:"lb_unbind_success_list,omitempty"`
	// 解绑失败的负载均衡器列表。

	LbUnbindFailedList *[]ModifyLb `json:"lb_unbind_failed_list,omitempty"`
	// 伸缩组活动类型

	Type *string `json:"type,omitempty"`
}

func (o ScalingActivityLogV2) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingActivityLogV2 struct{}"
	}

	return strings.Join([]string{"ScalingActivityLogV2", string(data)}, " ")
}

type ScalingActivityLogV2Status struct {
	value string
}

type ScalingActivityLogV2StatusEnum struct {
	SUCCESS ScalingActivityLogV2Status
	FAIL    ScalingActivityLogV2Status
	DING    ScalingActivityLogV2Status
}

func GetScalingActivityLogV2StatusEnum() ScalingActivityLogV2StatusEnum {
	return ScalingActivityLogV2StatusEnum{
		SUCCESS: ScalingActivityLogV2Status{
			value: "SUCCESS",
		},
		FAIL: ScalingActivityLogV2Status{
			value: "FAIL",
		},
		DING: ScalingActivityLogV2Status{
			value: "DING",
		},
	}
}

func (c ScalingActivityLogV2Status) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingActivityLogV2Status) UnmarshalJSON(b []byte) error {
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
