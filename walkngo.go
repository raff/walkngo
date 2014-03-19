package main

//
// This is the main program for a Go AST printer
//

import (
	"flag"
	"fmt"
	"os"

	"github.com/raff/walkngo/printer"
	"github.com/raff/walkngo/walker"
)

func main() {
	clang := flag.Bool("c", false, "print as C program or Go program")
	debug := flag.Bool("debug", false, "print debug info")

	flag.Parse()

	filename := flag.Args()[0]

	var walker *walkngo.GoWalker
	if *clang {
		var printer printer.CPrinter
		walker = walkngo.NewWalker(&printer, os.Stdout, *debug)
	} else {
		var printer printer.GoPrinter
		walker = walkngo.NewWalker(&printer, os.Stdout, *debug)
	}

	if err := walker.WalkFile(filename); err != nil {
		fmt.Println(err)
	}
}
