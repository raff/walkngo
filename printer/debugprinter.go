package printer

import (
	"fmt"
	"io"
)

//
// DebugPrinter wraps a Printer with debug messages
//
type DebugPrinter struct {
	P Printer
}

func (d *DebugPrinter) SetWriter(w io.Writer) {
	d.P.SetWriter(w)
}

func (d *DebugPrinter) UpdateLevel(delta int) {
	d.P.UpdateLevel(delta)
}

func (d *DebugPrinter) SameLine() {
	d.P.SameLine()
}

func (d *DebugPrinter) IsSameLine() bool {
	return d.P.IsSameLine()
}

func (d *DebugPrinter) Chop(line string) string {
	return d.P.Chop(line)
}

func (d *DebugPrinter) Print(values ...string) {
	d.P.Print(values...)
}

func (d *DebugPrinter) PrintLevel(term string, values ...string) {
	d.P.PrintLevel(term, values...)
}

func (d *DebugPrinter) PrintBlockStart() {
	d.P.PrintBlockStart()
}

func (d *DebugPrinter) PrintBlockEnd() {
	d.P.PrintBlockEnd()
}

func (d *DebugPrinter) PrintPackage(name string) {
	fmt.Println("/* PrintPackage", name, "*/")
	d.P.PrintPackage(name)
}

func (d *DebugPrinter) PrintImport(name, path string) {
	fmt.Println("/* PrintImport", name, path, "*/")
	d.P.PrintImport(name, path)
}

func (d *DebugPrinter) PrintType(name, typedef string) {
	fmt.Println("/* PrintType", name, typedef, "*/")
	d.P.PrintType(name, typedef)
}

func (d *DebugPrinter) PrintValue(vtype, typedef, names, values string, ntuple, vtuple bool) {
	fmt.Println("/* PrintValue", vtype, typedef, names, values, ntuple, vtuple, "*/")
	d.P.PrintValue(vtype, typedef, names, values, ntuple, vtuple)
}

func (d *DebugPrinter) PrintStmt(stmt, expr string) {
	fmt.Println("/* PrintStmt", stmt, expr, "*/")
	d.P.PrintStmt(stmt, expr)
}

func (d *DebugPrinter) PrintReturn(expr string, tuple bool) {
	fmt.Println("/* PrintReturn", expr, tuple, "*/")
	d.P.PrintReturn(expr, tuple)
}

func (d *DebugPrinter) PrintFunc(receiver, name, params, results string) {
	fmt.Println("/* PrintFunc", receiver, name, params, results, "*/")
	d.P.PrintFunc(receiver, name, params, results)
}

func (d *DebugPrinter) PrintFor(init, cond, post string) {
	fmt.Println("/* PrintFor", init, cond, post, "*/")
	d.P.PrintFor(init, cond, post)
}

func (d *DebugPrinter) PrintRange(key, value, expr string) {
	fmt.Println("/* PrintRange", key, value, expr, "*/")
	d.P.PrintRange(key, value, expr)
}

func (d *DebugPrinter) PrintSwitch(init, expr string) {
	fmt.Println("/* PrintSwitch", init, expr, "*/")
	d.P.PrintSwitch(init, expr)
}

func (d *DebugPrinter) PrintCase(expr string) {
	fmt.Println("/* PrintCase", expr, "*/")
	d.P.PrintCase(expr)
}

func (d *DebugPrinter) PrintEndCase() {
	fmt.Println("/* PrintEndCase", "*/")
	d.P.PrintEndCase()
}

func (d *DebugPrinter) PrintIf(init, cond string) {
	fmt.Println("/* PrintIf", init, cond, "*/")
	d.P.PrintIf(init, cond)
}

func (d *DebugPrinter) PrintElse() {
	fmt.Println("/* PrintElse", "*/")
	d.P.PrintElse()
}

func (d *DebugPrinter) PrintEmpty() {
	fmt.Println("/* PrintEmpty", "*/")
	d.P.PrintEmpty()
}

func (d *DebugPrinter) PrintAssignment(lhs, op, rhs string, ltuple, rtuple bool) {
	fmt.Println("/* PrintAssignment", lhs, op, rhs, ltuple, rtuple, "*/")
	d.P.PrintAssignment(lhs, op, rhs, ltuple, rtuple)
}

func (d *DebugPrinter) PrintSend(ch, value string) {
	fmt.Println("/* PrintSend", ch, value, "*/")
	d.P.PrintSend(ch, value)
}

