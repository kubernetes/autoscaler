package generator

import (
	"github.com/jmattheis/goverter/config"
	"github.com/jmattheis/goverter/method"
	"github.com/jmattheis/goverter/namer"
	"github.com/jmattheis/goverter/xtype"
)

func setupGenerator(converter *config.Converter) *generator {
	extend := map[xtype.Signature]*method.Definition{}
	for _, def := range converter.Extend {
		extend[def.Signature()] = def
	}

	lookup := map[xtype.Signature]*generatedMethod{}
	for _, method := range converter.Methods {
		lookup[method.Definition.Signature()] = &generatedMethod{
			Method:   method,
			Dirty:    true,
			Explicit: true,
		}
	}

	gen := generator{
		namer:  namer.New(),
		conf:   converter,
		lookup: lookup,
		extend: extend,
	}

	return &gen
}
