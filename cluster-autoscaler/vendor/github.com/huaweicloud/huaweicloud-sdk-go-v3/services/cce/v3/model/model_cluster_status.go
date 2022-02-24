package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type ClusterStatus struct {
	// 集群状态，取值如下 - Available：可用，表示集群处于正常状态。 - Unavailable：不可用，表示集群异常，需手动删除或联系管理员删除。 - ScalingUp：扩容中，表示集群正处于扩容过程中。 - ScalingDown：缩容中，表示集群正处于缩容过程中。 - Creating：创建中，表示集群正处于创建过程中。 - Deleting：删除中，表示集群正处于删除过程中。 - Upgrading：升级中，表示集群正处于升级过程中。 - Resizing：规格变更中，表示集群正处于变更规格中。 - RollingBack：回滚中，表示集群正处于回滚过程中。 - RollbackFailed：回滚异常，表示集群回滚异常，需联系管理员进行回滚重试。 - Empty：集群无任何资源

	Phase *string `json:"phase,omitempty"`
	// 作业ID

	JobID *string `json:"jobID,omitempty"`
	// 集群变为当前状态的原因，在集群在非“Available”状态下时，会返回此参数。

	Reason *string `json:"reason,omitempty"`
	// 集群变为当前状态的原因的详细信息，在集群在非“Available”状态下时，会返回此参数。

	Message *string `json:"message,omitempty"`
	// 集群中 kube-apiserver 的访问地址。

	Endpoints *[]ClusterEndpoints `json:"endpoints,omitempty"`
	// CBC资源锁定

	IsLocked *bool `json:"isLocked,omitempty"`
	// CBC资源锁定场景

	LockScene *string `json:"lockScene,omitempty"`
	// 锁定资源

	LockSource *string `json:"lockSource,omitempty"`
	// 锁定的资源ID

	LockSourceId *string `json:"lockSourceId,omitempty"`
	// 删除配置状态（仅删除请求响应包含）

	DeleteOption *interface{} `json:"deleteOption,omitempty"`
	// 删除状态信息（仅删除请求响应包含）

	DeleteStatus *interface{} `json:"deleteStatus,omitempty"`
}

func (o ClusterStatus) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ClusterStatus struct{}"
	}

	return strings.Join([]string{"ClusterStatus", string(data)}, " ")
}
