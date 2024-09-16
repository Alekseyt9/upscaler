package ost

import "sort"

func (t *OST) RankArr(items []Item) []int {
	type itemWithIndex struct {
		item  Item
		index int
	}
	itemsWithIndex := make([]itemWithIndex, len(items))
	for i, item := range items {
		itemsWithIndex[i] = itemWithIndex{item: item, index: i}
	}
	sort.Slice(itemsWithIndex, func(i, j int) bool {
		return itemsWithIndex[i].item.Key() < itemsWithIndex[j].item.Key()
	})

	ranks := make([]int, len(items))

	var traverse func(node *Node, left, right int, currentRank int) int
	traverse = func(node *Node, left, right int, currentRank int) int {
		if node == nil || left > right {
			return currentRank
		}

		nodeKey := node.firstItem().Key()

		if itemsWithIndex[right].item.Key() < nodeKey {
			return traverse(node.Left, left, right, currentRank)
		}

		if itemsWithIndex[left].item.Key() > nodeKey {
			leftCount := 0
			if node.Left != nil {
				leftCount = node.Left.count
			}
			currentRank += leftCount + len(node.Items)
			return traverse(node.Right, left, right, currentRank)
		}

		currentRank = traverse(node.Left, left, right, currentRank)

		for i := 0; i < len(node.Items); i++ {
			currentRank++
		}

		start := left
		for start <= right && itemsWithIndex[start].item.Key() < nodeKey {
			start++
		}
		end := start
		for end <= right && itemsWithIndex[end].item.Key() == nodeKey {
			ranks[itemsWithIndex[end].index] = currentRank - (len(node.Items) - (end - start))
			end++
		}

		return traverse(node.Right, end, right, currentRank)
	}

	traverse(t.root, 0, len(itemsWithIndex)-1, 0)

	return ranks
}
