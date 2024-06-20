package pkgload

import (
	"fmt"
	"go/types"
	"regexp"
	"strings"

	"github.com/jmattheis/goverter/method"
	"golang.org/x/tools/go/packages"
)

func New(workDir, buildTags string, paths []string) (*PackageLoader, error) {
	loader := &PackageLoader{
		lookup: map[string]*packages.Package{},
	}
	err := loader.load(workDir, buildTags, paths)
	return loader, err
}

type PackageLoader struct {
	lookup map[string]*packages.Package
}

func (g *PackageLoader) GetMatching(cwd, fullMethod string, opts *method.ParseOpts) ([]*method.Definition, error) {
	pkgName, name, err := ParseMethodString(cwd, fullMethod)
	if err != nil {
		return nil, err
	}

	pattern, err := regexp.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("could not parse name as regexp %q: %s", name, err)
	}

	if _, complete := pattern.LiteralPrefix(); complete {
		// not a regex
		m, err := g.getOneParsed(pkgName, name, opts)
		if err != nil {
			return nil, err
		}
		return []*method.Definition{m}, nil
	}

	// this is regexp, scan thru the package methods to find funcs that match the pattern
	var matches []*method.Definition

	pkg, err := g.getPkg(pkgName)
	if err != nil {
		return nil, err
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		loc := pattern.FindStringIndex(name)
		if len(loc) != 2 {
			continue
		}
		if loc[0] != 0 || loc[1] != len(name) {
			// we want full match only: e.g. CopyAbc.* won't match OtherCopyAbc
			continue
		}

		obj := scope.Lookup(name)
		m, err := method.Parse(obj, opts)
		if err == nil {
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf(`package %s does not have methods with names that match
the golang regexp pattern %q and a convert signature`, pkgName, name)
	}

	return matches, nil
}

func (g *PackageLoader) getPkg(pkgName string) (*packages.Package, error) {
	pkg := g.lookup[pkgName]
	if pkg == nil {
		return nil, fmt.Errorf("failed to load package %q:\nmake sure it's a valid golang package", pkgName)
	}

	if len(pkg.Errors) > 0 {
		var lines []string
		for _, err := range pkg.Errors {
			lines = append(lines, err.Error())
		}

		return nil, fmt.Errorf("failed to load package %q:\n%s", pkgName, strings.Join(lines, "\n"))
	}
	return pkg, nil
}

func (g *PackageLoader) GetOneRaw(pkgName, name string) (types.Object, error) {
	pkg, err := g.getPkg(pkgName)
	if err != nil {
		return nil, err
	}

	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("%q does not exist in package %q", name, pkgName)
	}
	return obj, nil
}

func (g *PackageLoader) GetOne(cwd, fullMethod string, opts *method.ParseOpts) (*method.Definition, error) {
	pkgName, name, err := ParseMethodString(cwd, fullMethod)
	if err != nil {
		return nil, err
	}
	return g.getOneParsed(pkgName, name, opts)
}

func (g *PackageLoader) getOneParsed(pkgName, name string, opts *method.ParseOpts) (*method.Definition, error) {
	obj, err := g.GetOneRaw(pkgName, name)
	if err != nil {
		return nil, err
	}

	def, err := method.Parse(obj, opts)
	if err != nil {
		return nil, err
	}
	return def, nil
}

// loadPackages is used to load extend packages, with caching support.
func (g *PackageLoader) load(workDir, buildTags string, paths []string) error {
	packagesCfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  workDir,
	}
	if buildTags != "" {
		packagesCfg.BuildFlags = append(packagesCfg.BuildFlags, "-tags", buildTags)
	}
	pkgs, err := packages.Load(packagesCfg, paths...)
	if err != nil {
		// This happens rare, and only if somebody uses advanced package pattern query in a wrong way.
		// The cause (err) usually has enough details to troubleshoot this issue.
		return fmt.Errorf("failed to load packages %s:\n%s", paths, err)
	}
	for _, pkg := range pkgs {
		g.lookup[pkg.PkgPath] = pkg
	}

	return nil
}
