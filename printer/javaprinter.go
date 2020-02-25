package printer

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

//
// JavaPrinter implement the Printer interface for Java programs
//
type JavaPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
}

func mod(name string) string {
	first, _ := utf8.DecodeRuneInString(name)
	if unicode.IsUpper(first) {
		return "public"
	}

	return "private"
}

func javatype(t string) string {
	switch t {
	case "string":
		return "String"

	case "int8":
		return "byte"

	case "int16":
		return "short"

	case "int32", "rune":
		return "int"

	case "int64":
		return "long"

	case "uint8", "byte":
		return "/* unsigned */ byte"

	case "uint16":
		return "/* unsigned */ short"

	case "uint32":
		return "/* unsigned */ int"

	case "uint64":
		return "/* unsigned */ long"

	case "float32":
		return "float"

	case "float64":
		return "double"

	case "bool":
		return "boolean"

	case "struct{}":
		return "Object"

	case "interface{}":
		return "/* interface */ Object"
	}

	return t
}

func (p *JavaPrinter) Reset() {
	p.level = 0
	p.sameline = false
}

func (p *JavaPrinter) PushContext() {
}

func (p *JavaPrinter) PopContext() {
}

func (p *JavaPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *JavaPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *JavaPrinter) SameLine() {
	p.sameline = true
}

func (p *JavaPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *JavaPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMA)
}

func (p *JavaPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *JavaPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *JavaPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *JavaPrinter) PrintBlockStart(b BlockType, empty bool) {
	var open string

	switch b {
	case CONST, VAR:
		return
	default:
		open = "{"
	}

	p.PrintLevel(NL, open)
	p.UpdateLevel(UP)
}

func (p *JavaPrinter) PrintBlockEnd(b BlockType) {
	var close string

	switch b {
	case CONST, VAR:
		return
	default:
		close = "}"
	}

	p.UpdateLevel(DOWN)
	p.PrintLevel(NONE, close)
}

func (p *JavaPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "package", name)
}

func (p *JavaPrinter) PrintImport(name, path string) {
	p.PrintLevel(NL, "import", name, path)
}

func (p *JavaPrinter) PrintType(name, typedef string) {
	//p.PrintLevel(NL, "type", name, typedef)

	cdef := mod(name)

	if strings.HasPrefix(typedef, "struct{") {
		typedef = typedef[6:]
		cdef += " class"
	} else if strings.HasPrefix(typedef, "interface{") {
		typedef = typedef[9:]
		cdef += " interface"
	}

	p.PrintLevel(NL, cdef, name, javatype(typedef))
}

func (p *JavaPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	def := ""
	if vtype == "const" {
		def = "final "
	}

	def += mod(names)

	p.PrintLevel(NONE, def, javatype(typedef), names)
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print(";\n")
}

func (p *JavaPrinter) PrintStmt(stmt, expr string) {
	if len(stmt) > 0 {
		p.PrintLevel(SEMI, stmt, expr)
	} else {
		p.PrintLevel(SEMI, expr)
	}
}

func (p *JavaPrinter) PrintReturn(expr string, tuple bool) {
	p.PrintStmt("return", expr)
}

func (p *JavaPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel(NONE, mod(name), "")
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "/* %s */ ", receiver)
	}
	if len(results) > 0 {
		if strings.ContainsAny(results, " ,") {
			// name type or multiple types
			fmt.Fprintf(p.w, "(%s) ", results)
		} else {
			p.Print(javatype(results), "")
		}
	} else {
		p.Print("void ")
	}
	fmt.Fprintf(p.w, "%s(%s) ", name, params)
}

func (p *JavaPrinter) PrintFor(init, cond, post string) {
	p.PrintLevel(NONE, "for ")
	if len(init) > 0 {
		p.Print(init)
	}

	p.Print("; ")
	p.Print(cond)
	p.Print("; ")

	if len(post) > 0 {
		p.Print(post)
	}

	p.Print("")
}

func (p *JavaPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" := range", expr)

}

func (p *JavaPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel(NONE, "switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *JavaPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *JavaPrinter) PrintEndCase() {
	// nothing to do
}

func (p *JavaPrinter) PrintIf(init, cond string) {
	p.PrintLevel(NONE, "if (")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond+")", "")
}

func (p *JavaPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *JavaPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *JavaPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	if op == ":=" {
		p.PrintLevel(SEMI, "var", lhs, "=", rhs)
	} else {
		p.PrintLevel(SEMI, lhs, op, rhs)
	}
}

func (p *JavaPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, ch, "<-", value)
}

func (p *JavaPrinter) FormatIdent(id string) string {
	if id == "nil" {
		return "null"
	}

	return id
}

func (p *JavaPrinter) FormatLiteral(lit string) string {
	return lit
}

func (p *JavaPrinter) FormatCompositeLit(typedef, elt string) string {
	return fmt.Sprintf("%s{%s}", typedef, elt)
}

func (p *JavaPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *JavaPrinter) FormatStar(expr string) string {
	return fmt.Sprintf("*%s", expr)
}

func (p *JavaPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *JavaPrinter) FormatUnary(op, operand string) string {
	return fmt.Sprintf("%s%s", op, operand)
}

func (p *JavaPrinter) FormatBinary(lhs, op, rhs string) string {
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *JavaPrinter) FormatPair(v Pair, t FieldType) string {
	switch t {
	case METHOD:
		parts := strings.SplitN(v.Value(), "(", 2)
		mdef := parts[0] + v.Name() + "(" + parts[1]
		return p.indent() + mdef + SEMI
	case PARAM, RECEIVER:
		return javatype(v.Value()) + " " + v.Name() + COMMA
	case FIELD:
		return p.indent() + javatype(v.Value()) + " " + v.Name() + SEMI
	default:
		return v.String() + COMMA
	}
}

func (p *JavaPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("[%s]%s", len, elt)
}

func (p *JavaPrinter) FormatArrayIndex(array, index string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *JavaPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *JavaPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("[%s]%s", key, elt)
}

func (p *JavaPrinter) FormatKeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *JavaPrinter) FormatStruct(fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct{\n%s}", fields)
	} else {
		return "struct{}"
	}
}

func (p *JavaPrinter) FormatInterface(methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("interface{\n%s}", methods)
	} else {
		return "interface{}"
	}
}

func (p *JavaPrinter) FormatChan(chdir, mtype string) string {
	return fmt.Sprintf("%s %s", chdir, mtype)
}

func (p *JavaPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}

func (p *JavaPrinter) FormatFuncType(params, results string, withFunc bool) string {
	prefix := ""
	if withFunc {
		prefix = "/* function */"
	}

	if len(results) == 0 {
		// no results
		return fmt.Sprintf("void %s(%s)", prefix, params)
	}

	if strings.ContainsAny(results, ", ") {
		// name type or multiple types
		return fmt.Sprintf("/* (%s) */ %s(%s)", results, prefix, params)
	}

	// just type
	return fmt.Sprintf("%s %s(%s)", javatype(results), prefix, params)
}

func (p *JavaPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("func%s %s", ftype, body)
}

func (p *JavaPrinter) FormatSelector(pname, sel string, isObject bool) string {
	return fmt.Sprintf("%s.%s", pname, sel)
}

func (p *JavaPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}
