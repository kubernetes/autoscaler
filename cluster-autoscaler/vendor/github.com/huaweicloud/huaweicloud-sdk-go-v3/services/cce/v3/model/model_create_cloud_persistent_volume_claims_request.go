package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateCloudPersistentVolumeClaimsRequest struct {
	// Namespace是对一组资源和对象的抽象集合，用来将系统内部的对象划分为不同的项目组或用户组。以小写字母开头，由小写字母、数字、中划线（-）组成，且不能以中划线（-）结尾。  使用namespace有如下约束：  - 用户自定义的namespace，使用前必须先[[创建Namespace](https://support.huaweicloud.com/api-cce/cce_02_0050.html)](tag:hws)[[创建Namespace](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0050.html)](tag:hws_hk)  - 系统自带的namespace：default  - 不能使用kube-system与kube-public

	Namespace string `json:"namespace"`
	// 集群ID，使用**https://Endpoint/uri**这种URL格式时必须指定此参数。获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	XClusterID *string `json:"X-Cluster-ID,omitempty"`

	Body *PersistentVolumeClaim `json:"body,omitempty"`
}

func (o CreateCloudPersistentVolumeClaimsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateCloudPersistentVolumeClaimsRequest struct{}"
	}

	return strings.Join([]string{"CreateCloudPersistentVolumeClaimsRequest", string(data)}, " ")
}
