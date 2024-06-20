package builder

import "github.com/dave/jennifer/jen"

type ErrorPath []ErrorElement

func (e ErrorPath) WrapErrors(errStmt *jen.Statement) *jen.Statement {
	if len(e) != 0 {
		switch elm := e[len(e)-1].(type) {
		case errElmField:
			return jen.Qual("fmt", "Errorf").Call(jen.Lit("error setting field "+string(elm)+": %w"), errStmt)
		case errElmIndex:
			return jen.Qual("fmt", "Errorf").Call(jen.Lit("error setting index %d: %w"), elm.stmt.Clone(), errStmt)
		}
	}
	return errStmt
}

func (e ErrorPath) WrapErrorsUsing(pkg string, errStmt *jen.Statement) *jen.Statement {
	args := []jen.Code{}
	for _, elm := range e {
		switch elm := elm.(type) {
		case errElmField:
			args = append(args, jen.Qual(pkg, "Field").Call(jen.Lit(string(elm))))
		case errElmIndex:
			args = append(args, jen.Qual(pkg, "Index").Call(elm.stmt.Clone()))
		case errElmKey:
			args = append(args, jen.Qual(pkg, "Key").Call(elm.stmt.Clone()))
		}
	}
	args = append([]jen.Code{errStmt}, args...)

	return jen.Qual(pkg, "Wrap").Call(args...)
}

func (e ErrorPath) Index(code *jen.Statement) ErrorPath { return append(e, errElmIndex{code}) }
func (e ErrorPath) Key(code *jen.Statement) ErrorPath   { return append(e, errElmKey{code}) }
func (e ErrorPath) Field(name string) ErrorPath         { return append(e, errElmField(name)) }

type ErrorElement interface{ _elm() }

type (
	errElmIndex struct{ stmt *jen.Statement }
	errElmKey   struct{ stmt *jen.Statement }
	errElmField string
)

func (errElmKey) _elm()   {}
func (errElmIndex) _elm() {}
func (errElmField) _elm() {}
