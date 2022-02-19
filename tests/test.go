package main

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
    var x X

    x.p.Hello()
}
