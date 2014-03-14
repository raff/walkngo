go-walk
=======

A "walker" for Go AST

This is the "skeleton" for something that can parse a Go source file and print it in some other format or, in any case, a way to understand how Go AST works.

When running with "go run" you should use:

    go run go-walk.go -- go-source-file.go
    
The '--' tell 'go run' to stop looking for go files, otherwise it gets utterly confused.

If you build it you can just run:

   go-walk go-source-file.go

