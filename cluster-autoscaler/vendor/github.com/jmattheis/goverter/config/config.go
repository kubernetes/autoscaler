package config

import (
	"fmt"
	"sort"

	"github.com/jmattheis/goverter/enum"
	"github.com/jmattheis/goverter/pkgload"
)

type RawLines struct {
	Location string
	Lines    []string
}

type RawConverter struct {
	Package       string
	InterfaceName string
	Converter     RawLines
	Methods       map[string]RawLines
	FileSource    string
}

type Raw struct {
	Converters []RawConverter
	Global     RawLines

	WorkDir              string
	BuildTags            string
	OuputBuildConstraint string

	EnumTransformers map[string]enum.Transformer
}

type context struct {
	Loader *pkgload.PackageLoader

	EnumTransformers map[string]enum.Transformer
}

func Parse(raw *Raw) ([]*Converter, error) {
	loader, err := pkgload.New(raw.WorkDir, raw.BuildTags, getPackages(raw))
	if err != nil {
		return nil, err
	}

	ctx := &context{Loader: loader, EnumTransformers: raw.EnumTransformers}

	global, err := parseGlobal(ctx, raw.Global)
	if err != nil {
		return nil, err
	}

	converters := []*Converter{}
	for _, rawConverter := range raw.Converters {
		converter, err := parseConverter(ctx, &rawConverter, *global)
		if err != nil {
			return nil, err
		}
		converters = append(converters, converter)
	}

	sort.Slice(converters, func(i, j int) bool {
		return converters[i].Name < converters[j].Name
	})

	return converters, nil
}

func formatLineError(lines RawLines, t, value string, err error) error {
	cmd, _ := parseCommand(value)
	msg := `error parsing 'goverter:%s' at
    %s
    %s

%s`
	return fmt.Errorf(msg, cmd, lines.Location, t, err)
}
