package maker

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

// Method describes the code and documentation
// tied into a method
type Method struct {
	Name string
	Code string
	Docs []string
}

// declaredType identifies the name and package of a type declaration.
type declaredType struct {
	Name    string
	Package string
}

// Fullname returns a scoped Package.Name string out of this declaredType.
func (dt declaredType) Fullname() string {
	return fmt.Sprintf("%s.%s", dt.Package, dt.Name)
}

// Lines return a []string consisting of
// the documentation and code appended
// in chronological order
func (m *Method) Lines() []string {
	var lines []string
	lines = append(lines, m.Docs...)
	lines = append(lines, m.Code)
	return lines
}

// GetTypeDeclarationName extract the name of the type of this declaration if it refers to a type declaration.
// Otherwise, it returns an empty string.
func GetTypeDeclarationName(decl ast.Decl) string {
	gd, ok := decl.(*ast.GenDecl)
	if !ok {
		return ""
	}

	if gd.Tok != token.TYPE {
		return ""
	}

	typeName := ""
	for _, spec := range gd.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			return ""
		}
		typeName = typeSpec.Name.Name
	}

	return typeName
}

// GetReceiverTypeName takes in the entire
// source code and a single declaration.
// It then checks if the declaration is a
// function declaration, if it is, it uses
// the GetReceiverType to check whether
// the declaration is a method or a function
// if it is a function we fatally stop.
// If it is a method we retrieve the type
// of the receiver based on the types
// start and end pos in combination with
// the actual source code.
// It then returns the name of the
// receiver type and the function declaration
//
// Behavior is undefined for a src []byte that
// isn't the source of the possible FuncDecl fl
func GetReceiverTypeName(src []byte, fl ast.Decl) (string, *ast.FuncDecl) {
	fd, ok := fl.(*ast.FuncDecl)
	if !ok {
		return "", nil
	}
	t, err := GetReceiverType(fd)
	if err != nil {
		return "", nil
	}
	st := string(src[t.Pos()-1 : t.End()-1])
	if len(st) > 0 && st[0] == '*' {
		st = st[1:]
	}
	return st, fd
}

// GetReceiverType checks if the FuncDecl
// is a function or a method. If it is a
// function it returns a nil ast.Expr and
// a non-nil err. If it is a method it uses
// a hardcoded 0 index to fetch the receiver
// because a method can only have 1 receiver.
// Which can make you wonder why it is a
// list in the first place, but this type
// from the `ast` pkg is used in other
// places than for receivers
func GetReceiverType(fd *ast.FuncDecl) (ast.Expr, error) {
	if fd.Recv == nil {
		return nil, fmt.Errorf("fd is not a method, it is a function")
	}
	return fd.Recv.List[0].Type, nil
}

// reMatchTypename matches any of the following to extract the <type>:
//
//	*<type>
//	[]<type>
//	[]*<type>
//	map[<keyType>]<type>
//	map[<keyType>]*<type>
var reMatchTypename = regexp.MustCompile(`^(\[\]|\*|\[\]\*|map\[\w+\]|map\[\w+\]\*)(\w+)$`)

