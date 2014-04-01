package printer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	NIL  = "nil"
	NULL = "NULL"
	IOTA = "iota"
)

//
// C implement the Printer interface for C programs
//
type CPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer

	iota int // incremented when 'const n = iota' or 'const n' - XXX: need to add a way to reset it
}

func (p *CPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *CPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *CPrinter) SameLine() {
	p.sameline = true
}

func (p *CPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *CPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMA)
}

func (p *CPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *CPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *CPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *CPrinter) PrintLevelIn(term string, values ...string) {
	p.level -= 1
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
	p.level += 1
}

func (p *CPrinter) PrintBlockStart() {
	p.PrintLevel("{\n")
	p.UpdateLevel(UP)
}

func (p *CPrinter) PrintBlockEnd() {
	p.UpdateLevel(DOWN)
	p.PrintLevel("}")
}

func (p *CPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "//package", name)
	p.PrintLevel(NL, "#include <go.h>")
}

func (p *CPrinter) PrintImport(name, path string) {
	p.PrintLevel(NL, "//import", name, path)

	switch path {
	case `"fmt"`:
		p.PrintLevel(NL, "#include <fmt.h>")

	case `"sync"`:
		p.PrintLevel(NL, "#include <sync.h>")

	case `"errors"`:
		p.PrintLevel(NL, "#include <errors.h>")

	case `"time"`:
		p.PrintLevel(NL, "#include <go_time.h>")
	}
}

func (p *CPrinter) PrintType(name, typedef string) {
	if strings.Contains(typedef, "%") {
		// FuncType
		p.PrintLevel(SEMI, "typedef", fmt.Sprintf(typedef, "("+name+")"))
	} else {
		p.PrintLevel(SEMI, "typedef", typedef, name)
	}
}

func (p *CPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	if vtype == "var" {
		vtype = ""
	} else if vtype == "const" && len(values) == 0 {
		values = p.FormatIdent(IOTA)
	}

	if len(typedef) == 0 {
		typedef, values = GuessType(values)
	} else if strings.Contains(typedef, "[") {
		// array or map ?
		i := strings.Index(typedef, "[")
		names += typedef[i:]
		typedef = typedef[0:i]
	}

	if ntuple && len(values) > 0 {
		names = fmt.Sprintf("tie(%s)", names)
	}

	p.PrintLevel(NONE, vtype, typedef, names)

	if len(values) > 0 {
		if vtuple {
			values = fmt.Sprintf("make_tuple(%s)", values)
		}

		p.Print(" =", values)
	}
	p.Print(";\n")
}

func (p *CPrinter) PrintStmt(stmt, expr string) {
	if stmt == "go" {
		// start a goroutine (or a thread)
		p.PrintLevel(SEMI, fmt.Sprintf("Goroutine([](){ %s; })", expr))
	} else if len(stmt) > 0 {
		p.PrintLevel(SEMI, stmt, expr)
	} else {
		p.PrintLevel(SEMI, expr)
	}
}

func (p *CPrinter) PrintReturn(expr string, tuple bool) {
	if tuple {
		expr = fmt.Sprintf("make_tuple(%s)", expr)
	}

	p.PrintStmt("return", expr)
}

func (p *CPrinter) PrintFunc(receiver, name, params, results string) {
	if len(receiver) == 0 && len(params) == 0 && len(results) == 0 && name == "main" {
		// the "main"
		results = "int"
		params = "int argc, char **argv"
	} else {
		if len(results) == 0 {
			results = "void"
		} else if IsMultiValue(results) {
			results = fmt.Sprintf("tuple<%s>", results)
		}

		if len(receiver) > 0 {
			parts := strings.SplitN(receiver, " ", 2)
			receiver = "/* " + parts[1] + " */ " + strings.TrimRight(parts[0], "*") + "::"
		}
	}

	fmt.Fprintf(p.w, "%s %s%s(%s) ", results, receiver, name, params)
}

func (p *CPrinter) PrintFor(init, cond, post string) {
	init = strings.TrimRight(init, SEMI)
	post = strings.TrimRight(post, SEMI)

	onlycond := len(init) == 0 && len(post) == 0

	if len(cond) == 0 {
		cond = "true"
	}

	if onlycond {
		// make it a while
		p.PrintLevel(NONE, "while (", cond)
	} else {
		p.PrintLevel(NONE, "for (")
		if len(init) > 0 {
			p.Print(init)
		}
		p.Print("; " + cond + ";")
		if len(post) > 0 {
			p.Print(" " + post)
		}

	}
	p.Print(") ")
}

func (p *CPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *CPrinter) PrintSwitch(init, expr string) {
	if len(init) > 0 {
		p.PrintLevel(SEMI, init)
	}
	p.PrintLevel(NONE, "switch (", expr, ")")
}

func (p *CPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(NL, "case", expr+":")
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *CPrinter) PrintEndCase() {
	p.PrintLevel(SEMI, "break") // XXX: need to check for previous fallthrough
}

func (p *CPrinter) PrintIf(init, cond string) {
	if len(init) > 0 {
		p.PrintLevel(NONE, init+" if ")
	} else {
		p.PrintLevel(NONE, "if ")
	}
	p.Print("(", cond, ") ")
}

func (p *CPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *CPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *CPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	if op == ":=" {
		// := means there are new variables to be declared (but of course I don't know the real type)
		rtype, rvalue := GuessType(rhs)
		lhs = rtype + " " + lhs
		rhs = rvalue
		op = "="
	}

	if ltuple {
		lhs = fmt.Sprintf("tie(%s)", lhs)
	}

	if rtuple {
		rhs = fmt.Sprintf("make_tuple(%s)", rhs)
	}

	p.PrintLevel(SEMI, lhs, op, rhs)
}

