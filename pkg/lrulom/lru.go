// Package lrulom provides an LRU cache wrapper with automatic loading of missing values.
package lrulom

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// LRULoadOnMiss is a wrapper around an LRU cache that automatically loads missing values using a provided load function.
type LRULoadOnMiss[K comparable, V any] struct {
	cache    *lru.Cache[K, V]
	loadFunc func(key K) (V, error)
	mu       sync.RWMutex
}

// New creates a new instance of LRULoadOnMiss with the specified size and load function.
//
// size specifies the maximum number of entries in the cache.
// loadFunc is a function that loads a value when it is not found in the cache.
//
// Returns a pointer to an LRULoadOnMiss instance or an error if the LRU cache creation fails.
func New[K comparable, V any](size int, loadFunc func(key K) (V, error)) (*LRULoadOnMiss[K, V], error) {
	cache, err := lru.New[K, V](size)
	if err != nil {
		return nil, err
	}
	return &LRULoadOnMiss[K, V]{
		cache:    cache,
		loadFunc: loadFunc,
	}, nil
}

// GetOrLoad attempts to retrieve a value from the cache. If the value is not present,
// it uses the load function to load the value, stores it in the cache, and returns it.
//
// key is the key of the value to retrieve.
//
// Returns the value associated with the key or an error if loading the value fails.
func (l *LRULoadOnMiss[K, V]) GetOrLoad(key K) (V, error) {
	l.mu.RLock()
	value, ok := l.cache.Get(key)
	l.mu.RUnlock()

	if ok {
		return value, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check if the value was loaded while acquiring the write lock
	if value, ok := l.cache.Get(key); ok {
		return value, nil
	}

	value, err := l.loadFunc(key)
	if err != nil {
		var zeroValue V
		return zeroValue, err
	}

	l.cache.Add(key, value)
	return value, nil
}

// Add adds a value to the cache.
//
// key is the key of the value to add.
// value is the value to store in the cache.
func (l *LRULoadOnMiss[K, V]) Add(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Add(key, value)
}

// Remove removes a value from the cache.
//
// key is the key of the value to remove.
func (l *LRULoadOnMiss[K, V]) Remove(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Remove(key)
}

// Get attempts to retrieve a value from the cache without triggering the load function.
//
// key is the key of the value to retrieve.
//
// Returns the value associated with the key and a boolean indicating whether the value was found in the cache.
func (l *LRULoadOnMiss[K, V]) Get(key K) (V, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.cache.Get(key)
}
