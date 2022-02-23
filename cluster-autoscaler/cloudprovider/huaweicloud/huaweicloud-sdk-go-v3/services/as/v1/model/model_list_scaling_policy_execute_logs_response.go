package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingPolicyExecuteLogsResponse struct {
	// 总记录数。

	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。

	Limit *int32 `json:"limit,omitempty"`
	// 伸缩策略执行日志列表。

	ScalingPolicyExecuteLog *[]ScalingPolicyExecuteLogList `json:"scaling_policy_execute_log,omitempty"`
	HttpStatusCode          int                            `json:"-"`
}

func (o ListScalingPolicyExecuteLogsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingPolicyExecuteLogsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingPolicyExecuteLogsResponse", string(data)}, " ")
}
