package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// RustPrinter implement the Printer interface for Rust programs
//
type RustPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
}

func (p *RustPrinter) Reset() {
	p.level = 0
	p.sameline = false
}

func (p *RustPrinter) PushContext() {
}

func (p *RustPrinter) PopContext() {
}

func (p *RustPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *RustPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *RustPrinter) SameLine() {
	p.sameline = true
}

func (p *RustPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *RustPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMANL)
}

func (p *RustPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *RustPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *RustPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *RustPrinter) PrintfLevel(term string, format string, values ...interface{}) {
	fmt.Fprintf(p.w, p.indent()+format+term, values...)
}

func (p *RustPrinter) PrintBlockStart(b BlockType) {
	var open string

	switch b {
	case CONST, VAR:
		open = "("
	default:
		open = "{"
	}

	p.PrintLevel(NL, open)
	p.UpdateLevel(UP)
}

func (p *RustPrinter) PrintBlockEnd(b BlockType) {
	var close string

	switch b {
	case CONST, VAR:
		close = ")"
	default:
		close = "}"
	}

	p.UpdateLevel(DOWN)
	p.PrintLevel(NONE, close)
}

func (p *RustPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "package", name)
}

func (p *RustPrinter) PrintImport(name, path string) {
	p.PrintLevel(NL, "import", name, path)
}

func (p *RustPrinter) PrintType(name, typedef string) {
	if strings.Contains(typedef, "%") {
		p.PrintfLevel(NL, typedef, name)
	} else {
		p.PrintLevel(NL, "type", name, typedef)
	}
}

func (p *RustPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	if vtype == "var" {
		if len(values) > 0 {
			vtype = "let"
		} else {
			vtype = ""
		}
	} else if vtype == "const" {
		vtype = "static"
	}
	p.PrintLevel(NONE, vtype, names)
	if len(typedef) > 0 {
		p.Print(" ", typedef)
	}
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print("\n")
}

func (p *RustPrinter) PrintStmt(stmt, expr string) {
	if len(stmt) > 0 {
		p.PrintLevel(SEMI, stmt, expr)
	} else {
		p.PrintLevel(SEMI, expr)
	}
}

func (p *RustPrinter) PrintReturn(expr string, tuple bool) {
	p.PrintStmt("return", expr)
}

func (p *RustPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel(NONE, "fn ")
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "(%s) ", receiver)
	}
	fmt.Fprintf(p.w, "%s(%s) ", name, params)
	if len(results) > 0 {
		p.Print("->", "")

		if strings.ContainsAny(results, " ,") {
			// name type or multiple types
			fmt.Fprintf(p.w, "(%s) ", results)
		} else {
			p.Print(results, "")
		}
	}
}

func (p *RustPrinter) PrintFor(init, cond, post string) {
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

func (p *RustPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *RustPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel(NONE, "switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *RustPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *RustPrinter) PrintEndCase() {
	// nothing to do
}

func (p *RustPrinter) PrintIf(init, cond string) {
	p.PrintLevel(NONE, "if ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond, "")
}

func (p *RustPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *RustPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *RustPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	if op == ":=" {
		lhs = "let " + lhs
		op = "="
	}

	p.PrintLevel(NL, lhs, op, rhs)
}

func (p *RustPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, ch, "<-", value)
}

func (p *RustPrinter) FormatIdent(id string) string {
	return id
}

func (p *RustPrinter) FormatLiteral(lit string) string {
	return lit
}

func (p *RustPrinter) FormatCompositeLit(typedef, elt string) string {
	return fmt.Sprintf("%s{%s}", typedef, elt)
}

func (p *RustPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *RustPrinter) FormatStar(expr string) string {
	return fmt.Sprintf("*%s", expr)
}

func (p *RustPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *RustPrinter) FormatUnary(op, operand string) string {
	return fmt.Sprintf("%s%s", op, operand)
}

func (p *RustPrinter) FormatBinary(lhs, op, rhs string) string {
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *RustPrinter) FormatPair(v Pair, t FieldType) string {
	switch t {
	case METHOD:
		return p.indent() + v.Name() + v.Value() + NL
	case FIELD:
		return fmt.Sprintf("%s%s: %s%s", p.indent(), v.Name(), v.Value(), COMMANL)
	case PARAM:
		return fmt.Sprintf("%s: %s%s", v.Name(), v.Value(), COMMA)
	default:
		return v.String() + COMMA
	}
}

func (p *RustPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("[%s]%s", len, elt)
}

func (p *RustPrinter) FormatArrayIndex(array, index string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *RustPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *RustPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("[%s]%s", key, elt)
}

func (p *RustPrinter) FormatKeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *RustPrinter) FormatStruct(fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct %%s {\n%s\n}", p.Chop(fields))
	} else {
		return "struct %%s;"
	}
}

func (p *RustPrinter) FormatInterface(methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("interface{\n%s}", methods)
	} else {
		return "interface{}"
	}
}

func (p *RustPrinter) FormatChan(chdir, mtype string) string {
	return fmt.Sprintf("%s %s", chdir, mtype)
}

func (p *RustPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}

func (p *RustPrinter) FormatFuncType(params, results string, withFunc bool) string {
	prefix := ""
	if withFunc {
		prefix = "fn"
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

func (p *RustPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("fn%s %s", ftype, body)
}

func (p *RustPrinter) FormatSelector(pname, sel string, isObject bool) string {
	return fmt.Sprintf("%s.%s", pname, sel)
}

func (p *RustPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}
