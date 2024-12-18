/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

type flagInfo struct {
	Name         string
	DefaultValue string
	Type         string
	Description  string
	SourceFile   string
}

type templateData struct {
	Flags     []flagInfo
	Timestamp string
}

var componentTemplates = map[string]string{
	"recommender": `# What are the parameters to VPA recommender?
This document is auto-generated from the flag definitions in the VPA recommender code.
Last updated: {{ .Timestamp }}

| Flag | Type | Default | Description |
|------|------|---------|-------------|
{{- range .Flags }}
| --{{ .Name }} | {{ .Type }} | {{ .DefaultValue }} | {{ .Description }} |
{{- end }}
`,
	"updater": `# What are the parameters to VPA updater?
This document is auto-generated from the flag definitions in the VPA updater code.
Last updated: {{ .Timestamp }}

| Flag | Type | Default | Description |
|------|------|---------|-------------|
{{- range .Flags }}
| --{{ .Name }} | {{ .Type }} | {{ .DefaultValue }} | {{ .Description }} |
{{- end }}
`,
	"admission": `# What are the parameters to VPA admission controller?
This document is auto-generated from the flag definitions in the VPA admission controller code.
Last updated: {{ .Timestamp }}

| Flag | Type | Default | Description |
|------|------|---------|-------------|
{{- range .Flags }}
| --{{ .Name }} | {{ .Type }} | {{ .DefaultValue }} | {{ .Description }} |
{{- end }}
`,
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Generates markdown documentation for VPA recommender flags.")
		fmt.Println("usage: vpa-recommender-generate-flags <source-dir> <output-file>")
		fmt.Println("  <source-dir>  - Path to the directory containing the Go source files")
		fmt.Println("  <output-file> - Path to the output markdown file")
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	outputFile := os.Args[2]

	flags, err := collectFlags(sourceDir)
	if err != nil {
		log.Fatalf("Error collecting flags: %v", err)
	}

	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})

	err = generateDocs(flags, outputFile)
	if err != nil {
		log.Fatalf("Error generating documentation: %v", err)
	}

	fmt.Println("Successfully generated documentation in", outputFile)
}

func extractFlagFromCall(call *ast.CallExpr, sourcePath string) *flagInfo {
	if len(call.Args) < 3 {
		return nil
	}

	// Extract flag name
	nameArg, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return nil
	}
	name := strings.Trim(nameArg.Value, "\"")

	// Extract default value
	var defaultValue string
	switch v := call.Args[1].(type) {
	case *ast.BasicLit:
		defaultValue = strings.Trim(v.Value, "\"")
	case *ast.BinaryExpr:
		// Handle Duration expressions like "1*time.Minute"
		if lit, ok := v.X.(*ast.BasicLit); ok {
			if sel, ok := v.Y.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "time" {
					number := strings.Trim(lit.Value, "\"")
					unit := sel.Sel.Name
					defaultValue = fmt.Sprintf("%s%s", number, unit)
				}
			}
		}
	case *ast.UnaryExpr:
		// Handle negative numbers
		if lit, ok := v.X.(*ast.BasicLit); ok {
			defaultValue = fmt.Sprintf("%s%s", v.Op.String(), lit.Value)
		}
	case *ast.Ident:
		defaultValue = v.Name
	default:
		defaultValue = "0"
	}

	// Extract description
	descArg, ok := call.Args[2].(*ast.BasicLit)
	if !ok {
		return nil
	}
	description := strings.Trim(descArg.Value, "\"")
	description = strings.ReplaceAll(description, "\n\t\t", " ")
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.TrimSpace(description)

	return &flagInfo{
		Name:         name,
		DefaultValue: defaultValue,
		Type:         call.Fun.(*ast.SelectorExpr).Sel.Name,
		Description:  description,
		SourceFile:   sourcePath,
	}
}

func collectFlags(sourceDir string) ([]flagInfo, error) {
	var flags []flagInfo

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") || info.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)

		if err != nil {
			return fmt.Errorf("error parsing file %s: %v", path, err)
		}

		fileFlags := extractFlags(file, getRelativePath(sourceDir, path))
		flags = append(flags, fileFlags...)

		return nil
	})

	return flags, err
}

func extractFlags(file *ast.File, sourcePath string) []flagInfo {
	var flags []flagInfo

	ast.Inspect(file, func(n ast.Node) bool {
		// Look for flag.String(), flag.Bool(), flag.Float64(), etc.
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "flag" {
						if flag := extractFlagFromCall(call, sourcePath); flag != nil {
							flags = append(flags, *flag)
						}
					}
				}
			}
		}
		return true
	})

	return flags
}

func generateDocs(flags []flagInfo, outputFile string) error {
	component := determineComponent(outputFile)
	markdownTemplate, ok := componentTemplates[component]
	if !ok {
		return fmt.Errorf("couldn't find template for %s ", component)
	}
	tmpl, err := template.New("flags").Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	data := templateData{
		Flags:     flags,
		Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer f.Close()

	_, err = io.Copy(f, &buf)
	return err
}

func getRelativePath(baseDir, fullPath string) string {
	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return fullPath
	}
	return rel
}

// We determine the component by the path of the output file.
// if the path contains "recommender", we generate recommender flags.
// if the path contains "updater", we generate updater flags.
// if the path contains "admission", we generate admission flags.
// Otherwise, we generate recommender flags.

func determineComponent(outputPath string) string {
	if strings.Contains(outputPath, "updater") {
		return "updater"
	}
	if strings.Contains(outputPath, "admission") {
		return "admission"
	}
	if strings.Contains(outputPath, "recommender") {
		return "recommender"
	}
	fmt.Println("Couldn't find component. default is recommender")
	return "recommender"
}
