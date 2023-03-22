package printer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	NULL = "null"
)

// ZigPrinter implement the Printer interface for Zig programs
type ZigPrinter struct {
	Printer

	level    int
	sameline bool
	w        io.Writer

	ctx *ZigContext
}

// ZigContext is the context for a (function) block
type ZigContext struct {
	context ContextType

	iota_count int // incremented when 'const n = iota' or 'const n'

	deferred int // used to generate unique names for "defer" callbacks

	receiver        string // the name of the receiver, to be converted to "this"
	ret_definitions string // used to define return variables
	ret_values      string // used to "fill" empty returns

	fall_through bool // fall through next case in switch

	caseType string // the object of a switch type assertion

	next *ZigContext
}

func (ctx *ZigContext) Selector(s string) string {
	if ctx != nil && ctx.receiver == s {
		s = "self"
	}

	return fmt.Sprintf("%s.", s)
}

func (p *ZigPrinter) Reset() {
	p.level = 0
	p.sameline = false

	p.ctx = nil
}

func (p *ZigPrinter) PushContext(c ContextType) {
	if p.ctx == nil {
		p.ctx = &ZigContext{context: c}
		return
	}

	p.ctx = &ZigContext{
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

func (p *ZigPrinter) PopContext() {
	p.ctx = p.ctx.next
}

func (p *ZigPrinter) SetWriter(w io.Writer) {
	p.w = w
}

func (p *ZigPrinter) UpdateLevel(delta int) {
	p.level += delta
}

func (p *ZigPrinter) SameLine() {
	p.sameline = true
}

func (p *ZigPrinter) IsSameLine() bool {
	return p.sameline
}

func (p *ZigPrinter) Chop(line string) string {
	return strings.TrimRight(line, COMMA)
}

func (p *ZigPrinter) indent() string {
	if p.sameline {
		p.sameline = false
		return ""
	}

	return strings.Repeat("  ", p.level)
}

func (p *ZigPrinter) Print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *ZigPrinter) PrintLevel(term string, values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
}

func (p *ZigPrinter) PrintLevelIn(term string, values ...string) {
	p.level -= 1
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "), term)
	p.level += 1
}

func (p *ZigPrinter) PrintBlockStart(b BlockType, empty bool) {
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

func (p *ZigPrinter) PrintBlockEnd(b BlockType) {
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

func (p *ZigPrinter) PrintPackage(name string) {
	p.PrintLevel(NL, "//package", name)
}

func (p *ZigPrinter) PrintImport(name, path string) {
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

func (p *ZigPrinter) PrintType(name, typedef string) {
	if strings.Contains(typedef, "%") {
		// FuncType
		p.PrintLevel(SEMI, "typedef", fmt.Sprintf(typedef, "("+name+")"))
	} else {
		p.PrintLevel(SEMI, "typedef", typedef, name)
	}
}

func (p *ZigPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	if vtype == "const" && len(values) == 0 {
		values = p.FormatIdent(IOTA, "")
	}

	if typedef == "" {
		typedef, values = zGuessType(values)
	} else if strings.Contains(typedef, "[") {
		i := strings.Index(typedef, "[")
		names += typedef[i:]
		typedef = typedef[:i]
	}

	if ntuple && len(values) > 0 {
		names = fmt.Sprintf("std::tie(%s)", names)
	}

	if typedef == "" {
		p.PrintLevel(NONE, vtype, names)
	} else {
		p.PrintLevel(NONE, vtype, names, ":", typedef)
	}

	if len(values) > 0 {
		if vtuple {
			values = fmt.Sprintf("std::make_tuple(%s)", values)
		}

		p.Print(" =", values)
	}
	p.Print(";\n")
}

func (p *ZigPrinter) PrintStmt(stmt, expr string) {
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

func (p *ZigPrinter) PrintReturn(expr string, tuple bool) {
	if len(expr) == 0 && len(p.ctx.ret_values) > 0 {
		expr = p.Chop(p.ctx.ret_values)
		tuple = strings.Contains(expr, ", ")
	}

	if tuple {
		expr = fmt.Sprintf("make_tuple(%s)", expr)
	}

	p.PrintStmt("return", expr)
}

func (p *ZigPrinter) PrintFunc(receiver, name, params, results string) {
	if len(receiver) == 0 && len(params) == 0 && len(results) == 0 && name == "main" {
		// the "main"
		results = "anyerror!void"
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

	fmt.Fprintf(p.w, "fn %s%s(%s) %s", receiver, name, params, results)
}

func (p *ZigPrinter) PrintFor(init, cond, post string) {
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

func (p *ZigPrinter) PrintRange(key, value, expr string) {
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

func (p *ZigPrinter) PrintSwitch(init, expr string) {
	if len(init) > 0 {
		p.PrintLevel(SEMI, init)
	}
	p.PrintLevel(NONE, "switch (", expr, ")")
}

func (p *ZigPrinter) PrintCase(expr string) {
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

func (p *ZigPrinter) PrintEndCase() {
	if p.ctx.caseType != "" {
		p.PrintLevel(NONE, "}")
		return
	}

	if !p.ctx.fall_through {
		p.PrintLevel(SEMI, "break")
	}
}

func (p *ZigPrinter) PrintIf(init, cond string) {
	if len(init) > 0 {
		p.PrintLevel(NONE, init+" if ")
	} else {
		p.PrintLevel(NONE, "if ")
	}
	p.Print("(", cond, ") ")
}

func (p *ZigPrinter) PrintElse() {
	p.Print(" else ")
}

func (p *ZigPrinter) PrintEmpty() {
	p.PrintLevel(SEMI, "")
}

func (p *ZigPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	if op == ":=" {
		// := means there are new variables to be declared (but of course I don't know the real type)
		rtype, rvalue := zGuessType(rhs)
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

func (p *ZigPrinter) PrintSend(ch, value string) {
	p.PrintLevel(SEMI, fmt.Sprintf("%s.Send(%s)", ch, value))
}

func (p *ZigPrinter) FormatIdent(id, itype string) (ret string) {
	switch id {
	case NIL:
		return NULL

	case IOTA:
		ret = strconv.Itoa(p.ctx.iota_count)
		p.ctx.iota_count += 1

	case "string":
		return "[]const u8"

	case "_":
		return "_"

	case "int8":
		return "i8"

	case "int16":
		return "i16"

	case "int32", "int":
		return "i32"

	case "int64":
		return "i64"

	case "uint8":
		return "u8"

	case "uint16":
		return "u16"

	case "uint32", "uint":
		return "u32"

	case "uint64":
		return "u64"

	case "float32":
		return "f32"

	case "float64":
		return "f64"

	default:
		return id
	}

	return
}

func (p *ZigPrinter) FormatLiteral(lit string) string {
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

func (p *ZigPrinter) FormatCompositeLit(typedef, elt string) string {
	return fmt.Sprintf("%s{%s}", typedef, elt)
}

func (p *ZigPrinter) FormatEllipsis(expr string) string {
	return fmt.Sprintf("...%s", expr)
}

func (p *ZigPrinter) FormatStar(expr string) string {
	return "*" + expr
}

func (p *ZigPrinter) FormatParen(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (p *ZigPrinter) FormatUnary(op, operand string) string {
	if op == "<-" {
		return fmt.Sprintf("%s.Receive()", operand)
	}

	return fmt.Sprintf("%s%s", op, operand)
}

func (p *ZigPrinter) FormatBinary(lhs, op, rhs string) string {
	if op == "&^" {
		// AND NOT
		op = "&"
		rhs = "~" + rhs
	}
	return fmt.Sprintf("%s %s %s", lhs, op, rhs)
}

func (p *ZigPrinter) FormatPair(v Pair, t FieldType) (ret string) {
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
		ret = name + ": " + value
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

func (p *ZigPrinter) FormatArray(alen, elt string) string {
	if alen == "" { // slice
		return fmt.Sprintf("[]%v", elt)
	} else {
		return fmt.Sprintf("%s[%s]", elt, alen)
	}
}

func (p *ZigPrinter) FormatArrayIndex(array, index, rtype string) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *ZigPrinter) FormatMapIndex(array, index, rtype string, check bool) string {
	return fmt.Sprintf("%s[%s]", array, index)
}

func (p *ZigPrinter) FormatSlice(slice, low, high, max string) string {
	if max == "" {
		return fmt.Sprintf("%s[%s:%s]", slice, low, high)
	} else {
		return fmt.Sprintf("%s[%s:%s:%s]", slice, low, high, max)
	}
}

func (p *ZigPrinter) FormatMap(key, elt string) string {
	return fmt.Sprintf("std::map<%s, %s>", key, elt)
}

func (p *ZigPrinter) FormatKeyValue(key, value string, isMap bool) string {
	if isMap {
		return fmt.Sprintf("{%s, %s}", key, value)
	}

	// struct
	return fmt.Sprintf(".%s=%s", key, value)
}

func (p *ZigPrinter) FormatStruct(name, fields string) string {
	if len(fields) > 0 {
		return fmt.Sprintf("struct _%s {\n%s}", name, fields)
	} else {
		return "struct{}"
	}
}

func (p *ZigPrinter) FormatInterface(name, methods string) string {
	if len(methods) > 0 {
		name = "_" + name
		return fmt.Sprintf("/* abstract */ struct %s {\n ~%s(){};\n%s}", name, name, methods)
	} else {
		return "std::any"
	}
}

func (p *ZigPrinter) FormatChan(chdir, mtype string) string {
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

func (p *ZigPrinter) FormatCall(fun, args string, isFuncLit bool) string {
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

func (p *ZigPrinter) FormatFuncType(params, results string, withFunc bool) string {
	if len(results) == 0 {
		results = "void"
	} else if IsMultiValue(results) {
		results = fmt.Sprintf("std::tuple<%s>", results)
	}

	// add %%s only if withFunc ?
	return fmt.Sprintf("%s %%s(%s)", results, params)
}

func (p *ZigPrinter) FormatFuncLit(ftype, body string) string {
	return fmt.Sprintf(ftype+"%s", "", body)
}

func (p *ZigPrinter) FormatSelector(pname, sel string, isObject bool) string {
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

func (p *ZigPrinter) FormatTypeAssert(orig, assert string) string {
	if assert == "type" {
		p.ctx.caseType = orig
		return ""
	}

	return fmt.Sprintf("typeAssert<%v>(%v)", assert, orig)
}

// Guess type and return type and new value
func zGuessType(value string) (string, string) {
	vtype := ""

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
		return "[]const u8", value

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if strings.Contains(value, ".") || strings.Contains(value, "E") {
			return "f64", value
		}
		return "i32", value
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
		if p, ok := FindMatch(value, '<', '>'); ok {
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
			if p, ok := FindMatch(value, '[', ']'); ok {
				t, _ := zGuessType(value[i+1 : p])
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

func getIdentifier(s string) string {
	for i, c := range s {
		if c == '<' || c == '[' {
			return s[0:i]
		}
	}

	return s
}
