package main

import (
	"fmt"

	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/pkg/ost"
)

func main() {
	test2()
}

func test1() {
	tree := ost.New()
	tree.Insert(ost.Int(10))
	tree.Insert(ost.Int(20))
	tree.Insert(ost.Int(30))
	rank := tree.Rank(ost.Int(20))
	fmt.Printf("Позиция элемента 20: %d\n", rank)
	tree.Delete(ost.Int(20))
	rank = tree.Rank(ost.Int(30))
	fmt.Printf("Позиция элемента 30: %d\n", rank)
}

func test2() {
	queue := ost.NewPOST()
	o1 := store.CacheQueueItem{
		ID:    273,
		Order: 273,
	}
	queue.Insert(o1)
	queue.Delete(o1)
	queue.Insert(o1)
}
