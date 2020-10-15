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
	"strings"
)

// 修改生命周期挂钩
type UpdateLifeCycleHookRequestBody struct {
	// 生命周期挂钩类型。INSTANCE_TERMINATING。INSTANCE_LAUNCHING。INSTANCE_TERMINATING 类型的挂钩负责在实例终止时将实例挂起，INSTANCE_LAUNCHING 类型的挂钩则是在实例启动时将实例挂起。
	LifecycleHookType *UpdateLifeCycleHookRequestBodyLifecycleHookType `json:"lifecycle_hook_type,omitempty"`
	// 生命周期挂钩默认回调操作。默认情况下，到达超时时间后执行的操作。ABANDON；CONTINUE；如果实例正在启动，则 CONTINUE 表示用户自定义操作已成功，可将实例投入使用。否则，ABANDON 表示用户自定义操作未成功，终止实例，伸缩活动置为失败，重新创建新实例。如果实例正在终止，则 ABANDON 和 CONTINUE 都允许终止实例。不过，ABANDON 将停止其他生命周期挂钩，而 CONTINUE 将允许完成其他生命周期挂钩。该字段缺省时默认为 ABANDON。
	DefaultResult *UpdateLifeCycleHookRequestBodyDefaultResult `json:"default_result,omitempty"`
	// 生命周期挂钩超时时间，取值范围300-86400，默认为3600，单位是秒。默认情况下，实例保持等待状态的时间。您可以延长超时时间，也可以在超时时间结束前进行 CONTINUE 或 ABANDON 操作。
	DefaultTimeout *int32 `json:"default_timeout,omitempty"`
	// SMN 服务中 Topic 的唯一的资源标识。为生命周期挂钩定义一个通知目标，当实例被生命周期挂钩挂起时向该通知目标发送消息。该消息包含实例的基本信息、用户自定义通知消息，以及可用于控制生命周期操作的令牌信息。
	NotificationTopicUrn *string `json:"notification_topic_urn,omitempty"`
	// 自定义通知消息，长度不超过256位，不能包含字符< > & ' ( )当配置了通知目标时，可向其发送用户自定义的通知内容。
	NotificationMetadata *string `json:"notification_metadata,omitempty"`
}

func (o UpdateLifeCycleHookRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateLifeCycleHookRequestBody", string(data)}, " ")
}

type UpdateLifeCycleHookRequestBodyLifecycleHookType struct {
	value string
}

type UpdateLifeCycleHookRequestBodyLifecycleHookTypeEnum struct {
	INSTANCE_TERMINATING UpdateLifeCycleHookRequestBodyLifecycleHookType
	INSTANCE_LAUNCHING   UpdateLifeCycleHookRequestBodyLifecycleHookType
}

func GetUpdateLifeCycleHookRequestBodyLifecycleHookTypeEnum() UpdateLifeCycleHookRequestBodyLifecycleHookTypeEnum {
	return UpdateLifeCycleHookRequestBodyLifecycleHookTypeEnum{
		INSTANCE_TERMINATING: UpdateLifeCycleHookRequestBodyLifecycleHookType{
			value: "INSTANCE_TERMINATING",
		},
		INSTANCE_LAUNCHING: UpdateLifeCycleHookRequestBodyLifecycleHookType{
			value: "INSTANCE_LAUNCHING",
		},
	}
}

func (c UpdateLifeCycleHookRequestBodyLifecycleHookType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateLifeCycleHookRequestBodyLifecycleHookType) UnmarshalJSON(b []byte) error {
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

type UpdateLifeCycleHookRequestBodyDefaultResult struct {
	value string
}

type UpdateLifeCycleHookRequestBodyDefaultResultEnum struct {
	ABANDON  UpdateLifeCycleHookRequestBodyDefaultResult
	CONTINUE UpdateLifeCycleHookRequestBodyDefaultResult
}

func GetUpdateLifeCycleHookRequestBodyDefaultResultEnum() UpdateLifeCycleHookRequestBodyDefaultResultEnum {
	return UpdateLifeCycleHookRequestBodyDefaultResultEnum{
		ABANDON: UpdateLifeCycleHookRequestBodyDefaultResult{
			value: "ABANDON",
		},
		CONTINUE: UpdateLifeCycleHookRequestBodyDefaultResult{
			value: "CONTINUE",
		},
	}
}

func (c UpdateLifeCycleHookRequestBodyDefaultResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateLifeCycleHookRequestBodyDefaultResult) UnmarshalJSON(b []byte) error {
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
