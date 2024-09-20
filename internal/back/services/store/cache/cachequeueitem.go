package cache

import "github.com/Alekseyt9/upscaler/pkg/ost"

// CacheQueueItem represents a task or file item in the queue.
type CacheQueueItem struct {
	ID    int64 // The unique identifier for the queue item.
	Order int64 // The order number that determines the item's position in the queue.
}

// Less compares the current CacheQueueItem with another item.
func (c CacheQueueItem) Less(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order < other.Order
}

// Greater compares the current CacheQueueItem with another item.
func (c CacheQueueItem) Greater(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order > other.Order
}

// Equal checks if the current CacheQueueItem is equal to another item.
func (c CacheQueueItem) Equal(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order == other.Order
}

// Key returns the ID of the CacheQueueItem as an integer.
func (c CacheQueueItem) Key() int {
	return int(c.ID)
}
