package main

import (
	"fmt"
	"os"

	"github.com/raff/walkngo/printer"
	"github.com/raff/walkngo/walker"
)

func main() {
	golang := true
	args := os.Args[1:] // skip program name

	for len(args) > 0 {
		if args[0] == "--" {
			// skip - this is to fool "go run"

		} else if args[0] == "-go" {
			golang = true
		} else if args[0] == "-c" {
			golang = false
		} else {
			break
		}

		args = args[1:]
	}

	var walker *walkngo.GoWalker
	if golang {
		var printer printer.GoPrinter
		walker = walkngo.NewWalker(&printer, os.Stdout)
	} else {
		var printer printer.CPrinter
		walker = walkngo.NewWalker(&printer, os.Stdout)
	}

	if err := walker.WalkFile(args[0]); err != nil {
		fmt.Println(err)
	}
}
