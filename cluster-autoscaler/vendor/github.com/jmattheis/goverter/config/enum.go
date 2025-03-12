package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jmattheis/goverter/enum"
	"github.com/jmattheis/goverter/pkgload"
)

const (
	EnumActionPanic  = "@panic"
	EnumActionError  = "@error"
	EnumActionIgnore = "@ignore"
)

type EnumMapping struct {
	Transformers []ConfiguredTransformer
	Map          map[string]string
}

type ConfiguredTransformer struct {
	Name        string
	Transformer enum.Transformer
	Config      string
}

func parseTransformer(ctx *context, name, config string) (ConfiguredTransformer, error) {
	t, ok := ctx.EnumTransformers[name]
	if !ok {
		t, ok = enum.DefaultTransformers[name]
	}

	if !ok {
		return ConfiguredTransformer{}, fmt.Errorf("transformer %q does not exist", name)
	}

	return ConfiguredTransformer{Name: name, Transformer: t, Config: config}, nil
}

func IsEnumAction(s string) bool {
	return strings.HasPrefix(s, "@")
}

func validateEnumAction(s string) error {
	switch s {
	case EnumActionPanic, EnumActionError, EnumActionIgnore:
		return nil
	default:
		return fmt.Errorf("invalid enum action %q, must be one of %q, %q, or %q", s, EnumActionPanic, EnumActionIgnore, EnumActionError)
	}
}

func parseIDPattern(cwd, rest string) (pattern enum.IDPattern, err error) {
	path, name, err := pkgload.ParseMethodString(cwd, rest)
	if err != nil {
		return pattern, err
	}

	pattern.Path, err = regexp.Compile(path)
	if err != nil {
		return pattern, err
	}
	pattern.Name, err = regexp.Compile(name)
	if err != nil {
		return pattern, err
	}
	return pattern, nil
}
