package cli

import (
	"fmt"
	"os"

	"github.com/jmattheis/goverter"
	"github.com/jmattheis/goverter/enum"
)

type RunOpts struct {
	EnumTransformers map[string]enum.Transformer
}

// Run runs the goverter cli with the given args and customizations.
func Run(args []string, opts RunOpts) {
	cfg, err := Parse(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.EnumTransformers != nil {
		for key, value := range opts.EnumTransformers {
			cfg.EnumTransformers[key] = value
		}
	}

	if err = goverter.GenerateConverters(cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
