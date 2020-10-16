/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//  弹性云服务器调度信息。  裸金属服务器场景不支持。
type NovaCreateServersSchedulerHint struct {
	// 反亲和性组信息。  UUID格式。
	Group *string `json:"group,omitempty"`
	// 与指定弹性云服务器满足反亲和性。   当前不支持该功能。
	DifferentHost *[]string `json:"different_host,omitempty"`
	// 与指定的弹性云服务器满足亲和性。   当前不支持该功能。
	SameHost *[]string `json:"same_host,omitempty"`
	// 将弹性云服务器scheduler到指定网段的host上，host网段的cidr。   当前不支持该功能。
	Cidr *string `json:"cidr,omitempty"`
	// 将弹性云服务器scheduler到指定网段的host上，host IP地址。   当前不支持该功能。
	BuildNearHostIp *string `json:"build_near_host_ip,omitempty"`
}

func (o NovaCreateServersSchedulerHint) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaCreateServersSchedulerHint", string(data)}, " ")
}
