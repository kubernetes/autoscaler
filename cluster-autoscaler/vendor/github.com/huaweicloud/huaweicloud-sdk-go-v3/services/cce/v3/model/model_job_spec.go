package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type JobSpec struct {
	// 任务的类型，例：“CreateCluster”- 创建集群。

	Type *string `json:"type,omitempty"`
	// 任务所在的集群的ID。

	ClusterUID *string `json:"clusterUID,omitempty"`
	// 任务操作的资源ID。

	ResourceID *string `json:"resourceID,omitempty"`
	// 任务操作的资源名称。

	ResourceName *string `json:"resourceName,omitempty"`
	// 扩展参数。

	ExtendParam map[string]string `json:"extendParam,omitempty"`
	// 子任务的列表。  - 包含了所有子任务的详细信息 - 在创建集群、节点等场景下，通常会由多个子任务共同组成创建任务，在子任务都完成后，任务才会完成

	SubJobs *[]Job `json:"subJobs,omitempty"`
}

func (o JobSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "JobSpec struct{}"
	}

	return strings.Join([]string{"JobSpec", string(data)}, " ")
}
