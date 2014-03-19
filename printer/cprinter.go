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
	if strings.HasPrefix(typedef, "struct{\n") {
		defs := strings.Split(typedef[8:len(typedef)-1], ";\n")
		p.PrintLevel("class", name, "{\n")
		p.UpdateLevel(1)
		for _, def := range defs {
			p.PrintLevel(def, ";\n")
		}
		p.UpdateLevel(-1)
		p.PrintLevel("}\n")
	} else {
		p.PrintLevel("typedef", typedef, name, ";\n")
	}
}

func (p *CPrinter) PrintValue(vtype, names, typedef, value string) {
	if vtype == "var" {
		vtype = ""
	}

	if len(typedef) == 0 {
		typedef, value = GuessType(value)
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

	if len(receiver) > 0 && len(params) > 0 {
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
	p.Print(") ")
}

func (p *CPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel("for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *CPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel("switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *CPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel("case", expr+":\n")
	} else {
		p.PrintLevel("default:\n")
	}
}

func (p *CPrinter) PrintIf(init, cond string) {
	if len(init) > 0 {
		p.PrintLevel(init + " if ")
	} else {
		p.PrintLevel("if ")
	}
	p.Print(cond, "")
}

func (p *CPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *CPrinter) PrintEmpty() {
	p.PrintLevel(";\n")
}

func (p *CPrinter) PrintAssignment(lhs, op, rhs string) {
	if op == ":=" {
		// := means there are new variables to be declared (but of course I don't know the real type)
		rtype, rvalue := GuessType(rhs)
		lhs = rtype + " " + lhs
		rhs = rvalue
		op = "="
	}

	p.PrintLevel(lhs, op, rhs, ";\n")
}

func (p *CPrinter) FormatIdent(id string) string {
	switch id {
	case "true", "false":
		return strings.ToUpper(id)

	case "nil":
		return "NULL"

	default:
		return id
	}
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

func (p *CPrinter) FormatPair(v Pair) string {
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
		for i, c := range value {
			if c != '*' {
				name = value[:i] + name
				value = value[i:]
				break
			}
		}
	}
	if len(name) > 0 && len(v.Value()) > 0 {
		return value + " " + name
	} else {
		return value + name
	}
}

func (p *CPrinter) FormatCall(fun, args string) string {
	switch fun {
	case "fmt.Printf":
		fun = "printf"
	case "fmt.Sprintf":
		fun = "sprintf"
	case "fmt.Fprintf":
		fun = "fprintf"
	case "fmt.Println":
		fun = "fprintf"
		args = `"%s\n", ` + args
	case "os.Open":
		fun = "open"
	}

	return fmt.Sprintf("%s(%s)", fun, args)
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
		case "true", "false", "TRUE", "FALSE":
			vtype = "bool"

		case "nil", "NULL":
			vtype = "void*"
		}
	}

	return vtype, value
}
