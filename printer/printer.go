package printer

import (
	"io"
)

type FieldType int

const (
	UP   = +1
	DOWN = -1

	METHOD FieldType = iota
	FIELD
	RECEIVER
	PARAM
	RESULT
)

//
// Printer is the interface to be implemented to print a program
//
type Printer interface {
	SetWriter(w io.Writer)
	UpdateLevel(delta int)
	SameLine()
	Print(values ...string)
	PrintLevel(values ...string)

	// print the package name
	PrintPackage(name string)

	// print a single import
	PrintImport(name, path string)

	// print a type definition
	PrintType(name, typedef string)

	// print a const/var definition
	PrintValue(vtype, names, typedef, values string)

	// print a 'special' statement (goto, break, continue, ...)
	PrintStmt(stmt, expr string)

	// print a function definition
	PrintFunc(receiver, name, params, results string)

	// print a "for" opening statement
	PrintFor(init, cond, post string)

	// print a "range" opening statement
	PrintRange(key, value, expr string)

	// print a "switch" opening statement
	PrintSwitch(init, expr string)

	// print a "case" opening statement
	PrintCase(expr string)

	// print an "if" opening statement
	PrintIf(init, cond string)

	// print the "else" statement
	PrintElse()

	// print an empty statement
	PrintEmpty()

	// print an assignment statement
	PrintAssignment(lhs, op, rhs string)

	// print a channel send statement
	PrintSend(ch, value string)

	////////////////////////////////////

	FormatIdent(id string) string

	FormatLiteral(lit string) string

	FormatUnary(op, operand string) string

	FormatBinary(lhs, op, rhs string) string

	FormatPair(p Pair, t FieldType) string

	FormatArray(len, elt string) string

	FormatMap(key, elt string) string

	FormatStruct(fields string) string

	FormatInterface(methods string) string

	FormatCall(fun, args string) string

	FormatFuncType(params, results string) string
}

//
// Pair contains a pair of values (name/value, name/type, etc.)
//
type Pair [2]string

//
// PairList is a list/slice of pair
//
type PairList []Pair

// Returns the "name" part of a pair
func (p Pair) Name() string {
	return p[0]
}

// Returns the "value" part of a pair
func (p Pair) Value() string {
	return p[1]
}

//
// Default format for a Pair ("name" + "value")
//
func (p Pair) String() string {
	if len(p.Name()) > 0 && len(p.Value()) > 0 {
		return p.Name() + " " + p.Value()
	} else {
		return p.Name() + p.Value()
	}
}