func (d *DebugPrinter) FormatIdent(id string) string {
	fmt.Println("/* FormatIdent", id, "*/")
	return d.P.FormatIdent(id)
}

func (d *DebugPrinter) FormatLiteral(lit string) string {
	fmt.Println("/* FormatLiteral", lit, "*/")
	return d.P.FormatLiteral(lit)
}

func (d *DebugPrinter) FormatCompositeLit(typedef, elt string) string {
	fmt.Println("/* FormatCompositeLit", typedef, elt, "*/")
	return d.P.FormatCompositeLit(typedef, elt)
}

func (d *DebugPrinter) FormatEllipsis(expr string) string {
	fmt.Println("/* FormatEllipsis", expr, "*/")
	return d.P.FormatEllipsis(expr)
}

func (d *DebugPrinter) FormatStar(expr string) string {
	fmt.Println("/* FormatStar", expr, "*/")
	return d.P.FormatStar(expr)
}

func (d *DebugPrinter) FormatParen(expr string) string {
	fmt.Println("/* FormatParen", expr, "*/")
	return d.P.FormatParen(expr)
}

func (d *DebugPrinter) FormatUnary(op, operand string) string {
	fmt.Println("/* FormatUnary", op, operand, "*/")
	return d.P.FormatUnary(op, operand)
}

func (d *DebugPrinter) FormatBinary(lhs, op, rhs string) string {
	fmt.Println("/* FormatBinary", lhs, op, rhs, "*/")
	return d.P.FormatBinary(lhs, op, rhs)
}

func (d *DebugPrinter) FormatPair(v Pair, t FieldType) string {
	fmt.Println("/* FormatPair", v, t, "*/")
	return d.P.FormatPair(v, t)
}

func (d *DebugPrinter) FormatArray(len, elt string) string {
	fmt.Println("/* FormatArray", len, elt, "*/")
	return d.P.FormatArray(len, elt)
}

func (d *DebugPrinter) FormatArrayIndex(array, index string) string {
	fmt.Println("/* FormatArrayIndex", array, index, "*/")
	return d.P.FormatArrayIndex(array, index)
}

func (d *DebugPrinter) FormatSlice(slice, low, high, max string) string {
	fmt.Println("/* FormatSlice", low, high, max, "*/")
	return d.P.FormatSlice(slice, low, high, max)
}

func (d *DebugPrinter) FormatMap(key, elt string) string {
	fmt.Println("/* FormatMap", key, elt, "*/")
	return d.P.FormatMap(key, elt)
}

func (d *DebugPrinter) FormatKeyValue(key, value string) string {
	fmt.Println("/* FormatKeyValue", key, value, "*/")
	return d.P.FormatKeyValue(key, value)
}

func (d *DebugPrinter) FormatStruct(fields string) string {
	fmt.Println("/* FormatStruct", fields, "*/")
	return d.P.FormatStruct(fields)
}

func (d *DebugPrinter) FormatInterface(methods string) string {
	fmt.Println("/* FormatInterface", methods, "*/")
	return d.P.FormatInterface(methods)
}

func (d *DebugPrinter) FormatChan(chdir, mtype string) string {
	fmt.Println("/* FormatChan", chdir, mtype, "*/")
	return d.P.FormatChan(chdir, mtype)
}

func (d *DebugPrinter) FormatCall(fun, args string, isFuncLit bool) string {
	fmt.Println("/* FormatCall", fun, args, isFuncLit, "*/")
	return d.P.FormatCall(fun, args, isFuncLit)
}

func (d *DebugPrinter) FormatFuncType(params, results string) string {
	fmt.Println("/* FormatFuncType", params, results, "*/")
	return d.P.FormatFuncType(params, results)
}

func (d *DebugPrinter) FormatFuncLit(ftype, body string) string {
	fmt.Println("/* FormatFuncLit", ftype, body, "*/")
	return d.P.FormatFuncLit(ftype, body)
}

func (d *DebugPrinter) FormatSelector(pname, sel string, isObject bool) string {
	fmt.Println("/* FormatSelector", pname, sel, isObject, "*/")
	return d.P.FormatSelector(pname, sel, isObject)
}

func (d *DebugPrinter) FormatTypeAssert(orig, assert string) string {
	fmt.Println("/* FormatTypeAssert", orig, assert, "*/")
	return d.P.FormatTypeAssert(orig, assert)
}
