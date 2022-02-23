package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 配置伸缩组通知
type CreateNotificationOption struct {
	// SMN服务中Topic的唯一的资源标识。

	TopicUrn string `json:"topic_urn"`
	// 通知场景，有以下五种类型。SCALING_UP：扩容成功。SCALING_UP_FAIL：扩容失败。SCALING_DOWN：减容成功。SCALING_DOWN_FAIL：减容失败。SCALING_GROUP_ABNORMAL：伸缩组发生异常

	TopicScene []CreateNotificationOptionTopicScene `json:"topic_scene"`
}

func (o CreateNotificationOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateNotificationOption struct{}"
	}

	return strings.Join([]string{"CreateNotificationOption", string(data)}, " ")
}

type CreateNotificationOptionTopicScene struct {
	value string
}

type CreateNotificationOptionTopicSceneEnum struct {
	SCALING_UP             CreateNotificationOptionTopicScene
	SCALING_UP_FAIL        CreateNotificationOptionTopicScene
	SCALING_DOWN           CreateNotificationOptionTopicScene
	SCALING_DOWN_FAIL      CreateNotificationOptionTopicScene
	SCALING_GROUP_ABNORMAL CreateNotificationOptionTopicScene
}

func GetCreateNotificationOptionTopicSceneEnum() CreateNotificationOptionTopicSceneEnum {
	return CreateNotificationOptionTopicSceneEnum{
		SCALING_UP: CreateNotificationOptionTopicScene{
			value: "[SCALING_UP",
		},
		SCALING_UP_FAIL: CreateNotificationOptionTopicScene{
			value: "SCALING_UP_FAIL",
		},
		SCALING_DOWN: CreateNotificationOptionTopicScene{
			value: "SCALING_DOWN",
		},
		SCALING_DOWN_FAIL: CreateNotificationOptionTopicScene{
			value: "SCALING_DOWN_FAIL",
		},
		SCALING_GROUP_ABNORMAL: CreateNotificationOptionTopicScene{
			value: "SCALING_GROUP_ABNORMAL]",
		},
	}
}

func (c CreateNotificationOptionTopicScene) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *CreateNotificationOptionTopicScene) UnmarshalJSON(b []byte) error {
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