func (p *CPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, fmt.Sprintf("%s.Send(%s)", ch, value))
}

func (p *CPrinter) FormatIdent(id string) (ret string) {
	switch id {
	case NIL:
		return NULL

	case IOTA:
		ret = strconv.Itoa(p.iota)
		p.iota += 1

	case "string":
		ret = "std::string"
	default:
		ret = id
	}

	return
}

func (p *CPrinter) FormatLiteral(lit string) string {
	if len(lit) == 0 {
		return lit
	}

	if lit[0] == '`' {
		lit = strings.Replace(lit[1:len(lit)-1], `"`, `\\"`, -1)
		lit = strings.Replace(lit, "\n", "\\n", -1)
		lit = `"` + lit + `"`
	}

	return lit
}

func (p *CPrinter) FormatCompositeLit(typedef, elt string) string {
	return fmt.Sprintf("%s{%s}", typedef, elt)
}

func (p *CPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *CPrinter) FormatStar(expr string) string {
	return "*" + expr
}

func (p *CPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *CPrinter) FormatUnary(op, operand string) string {
	if op == "<-" {
		return fmt.Sprintf("%s.Receive()", operand)
	}

	return fmt.Sprintf("%s%s", op, operand)
}

func (p *CPrinter) FormatBinary(lhs, op, rhs string) string {
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *CPrinter) FormatPair(v Pair, t FieldType) (ret string) {
	name, value := v.Name(), v.Value()

	if strings.HasPrefix(value, "[") {
		i := strings.LastIndex(value, "]")
		if i < 0 {
			// it should be an error

		} else {
			arr := value[:i+1]
			value = value[i+1:]

			if len(name) > 0 {
				name += arr
			} else {
				value += arr
			}
		}
	}

	if strings.HasPrefix(value, "*") {
		i := strings.LastIndex(value, "*") + 1
		value = value[i:] + value[0:i]
	}

	if t == METHOD {
		if len(name) == 0 {
			ret = fmt.Sprintf("// extends %s", value)
		} else {
			ret = "virtual " + fmt.Sprintf(value, name)
		}
	} else if t == RESULT && len(name) > 0 {
		ret = fmt.Sprintf("%s /* %s */", value, name)
	} else if len(name) > 0 && len(value) > 0 {
		ret = value + " " + name
	} else {
		ret = value + name
	}

	if t == METHOD || t == FIELD {
		ret = p.indent() + ret + SEMI
	} else {
		ret += COMMA
	}

	return
}

func (p *CPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("%s[%s]", elt, len)
}

func (p *CPrinter) FormatArrayIndex(array, index string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *CPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *CPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("map<%s, %s>", key, elt)
}

func (p *CPrinter) FormatKeyValue(key, value string) string {
	return fmt.Sprintf("{%s, %s}", key, value)
}

func (p *CPrinter) FormatStruct(fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct {\n%s}", fields)
	} else {
		return "struct{}"
	}
}

func (p *CPrinter) FormatInterface(methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("struct {\n%s}", methods)
	} else {
		return "struct{}"
	}
}

func (p *CPrinter) FormatChan(chdir, mtype string) string {
	var chtype string

	switch chdir {
	case CHAN_BIDI:
		chtype = "Chan"
	case CHAN_SEND:
		chtype = "SendChan"
	case CHAN_RECV:
		chtype = "ReceiveChan"
	}

	return fmt.Sprintf("%s<%s>", chtype, mtype)
}

func (p *CPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	if strings.HasPrefix(fun, "time::") {
		// need to rename :(
		fun = "go_" + fun
	}

	if isFuncLit {
		return fmt.Sprintf("[%s]%s", args, fun)
	} else {
		return fmt.Sprintf("%s(%s)", fun, args)
	}
}

func (p *CPrinter) FormatFuncType(params, results string, withFunc bool) string {
	if len(results) == 0 {
		results = "void"
	} else if IsMultiValue(results) {
		results = fmt.Sprintf("tuple<%s>", results)
	}

	// add %%s only if withFunc ?
	return fmt.Sprintf("%s %%s(%s)", results, params)
}

func (p *CPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf(ftype+"%s", "", body)
}

func (p *CPrinter) FormatSelector(pname, sel string, isObject bool) string {
	if isObject {
		return fmt.Sprintf("%s.%s", pname, sel)
	} else {
		return fmt.Sprintf("%s::%s", pname, sel)
	}
}

func (p *CPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}

//
// Guess type and return type and new value
//
func GuessType(value string) (string, string) {
	vtype := "void"

	if len(value) == 0 {
		return vtype, value
	}

	switch value[0] {
	case '[':
		// array or map declaration
		i := strings.Index(value, "{")
		if i >= 0 {
			vtype = value[:i]
			value = value[i:]
		}
	case '\'':
		vtype = "char"
	case '"':
		vtype = "string"

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		vtype = "int"

	default:
		switch value {
		case "true", "false":
			vtype = "bool"

		case NIL, NULL:
			vtype = "void*"
		}
	}

	return vtype, value
}

func IsPublic(name string) bool {
	return name[0] >= 'A' && name[0] <= 'Z'
}

func IsMultiValue(expr string) bool {
	return strings.Contains(expr, ",")
}
