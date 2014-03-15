package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

const (
	SP = " "
	SC = ";"
	NL = "\n"

	UP   = +1
	DOWN = -1
)

//
// Printer is the interface to be implemented to print a program
//
type Printer interface {
	setWriter(w io.Writer)
	updateLevel(delta int)
	print(values ...string)
	printLevel(values ...string)

	printPackage(name string)
	printImport(name, path string)
	printType(name, typedef string)
	printValue(vtype, names, typedef, values string)
	printFunc(receiver, name, params, results string)
	printFor(init, cond, post string)
	printSwitch(init, expr string)
	printIf(init, cond string)
	printElse()
	printEmpty()
}

//
// GoPrinter implement the Printer interface for Go programs
//
type GoPrinter struct {
	Printer

	level int
	w     io.Writer
}

func (p *GoPrinter) setWriter(w io.Writer) {
	p.w = w
}

func (p *GoPrinter) updateLevel(delta int) {
	p.level += delta
}

func (p *GoPrinter) indent() string {
	return strings.Repeat("  ", p.level)
}

func (p *GoPrinter) print(values ...string) {
	fmt.Fprint(p.w, strings.Join(values, " "))
}

func (p *GoPrinter) printLevel(values ...string) {
	fmt.Fprint(p.w, p.indent(), strings.Join(values, " "))
}

func (p *GoPrinter) printPackage(name string) {
	p.printLevel("package", name, NL)
}

func (p *GoPrinter) printImport(name, path string) {
	p.printLevel("import", name, path, NL)
}

func (p *GoPrinter) printType(name, typedef string) {
	p.printLevel("type", name, typedef, NL)
}

func (p *GoPrinter) printValue(vtype, names, typedef, value string) {
	p.printLevel(vtype, names)
	if len(typedef) > 0 {
		fmt.Fprint(p.w, SP, typedef)
	}
	if len(value) > 0 {
		fmt.Fprint(p.w, " = ", value)
	}
	fmt.Fprintln(p.w)
}

func (p *GoPrinter) printFunc(receiver, name, params, results string) {
	p.printLevel("func ")
	if len(receiver) > 0 {
		fmt.Fprintf(p.w, "(%s) ", receiver)
	}
	fmt.Fprintf(p.w, "%s(%s) ", name, params)
	if len(results) > 0 {
		if strings.ContainsAny(results, " ,") {
			// name type or multiple types
			fmt.Fprintf(p.w, "(%s) ", results)
		} else {
			fmt.Fprint(p.w, results, SP)
		}
	}
}

func (p *GoPrinter) printFor(init, cond, post string) {
	p.printLevel("for ")
	if len(init) > 0 {
		p.print(init)
	}
	if len(init) > 0 || len(post) > 0 {
		p.print("; ")
	}

	p.print(cond)

	if len(post) > 0 {
		p.print(";", post)
	}

	p.print(SP)
}

func (p *GoPrinter) printSwitch(init, expr string) {
	p.printLevel("switch ")
	if len(init) > 0 {
		p.print(init + "; ")
	}
	p.print(expr)
}

func (p *GoPrinter) printIf(init, cond string) {
	p.printLevel("if ")
	if len(init) > 0 {
		p.print(init + "; ")
	}
	p.print(cond, SP)
}

func (p *GoPrinter) printElse() {
	p.print(" else ")
}

func (p *GoPrinter) printEmpty() {
	p.printLevel(";\n")
}

type GoWalker struct {
	p      Printer
	parent ast.Node
	buffer bytes.Buffer
	flush  bool
}

func NewWalker(p Printer) *GoWalker {
	w := GoWalker{p: p, flush: true}
	p.setWriter(&w.buffer)
	return &w
}

