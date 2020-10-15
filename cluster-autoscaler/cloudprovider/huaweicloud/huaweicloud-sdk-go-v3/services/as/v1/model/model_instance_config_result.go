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

// 实例配置信息
type InstanceConfigResult struct {
	// 云服务器的规格ID。
	FlavorRef *string `json:"flavorRef,omitempty"`
	// 镜像ID，同image_id。
	ImageRef *string `json:"imageRef,omitempty"`
	// 磁盘组信息。
	Disk *[]Disk `json:"disk,omitempty"`
	// 登录云服务器的SSH密钥名称。
	KeyName *string `json:"key_name,omitempty"`
	// 该参数为预留字段。
	InstanceName *string `json:"instance_name,omitempty"`
	// 该参数为预留字段。
	InstanceId *string `json:"instance_id,omitempty"`
	// 登录云服务器的密码，非明文回显。
	AdminPass   *string      `json:"adminPass,omitempty"`
	Personality *Personality `json:"personality,omitempty"`
	PublicIp    *PublicIp    `json:"public_ip,omitempty"`
	// cloud-init用户数据，base64格式编码。
	UserData *string   `json:"user_data,omitempty"`
	Metadata *MetaData `json:"metadata,omitempty"`
	// 安全组信息。
	SecurityGroups *[]SecurityGroups `json:"security_groups,omitempty"`
	// 云服务器组ID。
	ServerGroupId *string `json:"server_group_id,omitempty"`
	// 在专属主机上创建弹性云服务器。
	Tenancy *string `json:"tenancy,omitempty"`
	// 专属主机的ID。
	DedicatedHostId *string `json:"dedicated_host_id,omitempty"`
}

func (o InstanceConfigResult) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"InstanceConfigResult", string(data)}, " ")
}
