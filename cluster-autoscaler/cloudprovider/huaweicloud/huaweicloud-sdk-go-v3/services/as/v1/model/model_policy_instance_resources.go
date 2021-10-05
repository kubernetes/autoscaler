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

// 配额资源
type PolicyInstanceResources struct {
	// 查询配额的类型。
	Type *string `json:"type,omitempty"`
	// 已使用的配额数量。
	Used *int32 `json:"used,omitempty"`
	// 配额总数量。
	Quota *int32 `json:"quota,omitempty"`
	// 配额上限。
	Max *int32 `json:"max,omitempty"`
}

func (o PolicyInstanceResources) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PolicyInstanceResources", string(data)}, " ")
}
