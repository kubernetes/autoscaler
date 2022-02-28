package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// spec是集合类的元素类型，内容为插件实例安装/升级的具体请求信息
type InstanceRequestSpec struct {
	// 待安装、升级插件的具体版本版本号，例如1.0.0

	Version string `json:"version"`
	// 集群id

	ClusterID string `json:"clusterID"`
	// 插件模板安装参数（各插件不同），升级插件时需要填写全量安装参数，未填写参数将使用插件模板中的默认值，当前插件安装参数可通过查询插件实例接口获取。

	Values map[string]interface{} `json:"values"`
	// 待安装插件模板名称，如coredns

	AddonTemplateName string `json:"addonTemplateName"`
}

func (o InstanceRequestSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "InstanceRequestSpec struct{}"
	}

	return strings.Join([]string{"InstanceRequestSpec", string(data)}, " ")
}
