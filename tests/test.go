package main

import "fmt"

const (
    k = "constant"
    )

type Y interface {
    Hello()
}

type X struct {
    p Y
}

func main() {
    fmt.Println("k=constant", k=="constant", 42, 3.4)
    fmt.Printf("k=constant %d %d %f\n", k=="constant", 42, 3.4)
}
