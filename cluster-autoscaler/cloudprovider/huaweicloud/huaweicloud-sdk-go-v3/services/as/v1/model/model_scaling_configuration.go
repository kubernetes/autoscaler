/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

// 伸缩配置详情
type ScalingConfiguration struct {
	// 伸缩配置ID，全局唯一。
	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 租户ID。
	Tenant *string `json:"tenant,omitempty"`
	// 伸缩配置名称。
	ScalingConfigurationName *string               `json:"scaling_configuration_name,omitempty"`
	InstanceConfig           *InstanceConfigResult `json:"instance_config,omitempty"`
	// 创建伸缩配置的时间，遵循UTC时间。
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
}

func (o ScalingConfiguration) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScalingConfiguration", string(data)}, " ")
}
