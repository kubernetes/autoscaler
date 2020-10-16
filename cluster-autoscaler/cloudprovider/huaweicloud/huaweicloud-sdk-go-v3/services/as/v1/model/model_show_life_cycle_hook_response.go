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

// Response Object
type ShowLifeCycleHookResponse struct {
	// 生命周期挂钩名称。
	LifecycleHookName *string `json:"lifecycle_hook_name,omitempty"`
	// 生命周期挂钩类型。INSTANCE_TERMINATING;INSTANCE_LAUNCHING
	LifecycleHookType *ShowLifeCycleHookResponseLifecycleHookType `json:"lifecycle_hook_type,omitempty"`
	// 生命周期挂钩默认回调操作。ABANDON;CONTINUE
	DefaultResult *ShowLifeCycleHookResponseDefaultResult `json:"default_result,omitempty"`
	// 生命周期挂钩超时时间，单位秒。
	DefaultTimeout *int32 `json:"default_timeout,omitempty"`
	// SMN服务中Topic的唯一的资源标识。
	NotificationTopicUrn *string `json:"notification_topic_urn,omitempty"`
	// SMN服务中Topic的资源名称。
	NotificationTopicName *string `json:"notification_topic_name,omitempty"`
	// 自定义通知消息。
	NotificationMetadata *string `json:"notification_metadata,omitempty"`
	// 生命周期挂钩创建时间，遵循UTC时间。
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
}

func (o ShowLifeCycleHookResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowLifeCycleHookResponse", string(data)}, " ")
}

type ShowLifeCycleHookResponseLifecycleHookType struct {
	value string
}

type ShowLifeCycleHookResponseLifecycleHookTypeEnum struct {
	INSTANCE_TERMINATING ShowLifeCycleHookResponseLifecycleHookType
	INSTANCE_LAUNCHING   ShowLifeCycleHookResponseLifecycleHookType
}

func GetShowLifeCycleHookResponseLifecycleHookTypeEnum() ShowLifeCycleHookResponseLifecycleHookTypeEnum {
	return ShowLifeCycleHookResponseLifecycleHookTypeEnum{
		INSTANCE_TERMINATING: ShowLifeCycleHookResponseLifecycleHookType{
			value: "INSTANCE_TERMINATING",
		},
		INSTANCE_LAUNCHING: ShowLifeCycleHookResponseLifecycleHookType{
			value: "INSTANCE_LAUNCHING",
		},
	}
}

func (c ShowLifeCycleHookResponseLifecycleHookType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowLifeCycleHookResponseLifecycleHookType) UnmarshalJSON(b []byte) error {
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

type ShowLifeCycleHookResponseDefaultResult struct {
	value string
}

type ShowLifeCycleHookResponseDefaultResultEnum struct {
	ABANDON  ShowLifeCycleHookResponseDefaultResult
	CONTINUE ShowLifeCycleHookResponseDefaultResult
}

func GetShowLifeCycleHookResponseDefaultResultEnum() ShowLifeCycleHookResponseDefaultResultEnum {
	return ShowLifeCycleHookResponseDefaultResultEnum{
		ABANDON: ShowLifeCycleHookResponseDefaultResult{
			value: "ABANDON",
		},
		CONTINUE: ShowLifeCycleHookResponseDefaultResult{
			value: "CONTINUE",
		},
	}
}

func (c ShowLifeCycleHookResponseDefaultResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowLifeCycleHookResponseDefaultResult) UnmarshalJSON(b []byte) error {
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
