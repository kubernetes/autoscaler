package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

// Basic handles basic data types.
type Basic struct{}

// Matches returns true, if the builder can create handle the given types.
func (*Basic) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return source.Basic && target.Basic &&
		source.BasicType.Kind() == target.BasicType.Kind()
}

// Build creates conversion source code for the given source and target type.
func (*Basic) Build(_ Generator, _ *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, errPath ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	if target.Named || (!target.Named && source.Named) {
		return nil, xtype.OtherID(target.TypeAsJen().Call(sourceID.Code)), nil
	}
	return nil, sourceID, nil
}

// BasicTargetPointerRule handles edge conditions if the target type is a pointer.
type BasicTargetPointerRule struct{}

// Matches returns true, if the builder can create handle the given types.
func (*BasicTargetPointerRule) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return source.Basic && target.Pointer && target.PointerInner.Basic
}

// Build creates conversion source code for the given source and target type.
func (*BasicTargetPointerRule) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, errPath ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	name := ctx.Name(target.ID())
	ctx.SetErrorTargetVar(jen.Nil())

	stmt, id, err := gen.Build(ctx, sourceID, source, target.PointerInner, errPath)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "*",
			SourceType: source.String,
			TargetID:   "*",
			TargetType: target.PointerInner.String,
		})
	}
	stmt = append(stmt, jen.Id(name).Op(":=").Add(id.Code))
	newID := jen.Op("&").Id(name)

	return stmt, xtype.OtherID(newID), err
}
