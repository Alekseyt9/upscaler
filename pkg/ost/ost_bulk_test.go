package ost

import (
	"math/rand"
	"testing"
)

const (
	elemscount = 1000
	findcount  = 10
)

func prepareTreeAndItems() (*OST, []Item) {
	tree := New()

	numbers := make([]Int, elemscount)
	for i := 0; i < elemscount; i++ {
		numbers[i] = Int(rand.Intn(1000))
		tree.Insert(numbers[i])
	}

	itemsToFind := make([]Item, findcount)
	for i := 0; i < findcount; i++ {
		itemsToFind[i] = numbers[rand.Intn(elemscount)]
	}

	return tree, itemsToFind
}

func BenchmarkRankAndRankArr(b *testing.B) {
	tree, itemsToFind := prepareTreeAndItems()

	b.Run("Rank", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, item := range itemsToFind {
				tree.Rank(item)
			}
		}
	})

	b.Run("RankArr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.RankArr(itemsToFind)
		}
	})
}
