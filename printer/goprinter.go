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

func (p *GoPrinter) PrintLevel(values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "))
}

func (p *GoPrinter) PrintPackage(name string) {
	p.PrintLevel("package", name, "\n")
}

func (p *GoPrinter) PrintImport(name, path string) {
	p.PrintLevel("import", name, path, "\n")
}

func (p *GoPrinter) PrintType(name, typedef string) {
	p.PrintLevel("type", name, typedef, "\n")
}

func (p *GoPrinter) PrintValue(vtype, names, typedef, value string) {
	p.PrintLevel(vtype, names)
	if len(typedef) > 0 {
		p.Print(" ", typedef)
	}
	if len(value) > 0 {
		p.Print(" =", value)
	}
	p.Print("\n")
}

func (p *GoPrinter) PrintStmt(stmt, expr string) {
	p.PrintLevel(stmt, expr, "\n")
}

func (p *GoPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel("func ")
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
	p.PrintLevel("for ")
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
	p.PrintLevel("for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *GoPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel("switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *GoPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel("case", expr+":\n")
	} else {
		p.PrintLevel("default:\n")
	}
}

func (p *GoPrinter) PrintIf(init, cond string) {
	p.PrintLevel("if ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond, "")
}

func (p *GoPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *GoPrinter) PrintEmpty() {
	p.PrintLevel(";\n")
}

func (p *GoPrinter) PrintAssignment(lhs, op, rhs string) {
	p.PrintLevel(lhs, op, rhs, "\n")
}

func (p *GoPrinter) FormatIdent(id string) string {
	return id
}

func (p *GoPrinter) FormatLiteral(lit string) string {
	return lit
}

func (p *GoPrinter) FormatPair(v Pair) string {
	return v.String()
}

func (p *GoPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("[%s]%s", len, elt)
}

func (p *GoPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("[%s]%s", key, elt)
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

func (p *GoPrinter) FormatCall(fun, args string) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}
