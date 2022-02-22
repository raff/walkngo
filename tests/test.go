package main

import "fmt"

const (
    k = "constant"
    )

var (
    I = 22
    J = "ehlo"
    K = []int{1,2,3}
    )

type Y interface {
    Hello()
}

type X struct {
    p Y
}

func testTypeAssert(t interface{}) string {
    if s, ok := t.(string); ok {
        return s
    }

    return ""
}

func testTypeSwitch(t interface{}) string {
    switch v := t.(type) {
    case string:
        return v

    case int, int8, int16, int32, int64:
        return fmt.Sprintf("%d", v)
    }

    return ""
}

func main() {
    fmt.Println("k=constant", k=="constant", 42, 3.4)
    fmt.Printf("k=constant %d %d %f\n", k=="constant", 42, 3.4)

    ch := make(chan int, 3)
    m := map[string]int{"a":1,"b":2}
    a := []int{}

    if i := a[22]; i < 10 {
        ch <- m["x"]
    }
}
