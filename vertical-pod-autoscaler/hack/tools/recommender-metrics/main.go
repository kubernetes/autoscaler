/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command recommender-metrics generates documentation from registered metric definitions.
package main

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type source struct {
	values map[string]ast.Expr
	funcs  map[string]*ast.FuncDecl
}

type metric struct {
	name, kind, help string
	labels           []string
}

type generator struct{ sources map[string]*source }

func load(dir string) (*source, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	s := &source{values: map[string]ast.Expr{}, funcs: map[string]*ast.FuncDecl{}}
	packageName := ""
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		match, err := build.Default.MatchFile(dir, entry.Name())
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		if packageName != "" && packageName != file.Name.Name {
			return nil, fmt.Errorf("expected one package in %s", dir)
		}
		packageName = file.Name.Name
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if v, ok := spec.(*ast.ValueSpec); ok {
						for i, name := range v.Names {
							if i < len(v.Values) {
								s.values[name.Name] = v.Values[i]
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					s.funcs[d.Name.Name] = d
				}
			default:
				// Other declarations do not define metric constructors or values.
			}
		}
	}
	if packageName == "" {
		return nil, fmt.Errorf("no Go source files in %s", dir)
	}
	return s, nil
}

// text accepts only statically resolvable strings. Unsupported definitions fail
// generation rather than silently producing incomplete documentation.
func (g *generator) text(pkg string, expr ast.Expr, bindings map[string]string, depth int) (string, error) {
	if depth > 32 {
		return "", errors.New("cyclic or overly nested string definition")
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strconv.Unquote(e.Value)
		}
	case *ast.ParenExpr:
		return g.text(pkg, e.X, bindings, depth+1)
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			a, err := g.text(pkg, e.X, bindings, depth+1)
			if err != nil {
				return "", err
			}
			b, err := g.text(pkg, e.Y, bindings, depth+1)
			return a + b, err
		}
	case *ast.Ident:
		if value, ok := bindings[e.Name]; ok {
			return value, nil
		}
		if s := g.sources[pkg]; s != nil {
			if value, ok := s.values[e.Name]; ok {
				return g.text(pkg, value, nil, depth+1)
			}
		}
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			return g.text(id.Name, e.Sel, nil, depth+1)
		}
	default:
		// Unsupported expressions are reported below.
	}
	return "", fmt.Errorf("unsupported string expression %T in %s", expr, pkg)
}

func (g *generator) parseMetric(pkg string, expr ast.Expr, bindings map[string]string, depth int) (metric, error) {
	if depth > 16 {
		return metric{}, errors.New("cyclic or overly nested metric definition")
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return metric{}, fmt.Errorf("expected metric constructor, got %T", expr)
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return metric{}, errors.New("expected qualified constructor")
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return metric{}, errors.New("unsupported constructor qualifier")
	}
	if id.Name != "prometheus" {
		s := g.sources[id.Name]
		if s == nil {
			return metric{}, fmt.Errorf("unsupported helper package %s", id.Name)
		}
		fn := s.funcs[sel.Sel.Name]
		if fn == nil || fn.Body == nil || len(fn.Body.List) != 1 {
			return metric{}, fmt.Errorf("helper %s must contain a single return statement", sel.Sel.Name)
		}
		ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return metric{}, errors.New("unsupported helper return")
		}
		params := map[string]string{}
		index := 0
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				if index >= len(call.Args) {
					return metric{}, errors.New("missing helper argument")
				}
				value, err := g.text(pkg, call.Args[index], bindings, 0)
				if err != nil {
					return metric{}, err
				}
				params[name.Name] = value
				index++
			}
		}
		if index != len(call.Args) {
			return metric{}, errors.New("unexpected helper arguments")
		}
		return g.parseMetric(id.Name, ret.Results[0], params, depth+1)
	}
	kinds := map[string]string{"NewGauge": "Gauge", "NewGaugeVec": "Gauge", "NewCounter": "Counter", "NewCounterVec": "Counter", "NewHistogram": "Histogram", "NewHistogramVec": "Histogram", "NewSummary": "Summary", "NewSummaryVec": "Summary"}
	kind, ok := kinds[sel.Sel.Name]
	if !ok {
		return metric{}, fmt.Errorf("unsupported Prometheus constructor %s", sel.Sel.Name)
	}
	vector := strings.HasSuffix(sel.Sel.Name, "Vec")
	want := 1
	if vector {
		want = 2
	}
	if len(call.Args) != want {
		return metric{}, fmt.Errorf("unexpected arguments to %s", sel.Sel.Name)
	}
	opts, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return metric{}, errors.New("expected literal metric options")
	}
	fields := map[string]string{}
	for _, element := range opts.Elts {
		kv, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return metric{}, errors.New("expected named metric options")
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			return metric{}, errors.New("unsupported options key")
		}
		switch key.Name {
		case "Namespace", "Subsystem", "Name", "Help":
			value, err := g.text(pkg, kv.Value, bindings, 0)
			if err != nil {
				return metric{}, err
			}
			fields[key.Name] = value
		case "ConstLabels":
			return metric{}, errors.New("constant labels are not supported")
		default:
			// Buckets and other options do not affect the documentation columns.
		}
	}
	if fields["Name"] == "" || fields["Help"] == "" {
		return metric{}, errors.New("metric name and help are required")
	}
	parts := []string{}
	for _, field := range []string{"Namespace", "Subsystem", "Name"} {
		if fields[field] != "" {
			parts = append(parts, fields[field])
		}
	}
	m := metric{name: strings.Join(parts, "_"), kind: kind, help: fields["Help"]}
	if vector {
		labels, ok := call.Args[1].(*ast.CompositeLit)
		if !ok {
			return metric{}, errors.New("expected literal label names")
		}
		for _, element := range labels.Elts {
			value, err := g.text(pkg, element, bindings, 0)
			if err != nil {
				return metric{}, err
			}
			m.labels = append(m.labels, value)
		}
	}
	return m, nil
}

