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

// 挂钩实例信息
type InstanceHangingInfos struct {
	// 生命周期挂钩名称。
	LifecycleHookName *string `json:"lifecycle_hook_name,omitempty"`
	// 生命周期操作令牌，用于指定生命周期回调对象。
	LifecycleActionKey *string `json:"lifecycle_action_key,omitempty"`
	// 伸缩实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 伸缩组ID。
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	// 伸缩实例挂钩的挂起状态。HANGING：挂起。CONTINUE：继续。ABANDON：终止。
	LifecycleHookStatus *InstanceHangingInfosLifecycleHookStatus `json:"lifecycle_hook_status,omitempty"`
	// 超时时间，遵循UTC时间，格式为：YYYY-MM-DDThh:mm:ssZZ。
	Timeout *sdktime.SdkTime `json:"timeout,omitempty"`
	// 生命周期挂钩默认回调操作。
	DefaultResult *string `json:"default_result,omitempty"`
}

func (o InstanceHangingInfos) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"InstanceHangingInfos", string(data)}, " ")
}

type InstanceHangingInfosLifecycleHookStatus struct {
	value string
}

type InstanceHangingInfosLifecycleHookStatusEnum struct {
	HANGING  InstanceHangingInfosLifecycleHookStatus
	CONTINUE InstanceHangingInfosLifecycleHookStatus
	ABANDON  InstanceHangingInfosLifecycleHookStatus
}

func GetInstanceHangingInfosLifecycleHookStatusEnum() InstanceHangingInfosLifecycleHookStatusEnum {
	return InstanceHangingInfosLifecycleHookStatusEnum{
		HANGING: InstanceHangingInfosLifecycleHookStatus{
			value: "HANGING",
		},
		CONTINUE: InstanceHangingInfosLifecycleHookStatus{
			value: "CONTINUE",
		},
		ABANDON: InstanceHangingInfosLifecycleHookStatus{
			value: "ABANDON",
		},
	}
}

func (c InstanceHangingInfosLifecycleHookStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceHangingInfosLifecycleHookStatus) UnmarshalJSON(b []byte) error {
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
