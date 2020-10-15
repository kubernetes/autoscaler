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

// Response Object
type ShowPolicyAndInstanceQuotaResponse struct {
	AllQuotas *PolicyInstanceQuotas `json:"AllQuotas,omitempty"`
}

func (o ShowPolicyAndInstanceQuotaResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowPolicyAndInstanceQuotaResponse", string(data)}, " ")
}
