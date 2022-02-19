package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// JavaPrinter implement the Printer interface for Java programs
//
type JavaPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer

	ctx *Jcontext
}

//
// Jcontext is the context for a (function) block
//
type Jcontext struct {
	iota int // incremented when 'const n = iota' or 'const n'

	deferred int // used to generate unique names for "defer" callbacks

	receiver        string // the name of the receiver, to be converted to "this"
	ret_definitions string // used to define return variables
	ret_values      string // used to "fill" empty returns

	fall_through bool // fall through next case in switch
	case_break   bool // got case break

	ctype ContextType
	next  *Jcontext
}

func (ctx *Jcontext) Selector(s string) string {
	if ctx != nil && ctx.receiver == s {
		return "this"
	}

	return s
}

func (ctx *Jcontext) findContextType(c ContextType) bool {
	for ; ctx != nil; ctx = ctx.next {
		if ctx.ctype == c {
			return true
		}
	}

	return false
}

func (ctx *Jcontext) mod(name string, funcdef bool) string {
	// fmt.Println("MOD", ctx.ctype, name)

	if (funcdef && ctx.next == nil) || !ctx.findContextType(FUNCONTEXT) {
		if IsPublic(name) {
			return "public "
		}

		return "private "
	}

	return ""
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
	p.ctx = nil
}

func (p *JavaPrinter) PushContext(c ContextType) {
	p.ctx = &Jcontext{ctype: c, next: p.ctx}
}

func (p *JavaPrinter) PopContext() {
	p.ctx = p.ctx.next
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

	cdef := p.ctx.mod(name, false)

	if strings.HasPrefix(typedef, "struct{") {
		typedef = typedef[6:]
		cdef += "class"
	} else if strings.HasPrefix(typedef, "interface{") {
		typedef = typedef[9:]
		cdef += "interface"
	}

	p.PrintLevel(NL, cdef, name, javatype(typedef))
}

func (p *JavaPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	def := ""
	if vtype == "const" {
		def = "final "
	}

	def += p.ctx.mod(names, false)

	if typedef == "" {
		typedef = "var"
	} else {
		typedef = javatype(typedef)
	}

	p.PrintLevel(NONE, def, typedef, names)
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print(";\n")
}

func (p *JavaPrinter) PrintStmt(stmt, expr string) {
	if stmt == "fallthrough" {
		p.ctx.fall_through = true
		return
	}

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
	if len(receiver) == 0 && len(params) == 0 && len(results) == 0 && name == "main" {
		// the "main"
		p.Print("public static void main(String[] args) ")
		return
	}

	p.PrintLevel(NONE, p.ctx.mod(name, true))
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "/* %s */ ", receiver)
		parts := strings.SplitN(receiver, " ", 2)
		p.ctx.receiver = parts[1]
	}

	if len(results) == 0 {
		p.Print("void ")
	} else if IsMultiValue(results) {
		// name type or multiple types
		fmt.Fprintf(p.w, "Tuple<%s> ", results)
	} else {
		p.Print(javatype(results), "")
	}

	fmt.Fprintf(p.w, "%s(%s) ", name, params)
}

func (p *JavaPrinter) PrintFor(init, cond, post string) {
	p.PrintLevel(NONE, "for (")
	if len(init) > 0 {
		p.Print(init)
	}

	p.Print("; ")
	p.Print(cond)
	p.Print("; ")

	if len(post) > 0 {
		p.Print(post)
	}

	p.Print(")")
}

func (p *JavaPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for (", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" : ", expr, ")")

}

func (p *JavaPrinter) PrintSwitch(init, expr string) {
	if len(init) > 0 {
		p.PrintLevel(SEMI, init)
	}
	p.PrintLevel(NONE, "switch (", expr, ")")
}

func (p *JavaPrinter) PrintCase(expr string) {
	p.ctx.fall_through = false

	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *JavaPrinter) PrintEndCase() {
	if !p.ctx.fall_through {
		p.PrintLevel(NL, "break")
	}
}

func (p *JavaPrinter) PrintIf(init, cond string) {
	if len(init) > 0 {
		p.PrintLevel(NONE, init+" if ")
	} else {
		p.PrintLevel(NONE, "if ")
	}
	p.Print("(", cond, ") ")
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
	if strings.HasPrefix(lit, "`") {
		return `"` + strings.Trim(lit, "`") + `"`
	}

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
		typedef := v.Value()
		tag := ""
		parts := strings.SplitN(typedef, " `", 2)
		if len(parts) == 2 {
			typedef = parts[0]
			tag = fmt.Sprintf(" @Tag(%q)", strings.TrimRight(parts[1], "`"))
		}

		return p.indent() + javatype(typedef) + " " + v.Name() + tag + SEMI
	default:
		return v.String() + COMMA
	}
}

func (p *JavaPrinter) FormatArray(len, elt string) string {
	return fmt.Sprintf("%s[%s]", elt, len)
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

func (p *JavaPrinter) FormatKeyValue(key, value string, isMap bool) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *JavaPrinter) FormatStruct(name, fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct{\n%s}", fields)
	} else {
		return "struct{}"
	}
}

func (p *JavaPrinter) FormatInterface(name, methods string) string {
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
	//fmt.Println("CALL", fun, args)

	switch fun {
	case "make":
		parts := strings.Split(args, ", ")
		if len(parts) == 2 { // make(type, len)
			if strings.Contains(parts[0], "[]") { // make slice
				atype := strings.Replace(parts[0], "[]", "", 1)
				alen := parts[1]

				return fmt.Sprintf("new %s[%s]", atype, alen)
			}
		}

	case "new":
		return fmt.Sprintf("new %s", args)

	case "len":
		return fmt.Sprintf("%s.length", args)

	case "strings.ToLower":
		return fmt.Sprintf("%v.toLowerCase()", args)

	case "strings.ToUpper":
		return fmt.Sprintf("%v.toUpperCase()", args)

	case "[]byte", "byte[]": // cast (string) to bytes
		return fmt.Sprintf("%v.getBytes()", args)

	case "string": // cast ([]byte) to string
		return fmt.Sprintf("String(%v)", args)

	case "byte", "char", "int", "uint", "bool", "run", "float32", "float64",
		"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64":
		return fmt.Sprintf("(%v)%v", javatype(fun), args)
	}
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

	if IsMultiValue(results) {
		// name type or multiple types
		return fmt.Sprintf("Tuple<%s> %s(%s)", results, prefix, params)
	}

	// just type
	return fmt.Sprintf("%s %s(%s)", javatype(results), prefix, params)
}

func (p *JavaPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("%s -> %s", ftype, body)
}

func (p *JavaPrinter) FormatSelector(pname, sel string, isObject bool) string {
	if isObject {
		return fmt.Sprintf("%s.%s", p.ctx.Selector(pname), sel)
	} else {
		return fmt.Sprintf("%s.%s", pname, sel)
	}
}

func (p *JavaPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}
