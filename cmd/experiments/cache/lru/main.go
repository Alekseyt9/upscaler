package main

import (
	"fmt"

	"github.com/Alekseyt9/upscaler/pkg/lrulom"
)

type val struct {
	v int
}

func main() {
	l, _ := lrulom.New[int, val](128, func(key int) (val, error) { return val{v: key}, nil })
	v, ok := l.GetOrLoad(200)
	fmt.Println(v, ok)
}
