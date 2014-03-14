package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func asString(v interface{}) string {
	if v != nil {
		return fmt.Sprintf("%#v", v)
	} else {
		return ""
	}
}

func getCondition(c interface{}) string {
	switch c := c.(type) {
	case *ast.Ident:
		// boolean ?
		return c.Name

	case *ast.BinaryExpr:
		return fmt.Sprintf("%v %v %v", c.X, c.Op.String(), c.Y)

	default:
		return fmt.Sprintf("%#v", c)
	}
}

func getType(t interface{}) string {
	switch t := t.(type) {
	case *ast.Ident:
		return t.Name

	default:
		return fmt.Sprintf("%#v", t)
	}
}

func getFieldList(l *ast.FieldList, omitempty bool) string {
	if l != nil {
		fields := []string{}
		for _, f := range l.List {
			fields = append(fields, getNames(f.Names)+" "+getType(f.Type))
		}

		return "(" + strings.Join(fields, ", ") + ")"
	} else if omitempty {
		return ""
	} else {
		return "()"
	}
}

func getNames(v []*ast.Ident) string {
	names := []string{}

	for _, n := range v {
		names = append(names, n.Name)
	}

	return strings.Join(names, ", ")
}

type GoWalker struct {
}

func (w GoWalker) Visit(node ast.Node) (ret ast.Visitor) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.File:
		fmt.Println("//- package", n.Name)
		for _, d := range n.Decls {
			w.Visit(d)
		}

	case *ast.ValueSpec:
		fmt.Print("//-  ", getNames(n.Names))
		if n.Type != nil {
			fmt.Print(" ", asString(n.Type))
		}
		if n.Values != nil {
			fmt.Printf(" = %#v", n.Values)
		}
		fmt.Println("")

	case *ast.ImportSpec:
		if n.Name != nil {
			fmt.Println("//-  ", n.Name, n.Path.Value)
		} else {
			fmt.Println("//-  ", n.Path.Value)
		}

	case *ast.TypeSpec:
		fmt.Println("//-  ", n.Name, asString(n.Type))

	case *ast.GenDecl:
		fmt.Println("//-", n.Tok.String(), "{")
		for _, s := range n.Specs {
			w.Visit(s)
		}
		fmt.Println("//- }")

	case *ast.FuncDecl:
		fmt.Println("//-")
		fmt.Println("//- func", getFieldList(n.Recv, true), n.Name, getFieldList(n.Type.Params, false), getFieldList(n.Type.Results, true))
		w.Visit(n.Body)

	case *ast.BlockStmt:
		fmt.Println("//- {")
		for _, i := range n.List {
			w.Visit(i)
		}
		fmt.Println("//- }")

	case *ast.IfStmt:
		fmt.Println("//- if (", getCondition(n.Cond), ")")
		w.Visit(n.Body)
		if n.Else != nil {
			fmt.Println("//- else")
			w.Visit(n.Else)
		}

	default:
		fmt.Printf("/* %#v */\n", n)
		ret = w
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

	var walker GoWalker
	ast.Walk(walker, f)
}
