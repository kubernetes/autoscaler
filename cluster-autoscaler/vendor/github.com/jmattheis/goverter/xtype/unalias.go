//go:build go1.22
// +build go1.22

package xtype

import "go/types"

func Unalias(t types.Type) types.Type {
	return types.Unalias(t)
}
