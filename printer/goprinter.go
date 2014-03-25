package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// GoPrinter implement the Printer interface for Go programs
//
type GoPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
}

func (p *GoPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *GoPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *GoPrinter) SameLine() {
	p.sameline = true
}

func (p *GoPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *GoPrinter) GetSeparator(ftype FieldType) string {
	if ftype == METHOD || ftype == FIELD {
		return ";\n" + p.indent()
	} else {
		return ", "
	}
}

func (p *GoPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *GoPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *GoPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *GoPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "package", name)
}

func (p *GoPrinter) PrintImport(name, path string) {
	p.PrintLevel(NL, "import", name, path)
}

func (p *GoPrinter) PrintType(name, typedef string) {
	p.PrintLevel(NL, "type", name, typedef)
}

func (p *GoPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	p.PrintLevel(NONE, vtype, names)
	if len(typedef) > 0 {
		p.Print(" ", typedef)
	}
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print("\n")
}

func (p *GoPrinter) PrintStmt(stmt, expr string) {
	if len(stmt) > 0 {
		p.PrintLevel(NL, stmt, expr)
	} else {
		p.PrintLevel(NL, expr)
	}
}

func (p *GoPrinter) PrintReturn(expr string, tuple bool) {
	p.PrintStmt("return", expr)
}

func (p *GoPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel(NONE, "func ")
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "(%s) ", receiver)
	}
	fmt.Fprintf(p.w, "%s(%s) ", name, params)
	if len(results) > 0 {
		if strings.ContainsAny(results, " ,") {
			// name type or multiple types
			fmt.Fprintf(p.w, "(%s) ", results)
		} else {
			p.Print(results, "")
		}
	}
}

func (p *GoPrinter) PrintFor(init, cond, post string) {
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

func (p *GoPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *GoPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel(NONE, "switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *GoPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(SEMI, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *GoPrinter) PrintEndCase() {
	// nothing to do
}

func (p *GoPrinter) PrintIf(init, cond string) {
	p.PrintLevel(NONE, "if ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond, "")
}

func (p *GoPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *GoPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *GoPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	p.PrintLevel(NL, lhs, op, rhs)
}

func (p *GoPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, ch, "<-", value)
}

func (p *GoPrinter) FormatIdent(id string) string {
	return id
}

func (p *GoPrinter) FormatLiteral(lit string) string {
	return lit
}

func (p *GoPrinter) FormatUnary(op, operand string) string {
	return fmt.Sprintf("%s%s", op, operand)
}

func (p *GoPrinter) FormatBinary(lhs, op, rhs string) string {
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *GoPrinter) FormatPair(v Pair, t FieldType) string {
	if t == METHOD {
		return v.Name() + v.Value()
	} else {
		return v.String()
	}
}

func (p *GoPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("[%s]%s", len, elt)
}

func (p *GoPrinter) FormatArrayIndex(array, index string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *GoPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *GoPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("[%s]%s", key, elt)
}

func (p *GoPrinter) FormatKeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *GoPrinter) FormatStruct(fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct{\n%s%s\n%s}", p.indent(), fields, p.indent())
	} else {
		return fmt.Sprintf("struct{}")
	}
}

func (p *GoPrinter) FormatInterface(methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("interface{\n%s%s\n%s}", p.indent(), methods, p.indent())
	} else {
		return fmt.Sprintf("interface{}")
	}
}

func (p *GoPrinter) FormatChan(chdir, mtype string) string {
	return fmt.Sprintf("%s %s", chdir, mtype)
}

func (p *GoPrinter) FormatCall(fun, args string) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}

func (p *GoPrinter) FormatFuncType(params, results string) string {
	if len(results) == 0 {
		// no results
		return fmt.Sprintf("(%s)", params)
	}

	if strings.ContainsAny(results, " ,") {
		// name type or multiple types
		return fmt.Sprintf("(%s) (%s)", params, results)
	}

	// just type
	return fmt.Sprintf("(%s) %s", params, results)
}

func (p *GoPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("func%s %s", ftype, body)
}
