package printer

import (
	"io"
	"unicode"
)

// FieldType describes the type of field in a list (struct field, param field, result field, etc.)
type FieldType int

// BlockType describes the type of block (code block, const definition block, etc.)
type BlockType int

const (
	UP   = +1
	DOWN = -1

	METHOD FieldType = iota
	FIELD
	RECEIVER
	PARAM
	RESULT

	CODE BlockType = iota
	CONST
	VAR
	TYPE
	STRUCT
	INTERFACE

	CHAN_BIDI = "chan"
	CHAN_SEND = "chan<-"
	CHAN_RECV = "<-chan"

	NONE    = ""
	NL      = "\n"
	SEMI    = ";\n"
	COLON   = ":\n"
	COMMA   = ", "
	COMMANL = ",\n"
)

//
// Printer is the interface to be implemented to print a program
//
type Printer interface {
	Reset()

	SetWriter(w io.Writer)
	UpdateLevel(delta int)
	SameLine()
	IsSameLine() bool
	Print(values ...string)
	PrintLevel(term string, values ...string)
	Chop(line string) string

	PushContext()
	PopContext()

	// print start block "{"
	PrintBlockStart(b BlockType)

	// print end block "}"
	PrintBlockEnd(b BlockType)

	// print the package name
	PrintPackage(name string)

	// print a single import
	PrintImport(name, path string)

	// print a type definition
	PrintType(name, typedef string)

	// print a const/var definition
	PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool)

	// print a 'special' statement (goto, break, continue, ...)
	PrintStmt(stmt, expr string)

	// print return statemement
	PrintReturn(expr string, tuple bool)

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

	// print a "case" closing statement (break, if needed)
	PrintEndCase()

	// print an "if" opening statement
	PrintIf(init, cond string)

	// print the "else" statement
	PrintElse()

	// print an empty statement
	PrintEmpty()

	// print an assignment statement
	PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool)

	// print a channel send statement
	PrintSend(ch, value string)

	////////////////////////////////////

	FormatIdent(id string) string

	FormatLiteral(lit string) string

	FormatCompositeLit(typedef, elt string) string

	FormatStar(expr string) string

	FormatEllipsis(expr string) string

	FormatParen(expr string) string

	FormatUnary(op, operand string) string

	FormatBinary(lhs, op, rhs string) string

	FormatPair(p Pair, t FieldType) string

	FormatArray(len, elt string) string

	FormatArrayIndex(array, index string) string

	FormatSlice(slice, low, high, max string) string

	FormatMap(key, elt string) string

	FormatKeyValue(key, value string) string

	FormatStruct(fields string) string

	FormatInterface(methods string) string

	FormatChan(chdir, mtype string) string

	FormatCall(fun, args string, isFuncLit bool) string

	FormatFuncType(params, results string, withFunc bool) string

	FormatFuncLit(ftype, body string) string

	FormatSelector(pname, sel string, isObject bool) string

	FormatTypeAssert(orig, assert string) string
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
// Default format for a Pair ("name" SP "value")
//
func (p Pair) String() string {
	if len(p.Name()) > 0 && len(p.Value()) > 0 {
		return p.Name() + " " + p.Value()
	} else {
		return p.Name() + p.Value()
	}
}

//
// Is this name public
//
func IsPublic(name string) bool {
	//
	// get the first "rune" and check if is UpperCase
	//
	for _, rune := range name {
		return unicode.IsUpper(rune)
	}

	// shouldn't get here, unless the string is empty
	return false
}
