package main

//
// This is the main program for a Go AST printer
//

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raff/walkngo/printer"
	"github.com/raff/walkngo/walker"
)

type Walker struct {
	*walkngo.GoWalker
}

func (w Walker) Walk(path string, info os.FileInfo, err error) error {
	fmt.Println()
	fmt.Println("//source:", path)

	if !info.IsDir() && strings.HasSuffix(path, ".go") {
		if err := w.WalkFile(path); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func main() {
	clang := flag.Bool("c", false, "print as C program or Go program")
	debug := flag.Bool("debug", false, "print debug info")

	flag.Parse()

	var walker Walker
	if *clang {
		var printer printer.CPrinter
		walker = Walker{walkngo.NewWalker(&printer, os.Stdout, *debug)}
	} else {
		var printer printer.GoPrinter
		walker = Walker{walkngo.NewWalker(&printer, os.Stdout, *debug)}
	}

	for _, f := range flag.Args() {
		filepath.Walk(f, walker.Walk)
	}
}