//
// Implement the Visitor interface for GoWalker
//
func (w *GoWalker) Visit(node ast.Node) (ret ast.Visitor) {
	if node == nil {
		return
	}

	pparent := w.parent
	w.parent = node

	switch n := node.(type) {
	case *ast.File:
		w.p.printPackage(n.Name.String())
		for _, d := range n.Decls {
			w.Visit(d)
		}

	case *ast.ImportSpec:
		w.p.printImport(identString(n.Name), n.Path.Value)

	case *ast.TypeSpec:
		w.p.printType(n.Name.String(), w.parseExpr(n.Type))

	case *ast.ValueSpec:
		vtype := (pparent.(*ast.GenDecl)).Tok.String()
		w.p.printValue(vtype, w.parseNames(n.Names), w.parseExpr(n.Type), w.parseExprList(n.Values))

	case *ast.GenDecl:
		for _, s := range n.Specs {
			w.Visit(s)
		}

	case *ast.FuncDecl:
		w.p.printFunc(w.parseFieldList(n.Recv, ","),
			n.Name.String(),
			w.parseFieldList(n.Type.Params, ","),
			w.parseFieldList(n.Type.Results, ","))
		w.Visit(n.Body)
		w.p.print(NL)

	case *ast.BlockStmt:
		w.p.print("{\n")
		w.p.updateLevel(UP)
		for _, i := range n.List {
			w.Visit(i)
		}
		w.p.updateLevel(DOWN)
		w.p.printLevel("}")

	case *ast.IfStmt:
		w.p.printIf(w.BufferVisit(n.Init), w.parseExpr(n.Cond))
		w.Visit(n.Body)
		if n.Else != nil {
			w.p.printElse()
			w.Visit(n.Else)
		}
		w.p.print(NL)

	case *ast.ForStmt:
		w.p.printFor(w.BufferVisit(n.Init), w.parseExpr(n.Cond), w.BufferVisit(n.Post))
		w.Visit(n.Body)
		w.p.print(NL)

	case *ast.SwitchStmt:
		w.p.printSwitch(w.BufferVisit(n.Init), w.parseExpr(n.Tag))
		w.Visit(n.Body)
		w.p.print(NL)

	case *ast.TypeSwitchStmt:
		w.p.printSwitch(w.BufferVisit(n.Init), w.BufferVisit(n.Assign))
		w.Visit(n.Body)
		w.p.print(NL)

	case *ast.CaseClause:
		if len(n.List) > 0 {
			w.p.printLevel("case", w.parseExprList(n.List), ":", NL)
		} else {
			w.p.printLevel("default:", NL)
		}
		w.p.updateLevel(UP)
		for _, i := range n.Body {
			w.Visit(i)
		}
		w.p.updateLevel(DOWN)

	case *ast.RangeStmt:
		w.p.printLevel("for", w.parseExpr(n.Key), ",", w.parseExpr(n.Value), ":= range", w.parseExpr(n.X))
		w.Visit(n.Body)
		w.p.print(NL)

	case *ast.BranchStmt:
		w.p.printLevel(n.Tok.String(), identString(n.Label), NL)

	case *ast.DeferStmt:
		w.p.printLevel("defer", w.parseExpr(n.Call), NL)

	case *ast.GoStmt:
		w.p.printLevel("go", w.parseExpr(n.Call), NL)

	case *ast.ReturnStmt:
		w.p.printLevel("return", w.parseExprList(n.Results), NL)

	case *ast.ExprStmt:
		w.p.printLevel(w.parseExpr(n.X), NL)

	case *ast.DeclStmt:
		w.Visit(n.Decl)

	case *ast.AssignStmt:
		w.p.printLevel(w.parseExprList(n.Lhs), n.Tok.String(), w.parseExprList(n.Rhs), NL)

	case *ast.IncDecStmt:
		w.p.printLevel(w.parseExpr(n.X)+n.Tok.String(), SP)

	case *ast.EmptyStmt:
		w.p.printEmpty()

	default:
		w.p.print(fmt.Sprintf("/* Node: %#v */\n", n))
		ret = w
	}

	w.Flush()

	w.parent = pparent
	return
}

func (w *GoWalker) Flush() {
	if w.flush && w.buffer.Len() > 0 {
		fmt.Print(w.buffer.String())
		w.buffer.Reset()
	}
}

func (w *GoWalker) BufferVisit(node ast.Node) (ret string) {
	w.Flush()

	prev := w.flush
	w.flush = false
	w.Visit(node)
	w.flush = prev

	ret = strings.TrimSpace(w.buffer.String())

	if w.flush {
		w.buffer.Reset()
	}

	return
}

