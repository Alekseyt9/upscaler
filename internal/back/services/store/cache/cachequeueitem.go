package cache

import "github.com/Alekseyt9/upscaler/pkg/ost"

type CacheQueueItem struct {
	ID    int64
	Order int64
}

func (c CacheQueueItem) Less(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order < other.Order
}

func (c CacheQueueItem) Greater(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order > other.Order
}

func (c CacheQueueItem) Equal(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order == other.Order
}

func (c CacheQueueItem) Key() int {
	return int(c.ID)
}
