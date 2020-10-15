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

//
type ProjectFlavorLimit struct {
}

func (o ProjectFlavorLimit) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ProjectFlavorLimit", string(data)}, " ")
}
