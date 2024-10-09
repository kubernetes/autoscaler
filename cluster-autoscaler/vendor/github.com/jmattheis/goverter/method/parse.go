package method

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

type ParamType int

const (
	ParamsRequired ParamType = iota
	ParamsOptional
	ParamsNone
)

type ParseOpts struct {
	Converter         types.Type
	OutputPackagePath string

	ErrorPrefix     string
	Params          ParamType
	ConvFunction    bool
	AllowTypeParams bool
}

// Parse parses an function into a Definition.
func Parse(obj types.Object, opts *ParseOpts) (*Definition, error) {
	formatErr := func(s string) error {
		return fmt.Errorf("%s:\n    %s\n\n%s", opts.ErrorPrefix, obj.String(), s)
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return nil, formatErr("must be a function")
	}

	if !xtype.Accessible(fn, opts.OutputPackagePath) {
		return nil, formatErr("must be exported")
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil, formatErr("must have a signature")
	}
	if sig.Results().Len() == 0 || sig.Results().Len() > 2 {
		return nil, formatErr("must have one or two returns")
	}
	returnError := false
	if sig.Results().Len() == 2 {
		if i, ok := sig.Results().At(1).Type().(*types.Named); ok && i.Obj().Name() == "error" && i.Obj().Pkg() == nil {
			returnError = true
		} else {
			return nil, formatErr("must have type error as second return but has: " + sig.Results().At(1).Type().String())
		}
	}

	methodDef := &Definition{
		ID:       fn.String(),
		OriginID: fn.String(),
		Parameters: Parameters{
			ReturnError: returnError,
			Target:      xtype.TypeOf(sig.Results().At(0).Type()),
			TypeParams:  sig.TypeParams().Len() > 0,
		},
		Name: fn.Name(),
	}

	if methodDef.TypeParams && !opts.AllowTypeParams {
		return nil, formatErr("must not be generic")
	}

	if opts.ConvFunction {
		methodDef.Call = jen.Id(xtype.ThisVar).Dot(fn.Name())
	} else {
		methodDef.Call = jen.Qual(fn.Pkg().Path(), fn.Name())
	}

	if opts.Params == ParamsNone && sig.Params().Len() > 0 {
		return nil, formatErr("must have no parameters")
	}

	switch sig.Params().Len() {
	case 2:
		if opts.Converter == nil {
			// converterInterface is used when searching for methods in the local package only
			return nil, formatErr("must have one parameter when using extend with a package")
		}

		actual := sig.Params().At(0).Type().String()
		if actual != opts.Converter.String() {
			return nil, formatErr(
				fmt.Sprintf("first parameter must be of type %s but was %s when having two parameters", opts.Converter.String(), actual))
		}
		methodDef.Parameters.SelfAsFirstParameter = true
		methodDef.Parameters.Source = xtype.TypeOf(sig.Params().At(1).Type())
	case 1:
		methodDef.Parameters.Source = xtype.TypeOf(sig.Params().At(0).Type())
	case 0:
		if opts.Params == ParamsRequired {
			return nil, formatErr("must have at least one parameter")
		}
	default:
		return nil, formatErr("must have one or two parameters")
	}

	return methodDef, nil
}
