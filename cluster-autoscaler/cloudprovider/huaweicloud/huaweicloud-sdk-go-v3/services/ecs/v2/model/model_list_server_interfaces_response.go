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

// Response Object
type ListServerInterfacesResponse struct {
	// 云服务器网卡信息列表
	InterfaceAttachments *[]InterfaceAttachment `json:"interfaceAttachments,omitempty"`
}

func (o ListServerInterfacesResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListServerInterfacesResponse", string(data)}, " ")
}
