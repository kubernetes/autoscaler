package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

// SkipCopy handles FlagSkipCopySameType.
type SkipCopy struct{}

// Matches returns true, if the builder can create handle the given types.
func (*SkipCopy) Matches(ctx *MethodContext, source, target *xtype.Type) bool {
	return ctx.Conf.SkipCopySameType && source.String == target.String
}

// Build creates conversion source code for the given source and target type.
func (*SkipCopy) Build(_ Generator, _ *MethodContext, sourceID *xtype.JenID, _, _ *xtype.Type, _ ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	return nil, sourceID, nil
}
