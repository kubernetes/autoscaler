package config

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/jmattheis/goverter/method"
)

const (
	configMap     = "map"
	configDefault = "default"
)

type Method struct {
	*method.Definition
	Common

	Constructor *method.Definition
	AutoMap     []string
	Fields      map[string]*FieldMapping
	EnumMapping *EnumMapping

	RawFieldSettings []string
}

type FieldMapping struct {
	Source   string
	Function *method.Definition
	Ignore   bool
}

func (m *Method) Field(targetName string) *FieldMapping {
	target, ok := m.Fields[targetName]
	if !ok {
		target = &FieldMapping{}
		m.Fields[targetName] = target
	}
	return target
}

func parseMethod(ctx *context, c *Converter, fn *types.Func, rawMethod RawLines) (*Method, error) {
	def, err := method.Parse(fn, &method.ParseOpts{
		ErrorPrefix:       "error parsing converter method",
		Converter:         nil,
		OutputPackagePath: c.OutputPackagePath,
		Params:            method.ParamsRequired,
		ConvFunction:      true,
	})
	if err != nil {
		return nil, err
	}

	m := &Method{
		Definition:  def,
		Common:      c.Common,
		Fields:      map[string]*FieldMapping{},
		EnumMapping: &EnumMapping{Map: map[string]string{}},
	}

	for _, value := range rawMethod.Lines {
		if err := parseMethodLine(ctx, c, m, value); err != nil {
			return m, formatLineError(rawMethod, fn.String(), value, err) // TODO get method type
		}
	}
	return m, nil
}

func parseMethodLine(ctx *context, c *Converter, m *Method, value string) (err error) {
	cmd, rest := parseCommand(value)
	fieldSetting := false
	switch cmd {
	case configMap:
		fieldSetting = true
		var source, target, custom string
		source, target, custom, err = parseMethodMap(rest)
		if err != nil {
			return err
		}
		f := m.Field(target)
		f.Source = source

		if custom != "" {
			opts := &method.ParseOpts{
				ErrorPrefix:       "error parsing type",
				OutputPackagePath: c.OutputPackagePath,
				Converter:         c.Type,
				Params:            method.ParamsOptional,
				AllowTypeParams:   true,
			}
			f.Function, err = ctx.Loader.GetOne(c.Package, custom, opts)
		}
	case "ignore":
		fieldSetting = true
		fields := strings.Fields(rest)
		for _, f := range fields {
			m.Field(f).Ignore = true
		}
	case "enum:map":
		fields := strings.Fields(rest)
		if len(fields) != 2 {
			return fmt.Errorf("invalid fields")
		}

		if IsEnumAction(fields[1]) {
			err = validateEnumAction(fields[1])
		}

		m.EnumMapping.Map[fields[0]] = fields[1]
	case "enum:transform":
		fields := strings.SplitN(rest, " ", 2)

		config := ""
		if len(fields) == 2 {
			config = fields[1]
		}

		var t ConfiguredTransformer
		t, err = parseTransformer(ctx, fields[0], config)
		m.EnumMapping.Transformers = append(m.EnumMapping.Transformers, t)
	case "autoMap":
		fieldSetting = true
		var s string
		s, err = parseString(rest)
		m.AutoMap = append(m.AutoMap, strings.TrimSpace(s))
	case configDefault:
		opts := &method.ParseOpts{
			ErrorPrefix:       "error parsing type",
			OutputPackagePath: c.OutputPackagePath,
			Converter:         c.Type,
			Params:            method.ParamsOptional,
			AllowTypeParams:   true,
		}
		m.Constructor, err = ctx.Loader.GetOne(c.Package, rest, opts)
	default:
		fieldSetting, err = parseCommon(&m.Common, cmd, rest)
	}
	if fieldSetting {
		m.RawFieldSettings = append(m.RawFieldSettings, value)
	}
	return err
}

func parseMethodMap(remaining string) (source, target, custom string, err error) {
	parts := strings.SplitN(remaining, "|", 2)
	if len(parts) == 2 {
		custom = strings.TrimSpace(parts[1])
	}

	fields := strings.Fields(parts[0])
	switch len(fields) {
	case 1:
		target = fields[0]
	case 2:
		source = fields[0]
		target = fields[1]
	case 0:
		err = fmt.Errorf("missing target field")
	default:
		err = fmt.Errorf("too many fields expected at most 2 fields got %d: %s", len(fields), remaining)
	}
	if err == nil && strings.ContainsRune(target, '.') {
		err = fmt.Errorf("the mapping target %q must be a field name but was a path.\nDots \".\" are not allowed.", target)
	}
	return source, target, custom, err
}
