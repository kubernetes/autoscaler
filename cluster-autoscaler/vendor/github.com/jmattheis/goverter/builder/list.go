package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

// List handles array / slice types.
type List struct{}

// Matches returns true, if the builder can create handle the given types.
func (*List) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return source.List && target.List && !target.ListFixed
}

// Build creates conversion source code for the given source and target type.
func (*List) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Nil())
	targetSlice := ctx.Name(target.ID())
	index := ctx.Index()

	indexedSource := xtype.VariableID(sourceID.Code.Clone().Index(jen.Id(index)))

	forBlock, newID, err := gen.Build(ctx, indexedSource, source.ListInner, target.ListInner, path.Index(jen.Id(index)))
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "[]",
			SourceType: source.ListInner.String,
			TargetID:   "[]",
			TargetType: target.ListInner.String,
		})
	}
	forBlock = append(forBlock, jen.Id(targetSlice).Index(jen.Id(index)).Op("=").Add(newID.Code))
	forStmt := jen.For(jen.Id(index).Op(":=").Lit(0), jen.Id(index).Op("<").Len(sourceID.Code.Clone()), jen.Id(index).Op("++")).
		Block(forBlock...)

	stmt := []jen.Code{}
	if source.ListFixed {
		stmt = []jen.Code{
			jen.Id(targetSlice).Op(":=").Make(target.TypeAsJen(), jen.Len(sourceID.Code.Clone())),
			forStmt,
		}
	} else {
		stmt = []jen.Code{
			jen.Var().Add(jen.Id(targetSlice), target.TypeAsJen()),
			jen.If(sourceID.Code.Clone().Op("!=").Nil()).Block(
				jen.Id(targetSlice).Op("=").Make(target.TypeAsJen(), jen.Len(sourceID.Code.Clone())),
				forStmt,
			),
		}
	}

	return stmt, xtype.VariableID(jen.Id(targetSlice)), nil
}
