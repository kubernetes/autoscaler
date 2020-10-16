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

// Request Object
type ShowResourceQuotaRequest struct {
}

func (o ShowResourceQuotaRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowResourceQuotaRequest", string(data)}, " ")
}