// FormatFieldList takes in the source code
// as a []byte and a FuncDecl parameters or
// return values as a FieldList.
// It then returns a []string with each
// param or return value as a single string.
// If the FieldList input is nil, it returns
// nil
func FormatFieldList(src []byte, fl *ast.FieldList, pkgName string, declaredTypes []declaredType) []string {
	if fl == nil {
		return nil
	}
	var parts []string
	for _, l := range fl.List {
		names := make([]string, len(l.Names))
		for i, n := range l.Names {
			names[i] = n.Name
		}
		t := string(src[l.Type.Pos()-1 : l.Type.End()-1])
		t2 := t
		// Try to match <modifier><type>. If matched variable `match` will look like this for t=="[]Category":
		// match[0][0] = "[]Category"
		// match[0][1] = "[]"
		// match[0][2] = "Category"
		match := reMatchTypename.FindAllStringSubmatch(t, -1)
		if match != nil {
			// Set `t` so it will compare correctly with `dt.Name` below
			t2 = match[0][2]
		}

		for _, dt := range declaredTypes {
			if t2 == dt.Name && pkgName != dt.Package {
				// The type of this field is the same as one declared in the source package,
				// and the source package is not the same as the destination package.
				if match != nil {
					// Add back `*`, `[]`, `[]*`, `map[<type>]` or `map[<type>]*` if there was a
					// match.
					t = match[0][1] + dt.Fullname()
				} else {
					t = dt.Fullname()
				}
			}
		}

		regexString := fmt.Sprintf(`(\*|\(|\s|^)%s\.`, regexp.QuoteMeta(pkgName))
		t = regexp.MustCompile(regexString).ReplaceAllString(t, "$1")

		if len(names) > 0 {
			typeSharingArgs := strings.Join(names, ", ")
			parts = append(parts, fmt.Sprintf("%s %s", typeSharingArgs, t))
		} else {
			parts = append(parts, t)
		}
	}
	return parts
}

// FormatCode sets the options of the imports
// pkg and then applies the Process method
// which by default removes all of the imports
// not used and formats the remaining docs,
// imports and code like `gofmt`. It will
// e.g. remove paranthesis around a unnamed
// single return type
func FormatCode(code string) ([]byte, error) {
	opts := &imports.Options{
		TabIndent: true,
		TabWidth:  2,
		Fragment:  true,
		Comments:  true,
	}
	return imports.Process("", []byte(code), opts)
}

// MakeInterface takes in all of the items
// required for generating the interface,
// it then simply concatenates them all
// to an array, joins this array to a string
// with newline and passes it on to FormatCode
// which then directly returns the result
func MakeInterface(comment, pkgName, ifaceName, ifaceComment string, methods []string, imports []string) ([]byte, error) {
	output := []string{
		"// " + comment,
		"",
		"package " + pkgName,
		"import (",
	}
	output = append(output, imports...)
	output = append(output,
		")",
		"",
	)
	if len(ifaceComment) > 0 {
		prefix := "// "
		if strings.HasPrefix(ifaceComment, "go:generate") {
			prefix = "//"
		}
		output = append(output, fmt.Sprintf("%s%s", prefix, strings.Replace(ifaceComment, "\n", "\n// ", -1)))
	}
	output = append(output, fmt.Sprintf("type %s interface {", ifaceName))
	output = append(output, methods...)
	output = append(output, "}")
	code := strings.Join(output, "\n")
	return FormatCode(code)
}

// ParseDeclaredTypes inspect given src code to find type declaractions.
func ParseDeclaredTypes(src []byte) (declaredTypes []declaredType) {
	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err.Error())
	}

	sourcePackageName := a.Name.Name

	name := ""
	for _, d := range a.Decls {
		name = GetTypeDeclarationName(d)
		if name != "" {
			declaredTypes = append(declaredTypes, declaredType{
				Name:    name,
				Package: sourcePackageName,
			})
		}
	}

	return
}

