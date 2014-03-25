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

	outdir string
	prefix string
	ext    string
}

func (w Walker) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var outpath string

	if len(w.outdir) > 0 {
		outpath = filepath.Join(w.outdir, path[len(w.prefix):])
	}

	if info.IsDir() {
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." { // assume we want to skip hidden folders
			return filepath.SkipDir
		}

		if len(outpath) > 0 {
			if err := os.MkdirAll(outpath, 0755); err != nil {
				fmt.Println(err)
			}
		}
	} else if strings.HasSuffix(path, ".go") {
		if len(outpath) > 0 {
			outpath = outpath[:len(outpath)-2] + w.ext
			f, err := os.Create(outpath)
			if err != nil {
				fmt.Println(err)
			} else {
				w.SetWriter(f)
				defer f.Close()
			}
		}

		if err := w.WalkFile(path); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func main() {
	clang := flag.Bool("c", false, "print as C program or Go program")
	debug := flag.Bool("debug", false, "print debug info")
	outd := flag.String("outdir", "", "create converted files in outdir")

	flag.Parse()

	var walker Walker

	if *clang {
		walker = Walker{walkngo.NewWalker(&printer.CPrinter{}, os.Stdout, *debug), *outd, "", "cc"}
	} else {
		walker = Walker{walkngo.NewWalker(&printer.GoPrinter{}, os.Stdout, *debug), *outd, "", "go"}
	}

	for _, f := range flag.Args() {
		walker.prefix = f

		filepath.Walk(f, walker.Walk)
	}
}
