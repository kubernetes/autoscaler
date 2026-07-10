package config

import (
	"strings"

	"github.com/jmattheis/goverter/pkgload"
)

func getPackages(raw *Raw) []string {
	lookup := map[string]struct{}{}
	for _, c := range raw.Converters {
		lookup[c.Package] = struct{}{}
		registerConverterLines(lookup, c.Package, c.Converter)
		registerConverterLines(lookup, c.Package, raw.Global)
		for _, m := range c.Methods {
			registerMethodLines(lookup, c.Package, m)
		}
	}

	var pkgs []string
	for pkg := range lookup {
		pkgs = append(pkgs, "pattern="+pkg)
	}

	return pkgs
}

func registerConverterLines(lookup map[string]struct{}, cwd string, lines RawLines) {
	for _, line := range lines.Lines {
		cmd, rest := parseCommand(line)
		if cmd == configExtend {
			for _, fullMethod := range strings.Fields(rest) {
				registerFullMethod(lookup, cwd, fullMethod)
			}
		}
	}
}

func registerMethodLines(lookup map[string]struct{}, cwd string, lines RawLines) {
	for _, line := range lines.Lines {
		cmd, rest := parseCommand(line)
		switch cmd {
		case configMap:
			if _, _, custom, err := parseMethodMap(rest); err == nil && custom != "" {
				registerFullMethod(lookup, cwd, custom)
			}
		case configDefault:
			registerFullMethod(lookup, cwd, rest)
		}
	}
}

func registerFullMethod(lookup map[string]struct{}, cwd, fullMethod string) {
	pkg, _, err := pkgload.ParseMethodString(cwd, fullMethod)
	if err == nil {
		lookup[pkg] = struct{}{}
	}
}
