package enum

import (
	"go/constant"
	"go/types"
)

func Detect(named *types.Named) (Enum, bool) {
	basic, ok := named.Underlying().(*types.Basic)
	if !ok {
		return Enum{}, false
	}

	if basic.Info()&(types.IsFloat|types.IsString|types.IsInteger) == 0 {
		return Enum{}, false
	}

	scope := named.Obj().Pkg().Scope()

	members := map[string]any{}
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}

		if types.Identical(named, c.Type()) {
			members[name] = constant.Val(c.Val())
		}
	}

	if len(members) == 0 {
		return Enum{}, false
	}

	return Enum{Type: named, Members: members}, true
}
