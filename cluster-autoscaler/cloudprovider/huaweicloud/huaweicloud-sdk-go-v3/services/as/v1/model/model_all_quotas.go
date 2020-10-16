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

// 配额列表
type AllQuotas struct {
	// 配额详情资源列表。
	Resources *[]AllResources `json:"resources,omitempty"`
}

func (o AllQuotas) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AllQuotas", string(data)}, " ")
}
