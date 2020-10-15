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

// 安全组信息
type SecurityGroupsResult struct {
	// 安全组ID
	Id *string `json:"id,omitempty"`
}

func (o SecurityGroupsResult) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"SecurityGroupsResult", string(data)}, " ")
}