// ParseStruct takes in a piece of source code as a
// []byte, the name of the struct it should base the
// interface on and a bool saying whether it should
// include docs.  It then returns an []Method where
// Method contains the method declaration(not the code)
// that is required for the interface and any documentation
// if included.
// It also returns a []string containing all of the imports
// including their aliases regardless of them being used or
// not, the imports not used will be removed later using the
// 'imports' pkg If anything goes wrong, this method will
// fatally stop the execution
func ParseStruct(src []byte, structName string, copyDocs bool, copyTypeDocs bool, pkgName string, declaredTypes []declaredType, importModule string, withNotExported bool) (methods []Method, imports []string, typeDoc string) {
	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, i := range a.Imports {
		if i.Name != nil {
			imports = append(imports, fmt.Sprintf("%s %s", i.Name.String(), i.Path.Value))
		} else {
			imports = append(imports, i.Path.Value)
		}
	}

	if importModule != "" {
		imports = append(imports, fmt.Sprintf(". %s", strconv.Quote(importModule)))
	}

	for _, d := range a.Decls {
		if a, fd := GetReceiverTypeName(src, d); a == structName {
			if !withNotExported && !fd.Name.IsExported() {
				continue
			}
			params := FormatFieldList(src, fd.Type.Params, pkgName, declaredTypes)
			ret := FormatFieldList(src, fd.Type.Results, pkgName, declaredTypes)
			mName := fd.Name.String()
			method := fmt.Sprintf("%s(%s) (%s)", mName, strings.Join(params, ", "), strings.Join(ret, ", "))
			var docs []string
			if fd.Doc != nil && copyDocs {
				for _, d := range fd.Doc.List {
					docs = append(docs, string(src[d.Pos()-1:d.End()-1]))
				}
			}
			methods = append(methods, Method{
				Name: mName,
				Code: method,
				Docs: docs,
			})
		}
	}

	if copyTypeDocs {
		pkg := &ast.Package{Files: map[string]*ast.File{"": a}}
		doc := doc.New(pkg, "", doc.AllDecls)
		for _, t := range doc.Types {
			if t.Name == structName {
				typeDoc = strings.TrimSuffix(t.Doc, "\n")
			}
		}
	}

	return
}

// MakeOptions contains options for the Make function.
type MakeOptions struct {
	Files           []string
	StructType      string
	Comment         string
	PkgName         string
	IfaceName       string
	IfaceComment    string
	ImportModule    string
	CopyDocs        bool
	CopyTypeDoc     bool
	ExcludeMethods  []string
	WithNotExported bool
}

// validateStructType checks input struct type against the parsed declared
// types and returns true when present
func validateStructType(types []declaredType, stType string) bool {
	for _, v := range types {
		if strings.EqualFold(v.Name, stType) {
			return true
		}

	}
	return false

}

func Make(options MakeOptions) ([]byte, error) {
	var (
		allMethods       []string
		allImports       []string
		allDeclaredTypes []declaredType

		mset = make(map[string]struct{})
		iset = make(map[string]struct{})
		tset = make(map[string]struct{})
	)

	var typeDoc string

	// First pass on all files to find declared types
	for _, f := range options.Files {
		b, err := os.ReadFile(f)
		if err != nil {
			return []byte{}, err
		}
		types := ParseDeclaredTypes(b)

		// Track if we've seen the input Struct type
		for _, t := range types {
			if _, ok := tset[t.Fullname()]; !ok {
				allDeclaredTypes = append(allDeclaredTypes, t)
				tset[t.Fullname()] = struct{}{}
			}
		}
	}

	// Validate at least one file contains the input struct Type
	if !validateStructType(allDeclaredTypes, options.StructType) {
		return []byte{},
			fmt.Errorf("%q structtype not found in input files",
				options.StructType)
	}

	excludedMethods := make(map[string]struct{}, len(options.ExcludeMethods))
	for _, mName := range options.ExcludeMethods {
		excludedMethods[mName] = struct{}{}
	}

	// Second pass to build up the interface
	for _, f := range options.Files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		methods, imports, parsedTypeDoc := ParseStruct(src, options.StructType, options.CopyDocs, options.CopyTypeDoc, options.PkgName, allDeclaredTypes, options.ImportModule, options.WithNotExported)
		for _, m := range methods {
			if _, ok := excludedMethods[m.Name]; ok {
				continue
			}

			if _, ok := mset[m.Code]; !ok {
				allMethods = append(allMethods, m.Lines()...)
				mset[m.Code] = struct{}{}
			}
		}
		for _, i := range imports {
			if _, ok := iset[i]; !ok {
				allImports = append(allImports, i)
				iset[i] = struct{}{}
			}
		}
		if typeDoc == "" {
			typeDoc = parsedTypeDoc
		}
	}

	if typeDoc != "" {
		options.IfaceComment = fmt.Sprintf("%s\n%s", options.IfaceComment, typeDoc)
	}

	result, err := MakeInterface(options.Comment, options.PkgName, options.IfaceName, options.IfaceComment, allMethods, allImports)
	if err != nil {
		return nil, err
	}

	return result, nil
}
