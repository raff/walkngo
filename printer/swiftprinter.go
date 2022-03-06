package printer

import (
	"fmt"
	"io"
	"strings"
)

//
// SwiftPrinter implement the Printer interface for Swift programs
//
type SwiftPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer
}

func (p *SwiftPrinter) Reset() {
	p.level = 0
	p.sameline = false
}

func (p *SwiftPrinter) PushContext(c ContextType) {
}

func (p *SwiftPrinter) PopContext() {
}

func (p *SwiftPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *SwiftPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *SwiftPrinter) SameLine() {
	p.sameline = true
}

func (p *SwiftPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *SwiftPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMA)
}

func (p *SwiftPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *SwiftPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *SwiftPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *SwiftPrinter) PrintBlockStart(b BlockType, empty bool) {
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

func (p *SwiftPrinter) PrintBlockEnd(b BlockType) {
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

func (p *SwiftPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "package", name)
}

func (p *SwiftPrinter) PrintImport(name, path string) {
	p.PrintLevel(NL, "import", name, path)
}

func (p *SwiftPrinter) PrintType(name, typedef string) {
	p.PrintLevel(NL, "type", name, typedef)
}

func (p *SwiftPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	switch vtype {
	case "const":
		vtype = "let"

	case "":
		vtype = "var"
	}

	p.PrintLevel(NONE, vtype, names)
	if len(typedef) > 0 {
		p.Print(": ", typedef)
	}
	if len(values) > 0 {
		p.Print(" =", values)
	}
	p.Print("\n")
}

func (p *SwiftPrinter) PrintStmt(stmt, expr string) {
	if len(stmt) > 0 {
		p.PrintLevel(NL, stmt, expr)
	} else {
		p.PrintLevel(NL, expr)
	}
}

func (p *SwiftPrinter) PrintReturn(expr string, tuple bool) {
	p.PrintStmt("return", expr)
}

func (p *SwiftPrinter) PrintFunc(receiver, name, params, results string) {
	p.PrintLevel(NONE, "func ")
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "(%s) ", receiver)
	}
	fmt.Fprintf(p.w, "%s(%s) ", name, params)
	if len(results) > 0 {
		if strings.ContainsAny(results, " ,") {
			// name type or multiple types
			fmt.Fprintf(p.w, "-> (%s) ", results)
		} else {
			fmt.Fprintf(p.w, "-> %s ", results)
		}
	}
}

func (p *SwiftPrinter) PrintFor(init, cond, post string) {
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

func (p *SwiftPrinter) PrintRange(key, value, expr string) {
	p.PrintLevel(NONE, "for", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(" in", expr)

}

func (p *SwiftPrinter) PrintSwitch(init, expr string) {
	p.PrintLevel(NONE, "switch ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(expr)
}

func (p *SwiftPrinter) PrintCase(expr string) {
	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *SwiftPrinter) PrintEndCase() {
	// nothing to do
}

func (p *SwiftPrinter) PrintIf(init, cond string) {
	p.PrintLevel(NONE, "if ")
	if len(init) > 0 {
		p.Print(init + "; ")
	}
	p.Print(cond, "")
}

func (p *SwiftPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *SwiftPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *SwiftPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	p.PrintLevel(NL, lhs, op, rhs)
}

func (p *SwiftPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, ch, "<-", value)
}

func (p *SwiftPrinter) FormatIdent(id string) (ret string) {
	switch id {
	//ase IOTA:
	//ret = strconv.Itoa(p.ctx.iota)
	//p.ctx.iota += 1

	case "string":
		ret = "String"
	case "int":
		ret = "Int"
	case "int8":
		ret = "Int8"
	case "int32":
		ret = "Int32"
	case "int64":
		ret = "Int64"
	case "uint":
		ret = "UInt"
	case "uint8":
		ret = "UInt8"
	case "uint32":
		ret = "UInt32"
	case "uint64":
		ret = "UInt64"
	case "float32":
		ret = "Float"
	case "float64":
		ret = "Double"
	case "bool":
		ret = "Bool"

	default:
		ret = id
	}

	return
}

func (p *SwiftPrinter) FormatLiteral(lit string) string {
	return lit
}

func (p *SwiftPrinter) FormatCompositeLit(typedef, elt string) string {
	if strings.HasPrefix(typedef, "Array<") || strings.HasPrefix(typedef, "Slice<") || strings.HasPrefix(typedef, "Dictionary<") {
		if len(elt) > 0 {
			return fmt.Sprintf("[ %s ]", elt)
		} else {
			return fmt.Sprintf("%s()", typedef)
		}
	} else {
		return fmt.Sprintf("%s{%s}", typedef, elt)
	}
}

func (p *SwiftPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *SwiftPrinter) FormatStar(expr string) string {
	return fmt.Sprintf("*%s", expr)
}

func (p *SwiftPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *SwiftPrinter) FormatUnary(op, operand string) string {
	return fmt.Sprintf("%s%s", op, operand)
}

func (p *SwiftPrinter) FormatBinary(lhs, op, rhs string) string {
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *SwiftPrinter) FormatPair(v Pair, t FieldType) string {
	switch t {
	case METHOD:
		return p.indent() + v.Name() + v.Value() + NL
	case FIELD:
		return p.indent() + v.String() + NL
	case PARAM:
		return v.Name() + ": " + v.Value() + COMMA
	default:
		return v.String() + COMMA
	}
}

func (p *SwiftPrinter) FormatArray(l, elt string) string {
	if len(l) == 0 {
		return fmt.Sprintf("Slice<%s>", elt)
	} else {
		return fmt.Sprintf("Array<%s>", elt)
	}
}

func (p *SwiftPrinter) FormatArrayIndex(array, index, ctype string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *SwiftPrinter) FormatMapIndex(array, index, ctype string, check bool) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *SwiftPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *SwiftPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("Dictionary<%s, %s>", key, elt)
}

func (p *SwiftPrinter) FormatKeyValue(key, value string, isMap bool) string {
	return fmt.Sprintf("%s: %s", key, value)
}

func (p *SwiftPrinter) FormatStruct(name, fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct{\n%s}", fields)
	} else {
		return "struct{}"
	}
}

func (p *SwiftPrinter) FormatInterface(name, methods string) string {
	if len(methods) > 0 {
		return fmt.Sprintf("interface{\n%s}", methods)
	} else {
		return "interface{}"
	}
}

func (p *SwiftPrinter) FormatChan(chdir, mtype string) string {
	return fmt.Sprintf("%s %s", chdir, mtype)
}

func (p *SwiftPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	return fmt.Sprintf("%s(%s)", fun, args)
}

func (p *SwiftPrinter) FormatFuncType(params, results string, withFunc bool) string {
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

func (p *SwiftPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf("func%s %s", ftype, body)
}

func (p *SwiftPrinter) FormatSelector(pname, sel string, isObject bool) string {
	return fmt.Sprintf("%s.%s", pname, sel)
}

func (p *SwiftPrinter) FormatTypeAssert(orig, assert string) string {
	return fmt.Sprintf("%s.(%s)", orig, assert)
}
