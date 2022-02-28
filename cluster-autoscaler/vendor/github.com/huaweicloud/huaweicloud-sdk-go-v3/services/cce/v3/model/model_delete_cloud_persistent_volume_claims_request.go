package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteCloudPersistentVolumeClaimsRequest struct {
	// 需要删除的PersistentVolumClaim的名称。

	Name string `json:"name"`
	// 指定PersistentVolumeClaim所在的命名空间。

	Namespace string `json:"namespace"`
	// 删除PersistentVolumeClaim后是否保留后端关联的云存储。false表示不删除，true表示删除，默认为false。

	DeleteVolume *string `json:"deleteVolume,omitempty"`
	// 删除PersistentVolumeClaim后是否保留后端关联的云存储。false表示不删除，true表示删除，默认为false。 云存储的类型，和deleteVolume搭配使用。即deleteVolume和storageType必须同时配置。     - bs：EVS云硬盘存储     - nfs：SFS弹性文件存储     - obs：OBS对象存储     [- efs：SFS Turbo极速文件存储](tag:hws)

	StorageType *string `json:"storageType,omitempty"`
	// 集群ID，使用**https://Endpoint/uri**这种URL格式时必须指定此参数。获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	XClusterID *string `json:"X-Cluster-ID,omitempty"`
}

func (o DeleteCloudPersistentVolumeClaimsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteCloudPersistentVolumeClaimsRequest struct{}"
	}

	return strings.Join([]string{"DeleteCloudPersistentVolumeClaimsRequest", string(data)}, " ")
}
