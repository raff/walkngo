package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// PythonPrinter implement the Printer interface for Python programs
//
type PythonPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
}

func (p *PythonPrinter) Reset() {
	p.level = 0
	p.sameline = false
}

func (p *PythonPrinter) PushContext() {
}

func (p *PythonPrinter) PopContext() {
}

func (p *PythonPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *PythonPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *PythonPrinter) SameLine() {
	p.sameline = true
}

func (p *PythonPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *PythonPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMA)
}

func (p *PythonPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *PythonPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *PythonPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *PythonPrinter) PrintBlockStart(b BlockType, empty bool) {
	/*
		var open string

		switch b {
		case CONST, VAR:
			open = "("
		default:
			open = "{"
		}

		p.PrintLevel(NL, open)
	*/

	p.UpdateLevel(UP)
	p.Print("\n")

	if empty {
		p.PrintLevel(NL, "pass")
	}
}

func (p *PythonPrinter) PrintBlockEnd(b BlockType) {
	/*
		var close string

		switch b {
		case CONST, VAR:
			close = ")"
		default:
			close = "}"
		}
	*/

	p.UpdateLevel(DOWN)

	//p.PrintLevel(NONE, close)
}

func (p *PythonPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "# package", name)
}

func (p *PythonPrinter) PrintImport(name, path string) {
	if len(name) > 0 {
		p.PrintLevel(NL, "import", path, "as", name)
	} else {
		p.PrintLevel(NL, "import", path)
	}
}

func (p *PythonPrinter) PrintType(name, typedef string) {
	p.PrintLevel(NL, "type", name, typedef)
}

func (p *PythonPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	p.PrintLevel(NONE, vtype, names)
	if len(typedef) > 0 {
		p.Print(" ", typedef)
	}
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print("\n")
}

func (p *PythonPrinter) PrintStmt(stmt, expr string) {
	if len(stmt) > 0 {
		p.PrintLevel(NL, stmt, expr)
	} else {
		p.PrintLevel(NL, expr)
	}
}

func (p *PythonPrinter) PrintReturn(expr string, tuple bool) {
	p.PrintStmt("return", expr)
}

func (p *PythonPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel(NONE, "def ")

	if len(receiver) > 0 {
		params = "self, " + params
	}

	fmt.Fprintf(p.w, "%s(%s):", name, params)
	if len(receiver) > 0 || len(results) > 0 {
		fmt.Fprintf(p.w, "  # receiver: %v, results: %v\n", receiver, results)
	}
}

func (p *PythonPrinter) PrintFor(init, cond, post string) {
	p.PrintLevel(NONE, "for ")
	if len(init) > 0 {
		p.Print(init)
	}
	if len(init) > 0 || len(post) > 0 {
		p.Print("; ")
	}

	p.Print(cond)

	if len(post) > 0 {
		p.Print(";", post)
	}

	p.Print("")
}

func (p *PythonPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *PythonPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel(NONE, "switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *PythonPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *PythonPrinter) PrintEndCase() {
	// nothing to do
}

func (p *PythonPrinter) PrintIf(init, cond string) {
	if len(init) > 0 {
		p.PrintLevel(NL, init)
	}
	p.PrintLevel(NONE, "if", cond, ":")
}

func (p *PythonPrinter) PrintElse() {
	p.PrintLevel(NONE, "else:")
}

func (p *PythonPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *PythonPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	p.PrintLevel(NL, lhs, op, rhs)
}

func (p *PythonPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, ch, "<-", value)
}

func (p *PythonPrinter) FormatIdent(id string) string {
	switch id {
	case "nil":
		return "None"

	case "true":
		return "True"

	case "false":
		return "False"
	}

	return id
}

func (p *PythonPrinter) FormatLiteral(lit string) string {
	if strings.HasPrefix(lit, "`") {
		return `"""` + strings.Trim(lit, "`") + `"""`
	} else {
		return lit
	}
}

func (p *PythonPrinter) FormatCompositeLit(typedef, elt string) string {
	return fmt.Sprintf("%s{%s}", typedef, elt)
}

func (p *PythonPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *PythonPrinter) FormatStar(expr string) string {
	return fmt.Sprintf("*%s", expr)
}

func (p *PythonPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *PythonPrinter) FormatUnary(op, operand string) string {
	if op == "!" {
		op = "not"
	}

	return fmt.Sprintf("%s%s", op, operand)
}

func (p *PythonPrinter) FormatBinary(lhs, op, rhs string) string {
	switch op {
	case "&&":
		op = "and"

	case "||":
		op = "or"
	}

	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *PythonPrinter) FormatPair(v Pair, t FieldType) string {
	switch t {
	case METHOD:
		return p.indent() + v.Name() + v.Value() + NL
	case FIELD:
		return p.indent() + v.String() + NL
	default:
		return v.String() + COMMA
	}
}

func (p *PythonPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("[%s]%s", len, elt)
}

func (p *PythonPrinter) FormatArrayIndex(array, index string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *PythonPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *PythonPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("[%s]%s", key, elt)
}

func (p *PythonPrinter) FormatKeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *PythonPrinter) FormatStruct(fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct{\n%s}", fields)
	} else {
		return "struct{}"
	}
}

func (p *PythonPrinter) FormatInterface(methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("interface{\n%s}", methods)
	} else {
		return "interface{}"
	}
}

func (p *PythonPrinter) FormatChan(chdir, mtype string) string {
	return fmt.Sprintf("%s %s", chdir, mtype)
}

func (p *PythonPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}

func (p *PythonPrinter) FormatFuncType(params, results string, withFunc bool) string {
	prefix := ""
	if withFunc {
		prefix = "func"
	}

	if len(results) == 0 {
		// no results
		return fmt.Sprintf("%s(%s)", prefix, params)
	}

	if strings.ContainsAny(results, ", ") {
		// name type or multiple types
		return fmt.Sprintf("%s(%s) (%s)", prefix, params, results)
	}

	// just type
	return fmt.Sprintf("%s(%s) %s", prefix, params, results)
}

func (p *PythonPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("func%s %s", ftype, body)
}

func (p *PythonPrinter) FormatSelector(pname, sel string, isObject bool) string {
	return fmt.Sprintf("%s.%s", pname, sel)
}

func (p *PythonPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}
