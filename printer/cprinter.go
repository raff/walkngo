package printer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	NIL  = "nil"
	NULL = "nullptr"
	IOTA = "iota"
)

// C implement the Printer interface for C programs
type CPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer

	ctx *CContext
}

// CContext is the context for a (function) block
type CContext struct {
	context ContextType

	iota_count int // incremented when 'const n = iota' or 'const n'

	deferred int // used to generate unique names for "defer" callbacks

	receiver        string // the name of the receiver, to be converted to "this"
	ret_definitions string // used to define return variables
	ret_values      string // used to "fill" empty returns

	fall_through bool // fall through next case in switch

	caseType string // the object of a switch type assertion

	next *CContext
}

func (ctx *CContext) Selector(s string) string {
	if ctx != nil && ctx.receiver == s {
		return "this->"
	}

	return fmt.Sprintf("%s.", s)
}

func (p *CPrinter) Reset() {
	p.level = 0
	p.sameline = false

	p.ctx = nil
}

func (p *CPrinter) PushContext(c ContextType) {
	if p.ctx == nil {
		p.ctx = &CContext{context: c}
		return
	}

	p.ctx = &CContext{
		context:         c,
		iota_count:      p.ctx.iota_count,
		deferred:        p.ctx.deferred,
		receiver:        p.ctx.receiver,
		ret_definitions: p.ctx.ret_definitions,
		ret_values:      p.ctx.ret_values,
		fall_through:    p.ctx.fall_through,
		next:            p.ctx,
	}
}

func (p *CPrinter) PopContext() {
	p.ctx = p.ctx.next
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

func (p *CPrinter) PrintBlockStart(b BlockType, empty bool) {
	var open string

	switch b {
	case CONST, VAR:
		open = "("
	default:
		open = "{"
	}

	p.PrintLevel(NL, open)
	p.UpdateLevel(UP)

	if b == CODE && len(p.ctx.ret_definitions) > 0 {
		p.PrintLevel(NL, p.ctx.ret_definitions)
		p.ctx.ret_definitions = "" // this gets printed only once
	}
}

func (p *CPrinter) PrintBlockEnd(b BlockType) {
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
		values = p.FormatIdent(IOTA, "")
	}

	if len(typedef) == 0 {
		typedef, values = GuessType(values)
	} else if strings.Contains(typedef, "[") {
		i := strings.Index(typedef, "[")
		names += typedef[i:]
		typedef = typedef[:i]
	}

	if ntuple && len(values) > 0 {
		names = fmt.Sprintf("std::tie(%s)", names)
	}

	p.PrintLevel(NONE, vtype, typedef, names)

	if len(values) > 0 {
		if vtuple {
			values = fmt.Sprintf("std::make_tuple(%s)", values)
		}

		p.Print(" =", values)
	}
	p.Print(";\n")
}

func (p *CPrinter) PrintStmt(stmt, expr string) {
	switch {
	case stmt == "fallthrough":
		p.ctx.fall_through = true
		p.PrintLevel(SEMI, "// fallthrough")

	case stmt == "go":
		// start a goroutine (or a thread)
		p.PrintLevel(SEMI, fmt.Sprintf("Goroutine([](){ %s; })", expr))

	case stmt == "defer":
		p.PrintLevel(SEMI, fmt.Sprintf("Deferred defer%d([](){ %s; })", p.ctx.deferred, expr))
		p.ctx.deferred++

	case len(stmt) > 0:
		p.PrintLevel(SEMI, stmt, expr)

	default:
		p.PrintLevel(SEMI, expr)
	}
}

func (p *CPrinter) PrintReturn(expr string, tuple bool) {
	if len(expr) == 0 && len(p.ctx.ret_values) > 0 {
		expr = p.Chop(p.ctx.ret_values)
		tuple = strings.Contains(expr, ", ")
	}

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
			results = fmt.Sprintf("std::tuple<%s>", results)
		}

		if len(receiver) > 0 {
			parts := strings.SplitN(receiver, " ", 2)
			receiver = "/* " + parts[1] + " */ " + strings.TrimRight(parts[0], "*") + "::"

			p.ctx.receiver = parts[1]
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
	// for maps a std::pair is returned where key is p.first and value is p.second
	if key == "_" {
		key, value = value, ""
	}

	p.PrintLevel(NONE, "for (auto", key)

	if len(value) > 0 {
		p.Print(",", value)
	}

	p.Print(":", expr, ") ")
}

func (p *CPrinter) PrintSwitch(init, expr string) {
	if len(init) > 0 {
		p.PrintLevel(SEMI, init)
	}
	p.PrintLevel(NONE, "switch (", expr, ")")
}

