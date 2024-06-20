package xtype

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

func toCode(t types.Type) *jen.Statement {
	switch cast := t.(type) {
	case *types.Named:
		return toCodeNamed(cast)
	case *types.Map:
		return jen.Map(toCode(cast.Key())).Add(toCode(cast.Elem()))
	case *types.Slice:
		return jen.Index().Add(toCode(cast.Elem()))
	case *types.Array:
		return jen.Index(jen.Lit(int(cast.Len()))).Add(toCode(cast.Elem()))
	case *types.Pointer:
		return jen.Op("*").Add(toCode(cast.Elem()))
	case *types.Basic:
		return toCodeBasic(cast.Kind())
	case *types.Struct:
		return toCodeStruct(cast)
	case *types.Interface:
		return toCodeInterface(cast)
	case *types.Signature:
		return jen.Func().Add(toCodeSignature(cast))
	}
	panic("unsupported type " + t.String())
}

func toCodeInterface(t *types.Interface) *jen.Statement {
	content := []jen.Code{}
	for i := 0; i < t.NumEmbeddeds(); i++ {
		content = append(content, toCode(t.EmbeddedType(i)))
	}

	for i := 0; i < t.NumExplicitMethods(); i++ {
		method := t.ExplicitMethod(i)
		content = append(content, toCodeFunc(method))
	}

	return jen.Interface(content...)
}

func toCodeFunc(t *types.Func) *jen.Statement {
	sig := t.Type().(*types.Signature)
	return jen.Id(t.Name()).Add(toCodeSignature(sig))
}

func toCodeSignature(t *types.Signature) *jen.Statement {
	jenParams := []jen.Code{}
	params := t.Params()
	for i := 0; i < params.Len(); i++ {
		jenParams = append(jenParams, toCode(params.At(i).Type()))
	}

	jenResults := []jen.Code{}
	results := t.Results()
	for i := 0; i < results.Len(); i++ {
		jenResults = append(jenResults, toCode(results.At(i).Type()))
	}
	return jen.Params(jenParams...).Params(jenResults...)
}

func toCodeNamed(t *types.Named) *jen.Statement {
	name := toCodeObj(t.Obj())

	args := t.TypeArgs()
	if args.Len() == 0 {
		return name
	}

	jenArgs := []jen.Code{}
	for i := 0; i < args.Len(); i++ {
		jenArgs = append(jenArgs, toCode(args.At(i)))
	}

	return name.Index(jen.List(jenArgs...))
}

func toCodeObj(obj types.Object) *jen.Statement {
	if obj.Pkg() == nil {
		return jen.Id(obj.Name())
	}
	return jen.Qual(obj.Pkg().Path(), obj.Name())
}

func toCodeStruct(t *types.Struct) *jen.Statement {
	fields := []jen.Code{}
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		tag := t.Tag(i)

		fieldType := toCode(f.Type())
		if tag != "" {
			fieldType = fieldType.Add(jen.Id("`" + tag + "`"))
		}

		if !f.Embedded() {
			fieldType = jen.Id(f.Name()).Add(fieldType)
		}

		fields = append(fields, fieldType)
	}

	return jen.Struct(fields...)
}

func toCodeBasic(t types.BasicKind) *jen.Statement {
	switch t {
	case types.String:
		return jen.String()
	case types.Int:
		return jen.Int()
	case types.Int8:
		return jen.Int8()
	case types.Int16:
		return jen.Int16()
	case types.Int32:
		return jen.Int32()
	case types.Int64:
		return jen.Int64()
	case types.Uint:
		return jen.Uint()
	case types.Uint8:
		return jen.Uint8()
	case types.Uint16:
		return jen.Uint16()
	case types.Uint32:
		return jen.Uint32()
	case types.Uint64:
		return jen.Uint64()
	case types.Bool:
		return jen.Bool()
	case types.Complex128:
		return jen.Complex128()
	case types.Complex64:
		return jen.Complex64()
	case types.Float32:
		return jen.Float32()
	case types.Float64:
		return jen.Float64()
	default:
		panic(fmt.Sprintf("unsupported type %d", t))
	}
}
