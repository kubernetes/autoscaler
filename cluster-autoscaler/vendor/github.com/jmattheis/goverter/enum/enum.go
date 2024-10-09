package enum

import "go/types"

type Enum struct {
	Type    *types.Named
	Members map[string]any
}