func (g *generator) metrics() ([]metric, error) {
	s := g.sources["recommender"]
	register := s.funcs["Register"]
	if register == nil || register.Body == nil {
		return nil, errors.New("registration function not found")
	}
	var result []metric
	seen := map[string]bool{}
	for _, stmt := range register.Body.List {
		expr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return nil, errors.New("unsupported Register statement")
		}
		call, ok := expr.X.(*ast.CallExpr)
		if !ok {
			return nil, errors.New("unsupported Register call")
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil, errors.New("unsupported registration")
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "prometheus" || sel.Sel.Name != "MustRegister" || call.Ellipsis.IsValid() {
			return nil, errors.New("expected prometheus.MustRegister with explicit collectors")
		}
		for _, arg := range call.Args {
			name, ok := arg.(*ast.Ident)
			if !ok {
				return nil, errors.New("expected collector identifier")
			}
			m, err := g.parseMetric("recommender", s.values[name.Name], nil, 0)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name.Name, err)
			}
			if seen[m.name] {
				return nil, fmt.Errorf("duplicate metric %s", m.name)
			}
			seen[m.name] = true
			result = append(result, m)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("no registered metrics found")
	}
	slices.SortFunc(result, func(a, b metric) int { return cmp.Compare(a.name, b.name) })
	return result, nil
}

const startMarker = "<!-- BEGIN GENERATED METRICS -->"
const endMarker = "<!-- END GENERATED METRICS -->"

func escape(s string) string {
	return strings.NewReplacer("\\", "\\\\", "|", "\\|", "\r\n", " ", "\n", " ").Replace(s)
}

func render(metrics []metric) string {
	var b strings.Builder
	fmt.Fprintln(&b, startMarker)
	fmt.Fprintln(&b, "<!-- Generated by hack/update-recommender-metrics.sh. Do not edit this table. -->")
	fmt.Fprintln(&b, "| Metric name | Type | Labels | Description |")
	fmt.Fprintln(&b, "| --- | --- | --- | --- |")
	for _, m := range metrics {
		labels := "None"
		if len(m.labels) > 0 {
			quoted := []string{}
			for _, label := range m.labels {
				quoted = append(quoted, "`"+escape(label)+"`")
			}
			labels = strings.Join(quoted, ", ")
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", escape(m.name), m.kind, labels, escape(m.help))
	}
	fmt.Fprintln(&b, endMarker)
	return b.String()
}

func replaceTable(doc, table string) (string, error) {
	if strings.Count(doc, startMarker) != 1 || strings.Count(doc, endMarker) != 1 {
		return "", errors.New("expected exactly one generated metrics block")
	}
	start, end := strings.Index(doc, startMarker), strings.Index(doc, endMarker)
	if end < start {
		return "", errors.New("invalid generated metrics markers")
	}
	return doc[:start] + strings.TrimSuffix(table, "\n") + doc[end+len(endMarker):], nil
}

func run(root string, check bool) error {
	g := &generator{sources: map[string]*source{}}
	for pkg, dir := range map[string]string{"metrics": "pkg/utils/metrics", "recommender": "pkg/utils/metrics/recommender"} {
		s, err := load(filepath.Join(root, dir))
		if err != nil {
			return err
		}
		g.sources[pkg] = s
	}
	metrics, err := g.metrics()
	if err != nil {
		return err
	}
	path := filepath.Join(root, "docs/recommender-metrics.md")
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated, err := replaceTable(string(original), render(metrics))
	if err != nil {
		return err
	}
	if check {
		if !bytes.Equal(original, []byte(updated)) {
			return errors.New("recommender metrics documentation is out of date; run hack/update-recommender-metrics.sh")
		}
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0644)
}

func main() {
	root := flag.String("root", ".", "VPA source root")
	check := flag.Bool("check", false, "verify documentation without writing it")
	flag.Parse()
	if err := run(*root, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
