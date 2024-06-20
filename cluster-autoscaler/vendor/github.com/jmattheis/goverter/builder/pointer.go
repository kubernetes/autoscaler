package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

// Pointer handles pointer types.
type Pointer struct{}

// Matches returns true, if the builder can create handle the given types.
func (*Pointer) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return source.Pointer && target.Pointer
}

// Build creates conversion source code for the given source and target type.
func (*Pointer) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, errPath ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Nil())

	stmt, outerVar, err := buildTargetVar(gen, ctx, sourceID, source, target, errPath)
	if err != nil {
		return nil, nil, err
	}

	valueSourceID := jen.Op("*").Add(sourceID.Code.Clone())
	if !source.PointerInner.Basic {
		valueSourceID = jen.Parens(valueSourceID)
	}

	innerID := xtype.OtherID(valueSourceID)
	innerID.ParentPointer = sourceID
	nextBlock, id, err := gen.Build(
		ctx, innerID, source.PointerInner, target.PointerInner, errPath)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "*",
			SourceType: source.PointerInner.String,
			TargetID:   "*",
			TargetType: target.PointerInner.String,
		})
	}

	pstmt, tmpID := id.Pointer(target.PointerInner, ctx.Name)

	ifBlock := append(nextBlock, pstmt...)
	ifBlock = append(ifBlock, outerVar.Clone().Op("=").Add(tmpID.Code))

	stmt = append(stmt,
		jen.If(sourceID.Code.Clone().Op("!=").Nil()).Block(ifBlock...),
	)

	return stmt, xtype.VariableID(outerVar), err
}

// SourcePointer handles type were only the source is a pointer.
type SourcePointer struct{}

// Matches returns true, if the builder can create handle the given types.
func (*SourcePointer) Matches(ctx *MethodContext, source, target *xtype.Type) bool {
	return ctx.Conf.UseZeroValueOnPointerInconsistency && source.Pointer && !target.Pointer
}

// Build creates conversion source code for the given source and target type.
func (*SourcePointer) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Id(target.ID()))

	valueSourceID := jen.Op("*").Add(sourceID.Code.Clone())
	if !source.PointerInner.Basic {
		valueSourceID = jen.Parens(valueSourceID)
	}

	innerID := xtype.OtherID(valueSourceID)
	innerID.ParentPointer = sourceID

	stmt, valueVar, err := buildTargetVar(gen, ctx, sourceID, source, target, path)
	if err != nil {
		return nil, nil, err
	}

	nextInner, nextID, err := gen.Build(ctx, innerID, source.PointerInner, target, path)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "*",
			SourceType: source.PointerInner.String,
		})
	}

	stmt = append(stmt,
		jen.If(sourceID.Code.Clone().Op("!=").Nil()).Block(
			append(nextInner, valueVar.Clone().Op("=").Add(nextID.Code))...,
		),
	)

	return stmt, xtype.VariableID(valueVar), nil
}

// TargetPointer handles type were only the target is a pointer.
type TargetPointer struct{}

// Matches returns true, if the builder can create handle the given types.
func (*TargetPointer) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return !source.Pointer && target.Pointer
}

// Build creates conversion source code for the given source and target type.
func (*TargetPointer) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Nil())
	stmt, id, err := gen.Build(ctx, sourceID, source, target.PointerInner, path)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "*",
			SourceType: source.String,
			TargetID:   "*",
			TargetType: target.PointerInner.String,
		})
	}

	pstmt, nextID := id.Pointer(target.PointerInner, ctx.Name)
	stmt = append(stmt, pstmt...)
	return stmt, nextID, nil
}
