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

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if info.IsDir() {
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." { // assume we want to skip hidden folders
			return filepath.SkipDir
		}
	} else if strings.HasSuffix(path, ".go") {
		fmt.Println("//source:", path)

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
		walker = Walker{walkngo.NewWalker(&printer.CPrinter{}, os.Stdout, *debug)}
	} else {
		walker = Walker{walkngo.NewWalker(&printer.GoPrinter{}, os.Stdout, *debug)}
	}

	for _, f := range flag.Args() {
		filepath.Walk(f, walker.Walk)
	}
}
