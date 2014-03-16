package printer

import "io"

const (
	UP   = +1
	DOWN = -1
)

//
// Printer is the interface to be implemented to print a program
//
type Printer interface {
	SetWriter(w io.Writer)
	UpdateLevel(delta int)
	SameLine()
	Print(values ...string)
	PrintLevel(values ...string)

	PrintPackage(name string)
	PrintImport(name, path string)
	PrintType(name, typedef string)
	PrintValue(vtype, names, typedef, values string)
	PrintStmt(stmt, expr string)
	PrintFunc(receiver, name, params, results string)
	PrintFor(init, cond, post string)
	PrintSwitch(init, expr string)
	PrintIf(init, cond string)
	PrintElse()
	PrintEmpty()
}
