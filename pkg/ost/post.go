package ost

import (
	"sync"
)

// POST is a thread-safe wrapper for the OST structure.
type POST struct {
	ost *OST
	mu  sync.RWMutex
}

// NewPOST creates a new POST instance.
func NewPOST() *POST {
	return &POST{
		ost: New(),
	}
}

// Insert inserts an item into the OST in a thread-safe manner.
func (p *POST) Insert(item Item) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ost.Insert(item)
}

// Delete removes an item from the OST in a thread-safe manner.
func (p *POST) Delete(item Item) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ost.Delete(item)
}

// Rank returns the rank of an item in the OST in a thread-safe manner.
func (p *POST) Rank(item Item) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ost.Rank(item)
}

func (n *Node) Count() int {
	if n == nil {
		return 0
	}
	return n.count
}