func (w *GoWalker) parseExpr(expr interface{}) string {
	if expr == nil {
		return ""
	}

	switch expr := expr.(type) {

	// a name
	case *ast.Ident:
		return expr.Name

		// *thing
	case *ast.StarExpr:
		return "*" + w.parseExpr(expr.X)

		// [len]type
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", w.parseExpr(expr.Len), w.parseExpr(expr.Elt))

		// [key]value
	case *ast.MapType:
		return fmt.Sprintf("[%s]%s", w.parseExpr(expr.Key), w.parseExpr(expr.Value))

		// interface{ things }
	case *ast.InterfaceType:
		return fmt.Sprintf("interface{%s}", w.parseFieldList(expr.Methods, ";"))

		// struct{ things }
	case *ast.StructType:
		return fmt.Sprintf("struct{%s}", w.parseFieldList(expr.Fields, ";"))

		// (params...) (result)
	case *ast.FuncType:
		return fmt.Sprintf("(%s) %s", w.parseFieldList(expr.Params, ","), wrapIf(w.parseFieldList(expr.Results, ",")))

		// "thing", 0, true, false, nil
	case *ast.BasicLit:
		return fmt.Sprintf("%v", expr.Value)

		// type{list}
	case *ast.CompositeLit:
		return fmt.Sprintf("%s{%s}", w.parseExpr(expr.Type), w.parseExprList(expr.Elts))

		// ...type
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", w.parseExpr(expr.Elt))

		// -3
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", expr.Op.String(), w.parseExpr(expr.X))

		// 3 + 2
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", w.parseExpr(expr.X), expr.Op.String(), w.parseExpr(expr.Y))

		// array[index]
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", w.parseExpr(expr.X), w.parseExpr(expr.Index))

		// key: value
	case *ast.KeyValueExpr:
		return fmt.Sprintf("%s: %s", w.parseExpr(expr.Key), w.parseExpr(expr.Value))

		// x[low:hi:max]
	case *ast.SliceExpr:
		if expr.Max == nil {
			return fmt.Sprintf("%s[%s:%s]", w.parseExpr(expr.X), w.parseExpr(expr.Low), w.parseExpr(expr.High))
		} else {
			return fmt.Sprintf("%s[%s:%s:%s]", w.parseExpr(expr.X), w.parseExpr(expr.Low), w.parseExpr(expr.High), w.parseExpr(expr.Max))
		}

		// package.member
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", w.parseExpr(expr.X), w.parseExpr(expr.Sel))

		// funcname(args)
	case *ast.CallExpr:
		return fmt.Sprintf("%s(%s%s)", w.parseExpr(expr.Fun), w.parseExprList(expr.Args), ifTrue("...", expr.Ellipsis > 0))

		// name.(type)
	case *ast.TypeAssertExpr:
		return fmt.Sprintf("%s.(%s)", w.parseExpr(expr.X), w.exprOr(expr.Type, "type"))

		// (expr)
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", w.parseExpr(expr.X))

		// func(params) (ret) { body }
	case *ast.FuncLit:
		/* &ast.FuncLit{Type:(*ast.FuncType)(0x2103115e0), Body:(*ast.BlockStmt)(0x2102ccf90)} */
		return fmt.Sprintf("func%s %s", w.parseExpr(expr.Type), w.BufferVisit(expr.Body))

	default:
		return fmt.Sprintf("/* Expr: %#v */", expr)
	}
}

func (w *GoWalker) parseExprList(l []ast.Expr) string {
	exprs := []string{}
	for _, e := range l {
		exprs = append(exprs, w.parseExpr(e))
	}
	return strings.Join(exprs, ", ")
}

func (w *GoWalker) parseFieldList(l *ast.FieldList, sep string) string {
	if l != nil {
		fields := []string{}
		for _, f := range l.List {
			field := w.parseNames(f.Names)
			if len(field) > 0 {
				field += " " + w.parseExpr(f.Type)
			} else {
				field = w.parseExpr(f.Type)
			}
			fields = append(fields, field)
		}

		return strings.Join(fields, sep)
	} else {
		return ""
	}
}

func (w *GoWalker) parseNames(v []*ast.Ident) string {
	names := []string{}

	for _, n := range v {
		names = append(names, n.Name)
	}

	return strings.Join(names, ", ")
}

func (w *GoWalker) exprOr(expr ast.Expr, v string) string {
	if expr != nil {
		return w.parseExpr(expr)
	} else {
		return v
	}
}

//
// identString return the Ident name or ""
// to use when it's ok to have an empty part (and you don't want to see '<nil>')
//
func identString(i *ast.Ident) (ret string) {
	if i != nil {
		ret = i.Name
	}
	return
}

//
// ifTrue retruns the input value if the condition is true, an empty string otherwise
//
func ifTrue(val string, cond bool) (ret string) {
	if cond {
		ret = val
	}
	return
}

func wrapIf(val string) (ret string) {
	if len(val) > 0 {
		ret = fmt.Sprintf("(%s)", val)
	}
	return
}

func main() {
	args := os.Args[1:] // skip program name
	if args[0] == "--" {
		// skip - this is to fool "go run"
		args = args[1:]
	}

	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, args[0], nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	var printer GoPrinter
	var walker = NewWalker(&printer)
	ast.Walk(walker, f)
}
