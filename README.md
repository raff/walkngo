walkngo
=======

A "walker" for Go AST

This is the "skeleton" for something that can parse a Go source file and print it in some other format or, in any case, a way to understand how Go AST works.

The walker package contains the AST walker/visitor, the printer package contains modules that can "print out" the tree as different (probably broken) languages.

The "GoPrinter" module generates a Go source that compile and should work just as good as the original. The "CPrinter" module tries to convert the Go source file to C.

The main program accepts a -c or a -go arguments to select the output.

   walkngo -c walkngo.go

Note:
=====

When running with "go run" you should use:

    go run walkngo.go -- go-source-file.go
    
The '--' tell 'go run' to stop looking for go files, otherwise it gets utterly confused.

If you build it you can just run:

    walkngo go-source-file.go
