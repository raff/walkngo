package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// C implement the Printer interface for C programs
//
type CPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
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

func (p *CPrinter) PrintLevel(values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "))
}

func (p *CPrinter) PrintPackage(name string) {
	p.PrintLevel("//package", name, "\n")
}

func (p *CPrinter) PrintImport(name, path string) {
	p.PrintLevel("#include", name, path, "\n")
}

func (p *CPrinter) PrintType(name, typedef string) {
	p.PrintLevel("typedef", typedef, name, ";\n")
}

func (p *CPrinter) PrintValue(vtype, names, typedef, value string) {
	if vtype == "var" {
		vtype = ""
	}

	if len(typedef) == 0 {
		typedef = "void"
	}

	p.PrintLevel(vtype, typedef, names)

	if len(value) > 0 {
		p.Print(" =", value)
	}
	p.Print(";\n")
}

func (p *CPrinter) PrintStmt(stmt, expr string) {
	p.PrintLevel(stmt, expr, ";\n")
}

func (p *CPrinter) PrintFunc(receiver, name, params, results string) {
	if len(results) == 0 {
		results = "void"
	}

	if len(receiver) > 0 {
		receiver += ", "
	}
	fmt.Fprintf(p.w, "%s %s(%s%s) ", results, name, receiver, params)
}

func (p *CPrinter) PrintFor(init, cond, post string) {
	onlycond := len(init) == 0 && len(post) == 0

	if len(cond) == 0 {
		cond = "true"
	}

	if onlycond {
		// make it a while
		p.PrintLevel("while (", cond)
	} else {
		p.PrintLevel("for (")
		if len(init) > 0 {
			p.Print(init)
		}
		p.Print("; ", cond, "; ")
		if len(post) > 0 {
			p.Print(post)
		}

	}
	p.Print(")" + " ")
}

func (p *CPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel("switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *CPrinter) PrintIf(init, cond string) {
	p.PrintLevel("if ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond, " ")
}

func (p *CPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *CPrinter) PrintEmpty() {
	p.PrintLevel(";\n")
}
