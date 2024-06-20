package xtype

import "go/types"

// Accessible checks if obj is accessible within outputPackagePath.
func Accessible(obj types.Object, outputPackagePath string) bool {
	if obj.Exported() {
		return true
	}

	pkg := obj.Pkg()
	return pkg == nil || pkg.Path() == outputPackagePath
}
