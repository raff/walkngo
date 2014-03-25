package walkngo

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"

	"github.com/raff/walkngo/printer"
)

//
// GoWalker is the context for the AST visitor
//
type GoWalker struct {
	p      printer.Printer
	parent ast.Node
	flush  bool
	buffer bytes.Buffer
	writer io.Writer
	debug  bool
}

func NewWalker(p printer.Printer, out io.Writer, debug bool) *GoWalker {
	w := GoWalker{p: p, flush: true, writer: out, debug: debug}
	p.SetWriter(&w.buffer)
	return &w
}

func (w *GoWalker) SetWriter(writer io.Writer) (old io.Writer) {
	w.Flush()

	old, w.writer = w.writer, writer
	return
}

func (w *GoWalker) WalkFile(filename string) error {
	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	w.p.Print(fmt.Sprintf("//source: %s\n", filename))

	ast.Walk(w, f)
	return nil
}

//
// Implement the Visitor interface for GoWalker
//
func (w *GoWalker) Visit(node ast.Node) (ret ast.Visitor) {
	if node == nil {
		return
	}

	if w.debug {
		w.p.Print(fmt.Sprintf("/* Node: %#v */\n", node))
	}

	pparent := w.parent
	w.parent = node

	switch n := node.(type) {
	case *ast.File:
		w.p.PrintPackage(n.Name.String())
		for _, d := range n.Decls {
			w.Visit(d)
		}

	case *ast.ImportSpec:
		w.p.PrintImport(w.parseExpr(n.Name), n.Path.Value)

	case *ast.TypeSpec:
		w.p.PrintType(n.Name.String(), w.parseExpr(n.Type))

	case *ast.ValueSpec:
		vtype := (pparent.(*ast.GenDecl)).Tok.String()
		w.p.PrintValue(vtype, w.parseExpr(n.Type), w.parseNames(n.Names), w.parseExprList(n.Values), len(n.Names) > 1, len(n.Values) > 1)

	case *ast.GenDecl:
		w.p.Print("\n")
		for _, s := range n.Specs {
			w.Visit(s)
		}

	case *ast.FuncDecl:
		w.p.Print("\n")
		w.p.PrintFunc(w.parseFieldList(n.Recv, printer.RECEIVER, ", "),
			n.Name.String(),
			w.parseFieldList(n.Type.Params, printer.PARAM, ", "),
			w.parseFieldList(n.Type.Results, printer.RESULT, ", "))
		w.Visit(n.Body)
		w.p.Print("\n")

	case *ast.BlockStmt:
		w.p.PrintLevel("{\n")
		w.p.UpdateLevel(printer.UP)
		for _, i := range n.List {
			w.Visit(i)
		}
		w.p.UpdateLevel(printer.DOWN)
		w.p.PrintLevel("}")

	case *ast.IfStmt:
		if !w.p.IsSameLine() {
			w.p.Print("\n")
		}
		w.p.PrintIf(w.BufferVisit(n.Init), w.parseExpr(n.Cond))
		w.p.SameLine()
		w.Visit(n.Body)
		if n.Else != nil {
			w.p.SameLine()
			w.p.PrintElse()
			w.p.SameLine()
			w.Visit(n.Else)
		}
		w.p.Print("\n")

	case *ast.ForStmt:
		w.p.Print("\n")
		w.p.PrintFor(w.BufferVisit(n.Init), w.parseExpr(n.Cond), w.BufferVisit(n.Post))
		w.Visit(n.Body)
		w.p.Print("\n")

	case *ast.SwitchStmt:
		w.p.Print("\n")
		w.p.PrintSwitch(w.BufferVisit(n.Init), w.parseExpr(n.Tag))
		w.Visit(n.Body)
		w.p.Print("\n")

	case *ast.TypeSwitchStmt:
		w.p.Print("\n")
		w.p.PrintSwitch(w.BufferVisit(n.Init), w.BufferVisit(n.Assign))
		w.Visit(n.Body)
		w.p.Print("\n")

	case *ast.CaseClause:
		w.p.PrintCase(w.parseExprList(n.List))
		w.p.UpdateLevel(printer.UP)
		for _, i := range n.Body {
			w.Visit(i)
		}
		w.p.UpdateLevel(printer.DOWN)

	case *ast.RangeStmt:
		w.p.Print("\n")
		w.p.PrintRange(w.parseExpr(n.Key), w.parseExpr(n.Value), w.parseExpr(n.X))
		w.Visit(n.Body)
		w.p.Print("\n")

	case *ast.BranchStmt:
		w.p.PrintStmt(n.Tok.String(), w.parseExpr(n.Label))

	case *ast.DeferStmt:
		w.p.PrintStmt("defer", w.parseExpr(n.Call))

	case *ast.GoStmt:
		w.p.PrintStmt("go", w.parseExpr(n.Call))

	case *ast.ReturnStmt:
		w.p.PrintReturn(w.parseExprList(n.Results), len(n.Results) > 1)

	case *ast.ExprStmt:
		w.p.PrintStmt("", w.parseExpr(n.X))

	case *ast.DeclStmt:
		w.Visit(n.Decl)

	case *ast.AssignStmt:
		w.p.PrintAssignment(w.parseExprList(n.Lhs), n.Tok.String(), w.parseExprList(n.Rhs), len(n.Lhs) > 1, len(n.Rhs) > 1)

	case *ast.IncDecStmt:
		w.p.PrintLevel(w.parseExpr(n.X)+n.Tok.String(), " ")

	case *ast.SendStmt:
		w.p.PrintSend(w.parseExpr(n.Chan), w.parseExpr(n.Value))

	case *ast.EmptyStmt:
		w.p.PrintEmpty()

	default:
		w.p.Print(fmt.Sprintf("/* Node: %#v */\n", n))
		ret = w
	}

	w.Flush()

	w.parent = pparent
	return
}

