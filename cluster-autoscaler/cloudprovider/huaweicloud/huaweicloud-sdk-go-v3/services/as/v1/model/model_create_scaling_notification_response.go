/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateScalingNotificationResponse struct {
	// SMN服务中Topic的唯一的资源标识。
	TopicUrn *string `json:"topic_urn,omitempty"`
	// 通知场景，有以下五种类型。SCALING_UP：扩容成功。SCALING_UP_FAIL：扩容失败。SCALING_DOWN：减容成功。SCALING_DOWN_FAIL：减容失败。SCALING_GROUP_ABNORMAL：伸缩组发生异常
	TopicScene *[]string `json:"topic_scene,omitempty"`
	// SMN服务中Topic的资源名称。
	TopicName *string `json:"topic_name,omitempty"`
}

func (o CreateScalingNotificationResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingNotificationResponse", string(data)}, " ")
}
