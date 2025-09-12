package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

// Map handles map types.
type Map struct{}

// Matches returns true, if the builder can create handle the given types.
func (*Map) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return source.Map && target.Map
}

// Build creates conversion source code for the given source and target type.
func (*Map) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, errPath ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Nil())
	targetMap := ctx.Name(target.ID())
	key, value := ctx.Map()

	errPath = errPath.Key(jen.Id(key))

	block, newKey, err := gen.Build(ctx, xtype.VariableID(jen.Id(key)), source.MapKey, target.MapKey, errPath)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "[]",
			SourceType: "<mapkey> " + source.MapKey.String,
			TargetID:   "[]",
			TargetType: "<mapkey> " + target.MapKey.String,
		})
	}
	valueStmt, valueKey, err := gen.Build(
		ctx, xtype.VariableID(jen.Id(value)), source.MapValue, target.MapValue, errPath)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "[]",
			SourceType: "<mapvalue> " + source.MapValue.String,
			TargetID:   "[]",
			TargetType: "<mapvalue> " + target.MapValue.String,
		})
	}
	block = append(block, valueStmt...)
	block = append(block, jen.Id(targetMap).Index(newKey.Code).Op("=").Add(valueKey.Code))

	stmt := []jen.Code{
		jen.Var().Add(jen.Id(targetMap), target.TypeAsJen()),
		jen.If(sourceID.Code.Clone().Op("!=").Nil()).Block(
			jen.Id(targetMap).Op("=").Make(target.TypeAsJen(), jen.Len(sourceID.Code.Clone())),
			jen.For(jen.List(jen.Id(key), jen.Id(value)).Op(":=").Range().Add(sourceID.Code)).
				Block(block...),
		),
	}

	return stmt, xtype.VariableID(jen.Id(targetMap)), nil
}
