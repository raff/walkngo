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
	debug := flag.Bool("debug", false, "print AST nodes")
	pdebug := flag.Bool("debug-printer", false, "print Printer calls")
	outd := flag.String("outdir", "", "create converted files in outdir")
	lang := flag.String("lang", "go", "convert to specified language (go, c, rust)")

	flag.Parse()

	var p printer.Printer

	switch *lang {
	case "c":
		p = &printer.CPrinter{}
		*lang = "cc"

	case "go":
		p = &printer.GoPrinter{}
		*lang = "go"

	case "rust":
		p = &printer.RustPrinter{}
		*lang = "rust"

	default:
		fmt.Println("unsupported language", *lang, "use c, go or rust")
		return
	}

	if *pdebug {
		p = &printer.DebugPrinter{p}
	}

	walker := Walker{walkngo.NewWalker(p, os.Stdout, *debug), *outd, "", *lang}

	for _, f := range flag.Args() {
		walker.prefix = f

		filepath.Walk(f, walker.Walk)
	}
}
