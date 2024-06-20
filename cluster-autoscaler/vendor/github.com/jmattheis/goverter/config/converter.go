package config

import (
	"go/types"
	"strings"

	"github.com/jmattheis/goverter/enum"
	"github.com/jmattheis/goverter/method"
)

const (
	configExtend = "extend"
)

var DefaultConfig = ConverterConfig{
	OutputFile:        "./generated/generated.go",
	OutputPackageName: "generated",
	Common:            Common{Enum: enum.Config{Enabled: true}},
}

type Converter struct {
	ConverterConfig
	Package    string
	FileSource string
	Type       types.Type
	Methods    map[string]*Method
}

type ConverterConfig struct {
	Common
	Name              string
	OutputFile        string
	OutputPackagePath string
	OutputPackageName string
	Extend            []*method.Definition
	Comments          []string
}

func (conf *ConverterConfig) PackageID() string {
	if conf.OutputPackageName == "" {
		return conf.OutputPackagePath
	}
	return conf.OutputPackagePath + ":" + conf.OutputPackageName
}

func parseGlobal(ctx *context, global RawLines) (*ConverterConfig, error) {
	c := Converter{ConverterConfig: DefaultConfig}
	err := parseConverterLines(ctx, &c, "global", global)
	return &c.ConverterConfig, err
}

func parseConverter(ctx *context, rawConverter *RawConverter, global ConverterConfig) (*Converter, error) {
	v, err := ctx.Loader.GetOneRaw(rawConverter.Package, rawConverter.InterfaceName)
	if err != nil {
		return nil, err
	}
	namedType := v.Type()
	interfaceType := namedType.Underlying().(*types.Interface)

	c := &Converter{
		ConverterConfig: global,
		Type:            namedType,
		FileSource:      rawConverter.FileSource,
		Package:         rawConverter.Package,
		Methods:         map[string]*Method{},
	}
	if c.Name == "" {
		c.Name = rawConverter.InterfaceName + "Impl"
	}

	if err := parseConverterLines(ctx, c, c.Type.String(), rawConverter.Converter); err != nil {
		return nil, err
	}

	for i := 0; i < interfaceType.NumMethods(); i++ {
		fun := interfaceType.Method(i)
		def, err := parseMethod(ctx, c, fun, rawConverter.Methods[fun.Name()])
		if err != nil {
			return nil, err
		}
		c.Methods[fun.Name()] = def
	}

	return c, nil
}

func parseConverterLines(ctx *context, c *Converter, source string, raw RawLines) error {
	for _, value := range raw.Lines {
		if err := parseConverterLine(ctx, c, value); err != nil {
			return formatLineError(raw, source, value, err)
		}
	}

	return nil
}

func parseConverterLine(ctx *context, c *Converter, value string) (err error) {
	cmd, rest := parseCommand(value)
	switch cmd {
	case "converter":
		// only a marker interface
	case "name":
		c.Name, err = parseString(rest)
	case "output:file":
		c.OutputFile, err = parseString(rest)
	case "output:package":
		c.OutputPackageName = ""
		var pkg string
		pkg, err = parseString(rest)

		parts := strings.SplitN(pkg, ":", 2)
		switch len(parts) {
		case 2:
			c.OutputPackageName = parts[1]
			fallthrough
		case 1:
			c.OutputPackagePath = parts[0]
		}
	case "struct:comment":
		c.Comments = append(c.Comments, rest)
	case "enum:exclude":
		var pattern enum.IDPattern
		pattern, err = parseIDPattern(c.Package, rest)
		c.Enum.Excludes = append(c.Enum.Excludes, pattern)
	case configExtend:
		for _, name := range strings.Fields(rest) {
			opts := &method.ParseOpts{
				ErrorPrefix:       "error parsing type",
				OutputPackagePath: c.OutputPackagePath,
				Converter:         c.Type,
				Params:            method.ParamsRequired,
			}
			var defs []*method.Definition
			defs, err = ctx.Loader.GetMatching(c.Package, name, opts)
			if err != nil {
				break
			}
			c.Extend = append(c.Extend, defs...)
		}
	default:
		_, err = parseCommon(&c.Common, cmd, rest)
	}
	return err
}
