package main

import (
	"fmt"

	GoOST "github.com/Alekseyt9/upscaler/pkg/ost"
)

func main() {

	tree := GoOST.New()

	tree.Insert(GoOST.Int(10))
	tree.Insert(GoOST.Int(20))
	tree.Insert(GoOST.Int(30))

	rank := tree.Rank(GoOST.Int(20))
	fmt.Printf("Позиция элемента 20: %d\n", rank)

	tree.Delete(GoOST.Int(20))

	rank = tree.Rank(GoOST.Int(30))
	fmt.Printf("Позиция элемента 30: %d\n", rank)
}