func (p *CPrinter) PrintCase(expr string) {
	p.ctx.fall_through = false

	if p.ctx.caseType != "" {
		if len(expr) > 0 {
			p.PrintLevel(NL,
				fmt.Sprintf("std::tie(v, ok) = typeAssert<%v>(%v); if (ok) {", p.ctx.caseType, expr))
		} else {
			p.PrintLevel(NL, "} else {")
		}

		return
	}

	if len(expr) > 0 {
		p.PrintLevel(COLON, "case", expr)
	} else {
		p.PrintLevel(NL, "default:")
	}
}

func (p *CPrinter) PrintEndCase() {
	if p.ctx.caseType != "" {
		p.PrintLevel(NONE, "}")
		return
	}

	if !p.ctx.fall_through {
		p.PrintLevel(SEMI, "break")
	}
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
		lhs = fmt.Sprintf("std::tie(%s)", lhs)
	}

	if rtuple {
		rhs = fmt.Sprintf("std::make_tuple(%s)", rhs)
	}

	p.PrintLevel(SEMI, lhs, op, rhs)
}

func (p *CPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, fmt.Sprintf("%s.Send(%s)", ch, value))
}

func (p *CPrinter) FormatIdent(id, itype string) (ret string) {
	switch id {
	case NIL:
		return NULL

	case IOTA:
		ret = strconv.Itoa(p.ctx.iota_count)
		p.ctx.iota_count += 1

	case "string":
		ret = "std::string"

	case "_":
		ret = "std::ignore"

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
	if op == "&^" {
		// AND NOT
		op = "&"
		rhs = "~" + rhs
	}
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *CPrinter) FormatPair(v Pair, t FieldType) (ret string) {
	name, value := v.Name(), v.Value()

	if strings.HasSuffix(value, "]") {
		i := strings.LastIndex(value, "[")
		if i < 0 {
			// it should be an error

		} else {
			arr := value[i:]
			value = value[:i]

			if len(name) > 0 {
				name += arr
			} else {
				value += arr
			}
		}
	}

	if strings.HasPrefix(value, "*") {
		i := strings.LastIndex(value, "*") + 1
		value = value[i:] + value[:i]
	}

	if t == METHOD {
		if len(name) == 0 {
			ret = fmt.Sprintf("// extends %s", value)
		} else {
			ret = "virtual " + fmt.Sprintf(value, name)
		}
	} else if t == RESULT && len(name) > 0 {
		ret = fmt.Sprintf("%s /* %s */", value, name)
		if p.ctx != nil {
			p.ctx.ret_definitions += fmt.Sprintf("%s %s;", value, name)
			p.ctx.ret_values += fmt.Sprintf("%s, ", name)
		}
	} else if t == PARAM && strings.Contains(value, "%s") {
		ret = fmt.Sprintf(value, name)
	} else if t == FIELD && len(name) == 0 {
		ret = getIdentifier(value) + " " + value
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

func (p *CPrinter) FormatArray(alen, elt string) string {
	if alen == "" { // slice
		return fmt.Sprintf("std::vector<%v>", elt)
	} else {
		return fmt.Sprintf("%s[%s]", elt, alen)
	}
}

func (p *CPrinter) FormatArrayIndex(array, index, rtype string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *CPrinter) FormatMapIndex(array, index, rtype string, check bool) string {
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
	return fmt.Sprintf("std::map<%s, %s>", key, elt)
}

func (p *CPrinter) FormatKeyValue(key, value string, isMap bool) string {
	if isMap {
		return fmt.Sprintf("{%s, %s}", key, value)
	}

	// struct
	return fmt.Sprintf(".%s=%s", key, value)
}

func (p *CPrinter) FormatStruct(name, fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct _%s {\n%s}", name, fields)
	} else {
		return "struct{}"
	}
}

func (p *CPrinter) FormatInterface(name, methods string) string {
	if len(methods) > 0 {
		name = "_" + name
		return fmt.Sprintf("/* abstract */ struct %s {\n ~%s(){};\n%s}", name, name, methods)
	} else {
		return "std::any"
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
		// need to rename, to avoid conflicts with C/C++ "time" :(
		fun = "go_" + fun
	}

	if isFuncLit {
		return fmt.Sprintf("[](%s)->%s", args, fun)
	} else if fun == "len" {
		return fmt.Sprintf("%v.size()", args)

		//} else if fun == "make" {
		//	return FormatMake(args)
	} else {
		return fmt.Sprintf("%s(%s)", fun, args)
	}
}

func (p *CPrinter) FormatFuncType(params, results string, withFunc bool) string {
	if len(results) == 0 {
		results = "void"
	} else if IsMultiValue(results) {
		results = fmt.Sprintf("std::tuple<%s>", results)
	}

	// add %%s only if withFunc ?
	return fmt.Sprintf("%s %%s(%s)", results, params)
}

func (p *CPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf(ftype+"%s", "", body)
}

func (p *CPrinter) FormatSelector(pname, sel string, isObject bool) string {
	switch {
	case pname == "io" && sel == "ReadSeeker":
		pname = "std"
		sel = "istream"

	case pname == "io" && sel == "SeekCurrent":
		pname = "std::ios"
		sel = "cur"

	case pname == "io" && sel == "SeekStart":
		pname = "std::ios"
		sel = "beg"

	case pname == "io" && sel == "SeekEnd":
		pname = "std::ios"
		sel = "end"

	case pname == "fmt" && sel == "Printf":
		pname = "std"
		sel = "printf"

	case pname == "fmt" && sel == "Sprintf":
		pname = "std"
		sel = "sprintf"

	case pname == "strconv" && sel == "Itoa":
		pname = "std"
		sel = "to_string"
	}

	if isObject {
		return fmt.Sprintf("%s%s", p.ctx.Selector(pname), sel)
	} else {
		return fmt.Sprintf("%s::%s", pname, sel)
	}
}

func (p *CPrinter) FormatTypeAssert(orig, assert string) string {
	if assert == "type" {
		p.ctx.caseType = orig
		return ""
	}

	return fmt.Sprintf("typeAssert<%v>(%v)", assert, orig)
}

// Guess type and return type and new value
func GuessType(value string) (string, string) {
	vtype := "auto"

	//
	// no value ?
	//
	if len(value) == 0 {
		return vtype, value
	}

	//
	// check if basic type
	//
	switch value[0] {
	case '\'':
		return "char", value
	case '"':
		return "std::string", value

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if strings.Contains(value, ".") || strings.Contains(value, "E") {
			return "float64", value
		}
		return "int", value
	}

	//
	// boolean values
	//
	switch value {
	case "true", "false":
		return "bool", value

		//
		// null values
		//
	case NIL, NULL:
		return "void*", value
	}

	//
	// x = make(t) -> t x()
	//
	if strings.HasPrefix(value, "make(") {
		i := strings.IndexAny(value, ",)")
		return value[5:i], value
	}

	//
	// a map
	//
	if strings.HasPrefix(value, "std::map<") {
		// a map
		if p, ok := findMatch(value, '<'); ok {
			return value[:p+1], value
		}
	}

	//
	// an array
	//
	i := strings.IndexAny(value, "[({") // use this instead of s.Contains("[") to catch f(x[])
	if i >= 0 {
		switch value[i] {
		case '[':
			// should be an array
			if p, ok := findMatch(value, '['); ok {
				t, _ := GuessType(value[i+1 : p])
				return t, value
			}

		case '{':
			// should be struct
			return value[:i], value

		}
	}

	//
	// don't know - let's use auto
	//
	return vtype, value
}

func IsMultiValue(expr string) bool {
	return strings.Contains(expr, ",")
}

func FormatMake(args string) string {
	if strings.HasPrefix(args, "std::map<") {
		// make map
		p := strings.LastIndex(args, ">")
		return args[:p+1] + "()"
	} else if strings.Contains(args, "[]") {
		// make slice
		// TODO: this could be make([]x, cap, max)

		p := strings.LastIndex(args, "]")
		slicedef, n := args[:p], args[p+1:]
		if len(n) > 0 {
			// should be ", N"
			n = n[2:]
		}

		return fmt.Sprintf("%s%s]", slicedef, n)
	} else {
		// make chan

		p := strings.LastIndex(args, ">")
		chandef, n := args[:p+1], args[p+1:]
		if len(n) > 0 {
			// should be ", N"
			n = n[2:]
		}

		return fmt.Sprintf("%s(%s)", chandef, n)
	}
}

// findMatch finds the matching closing character given the opening character.
// used to find matching braces or parenthesis with support for nesting.
// NOTE: this version does NOT not check if the closing character is inside quotes.
func findMatch(s string, ch byte) (int, bool) {
	var closing byte
	var cnt int

	open := strings.IndexByte(s, ch)
	if open < 0 {
		return 0, false
	}

	switch ch {
	case '<':
		closing = '>'
	case '[':
		closing = ']'
	case '{':
		closing = '}'
	case '(':
		closing = ')'
	}

	for p := open; p < len(s); p++ {
		if s[p] == ch {
			cnt++
		} else if s[p] == closing {
			cnt--
			if cnt == 0 {
				return p, true
			}
		}
	}

	return 0, false
}

func getIdentifier(s string) string {
	for i, c := range s {
		if c == '<' || c == '[' {
			return s[0:i]
		}
	}

	return s
}