func (w *GoWalker) Flush() {
	if w.flush && w.buffer.Len() > 0 {
		w.buffer.WriteTo(w.writer)
		w.buffer.Reset()
	}
}

func (w *GoWalker) BufferVisit(node ast.Node) (ret string) {
	w.Flush()

	prev := w.flush
	w.flush = false

	w.Visit(node)

	w.flush = prev

	ret = w.buffer.String()
	w.buffer.Reset()

	if prev == true {
		ret = strings.TrimSpace(ret)
	}

	return
}

func (w *GoWalker) parseExpr(expr interface{}) string {
	if expr == nil {
		return ""
	}

	if w.debug {
		w.p.Print(fmt.Sprintf("/* Expr: %#v */\n", expr))
	}

	switch expr := expr.(type) {

	// a name or a predefined constant
	case *ast.Ident:
		if expr == nil {
			return ""
		}
		return w.p.FormatIdent(expr.Name)

		// *thing
	case *ast.StarExpr:
		return "*" + w.parseExpr(expr.X)

		// [len]type
	case *ast.ArrayType:
		return w.p.FormatArray(w.parseExpr(expr.Len), w.parseExpr(expr.Elt))

		// [key]value
	case *ast.MapType:
		return w.p.FormatMap(w.parseExpr(expr.Key), w.parseExpr(expr.Value))

		// interface{ things }
	case *ast.InterfaceType:
		return w.p.FormatInterface(w.parseFieldList(expr.Methods, printer.METHOD, ";\n"))

		// struct{ things }
	case *ast.StructType:
		return w.p.FormatStruct(w.parseFieldList(expr.Fields, printer.FIELD, ";\n"))

		// <-chan type
	case *ast.ChanType:
		chdir := printer.CHAN_BIDI
		if expr.Dir == ast.SEND {
			chdir = printer.CHAN_SEND
		} else if expr.Dir == ast.RECV {
			chdir = printer.CHAN_RECV
		}
		return w.p.FormatChan(chdir, w.parseExpr(expr.Value))

		// (params) (results)
	case *ast.FuncType:
		return w.p.FormatFuncType(
			w.parseFieldList(expr.Params, printer.PARAM, ", "),
			w.parseFieldList(expr.Results, printer.RESULT, ", "))

		// "thing", 0, 1.2, 'x', etc.
	case *ast.BasicLit:
		return w.p.FormatLiteral(expr.Value)

		// type{list}
	case *ast.CompositeLit:
		return fmt.Sprintf("%s{%s}", w.parseExpr(expr.Type), w.parseExprList(expr.Elts))

		// ...type
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", w.parseExpr(expr.Elt))

		// -3
	case *ast.UnaryExpr:
		return w.p.FormatUnary(expr.Op.String(), w.parseExpr(expr.X))

		// 3 + 2
	case *ast.BinaryExpr:
		return w.p.FormatBinary(w.parseExpr(expr.X), expr.Op.String(), w.parseExpr(expr.Y))

		// array[index]
	case *ast.IndexExpr:
		return w.p.FormatArrayIndex(w.parseExpr(expr.X), w.parseExpr(expr.Index))

		// key: value
	case *ast.KeyValueExpr:
		return w.p.FormatKeyValue(w.parseExpr(expr.Key), w.parseExpr(expr.Value))

		// x[low:hi:max]
	case *ast.SliceExpr:
		return w.p.FormatSlice(w.parseExpr(expr.X), w.parseExpr(expr.Low), w.parseExpr(expr.High), w.parseExpr(expr.Max))

		// package.member
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", w.parseExpr(expr.X), w.parseExpr(expr.Sel))

		// funcname(args)
	case *ast.CallExpr:
		return w.p.FormatCall(w.parseExpr(expr.Fun), w.parseExprList(expr.Args)+ifTrue("...", expr.Ellipsis > 0))

		// name.(type)
	case *ast.TypeAssertExpr:
		return fmt.Sprintf("%s.(%s)", w.parseExpr(expr.X), w.exprOr(expr.Type, "type"))

		// (expr)
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", w.parseExpr(expr.X))

		// func(params) (ret) { body }
	case *ast.FuncLit:
		return w.p.FormatFuncLit(w.parseExpr(expr.Type), w.BufferVisit(expr.Body))
	}

	return fmt.Sprintf("/* Expr: %#v */", expr)
}

func (w *GoWalker) parseExprList(l []ast.Expr) string {
	exprs := []string{}
	for _, e := range l {
		exprs = append(exprs, w.parseExpr(e))
	}
	return strings.Join(exprs, ", ")
}

func (w *GoWalker) parseFieldList(l *ast.FieldList, ftype printer.FieldType, sep string) string {
	if l != nil {
		fields := []string{}
		for _, f := range l.List {
			field := printer.Pair{w.parseNames(f.Names), w.parseExpr(f.Type)}
			fields = append(fields, w.p.FormatPair(field, ftype))
		}

		return strings.Join(fields, sep)
	} else {
		return ""
	}
}

func (w *GoWalker) parseNames(v []*ast.Ident) string {
	names := make([]string, len(v))

	for i, n := range v {
		names[i] = n.Name
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
// ifTrue retruns the input value if the condition is true, an empty string otherwise
//
func ifTrue(val string, cond bool) (ret string) {
	if cond {
		ret = val
	}
	return
}
