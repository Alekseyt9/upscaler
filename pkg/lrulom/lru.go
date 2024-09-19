package lrulom

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// LRULoadOnMiss обертка для LRU-кэша с автоматической подгрузкой
type LRULoadOnMiss[K comparable, V any] struct {
	cache    *lru.Cache[K, V]
	loadFunc func(key K) (V, error)
	mu       sync.RWMutex
}

// New создает новый экземпляр LRULoadOnMiss с заданным размером и функцией загрузки
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

// GetOrLoad пытается получить значение из кэша или загружает его, если значение отсутствует
func (l *LRULoadOnMiss[K, V]) GetOrLoad(key K) (V, error) {
	l.mu.RLock()
	value, ok := l.cache.Get(key)
	l.mu.RUnlock()

	if ok {
		return value, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

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

// Add добавляет значение в кэш
func (l *LRULoadOnMiss[K, V]) Add(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Add(key, value)
}

// Remove удаляет значение из кэша
func (l *LRULoadOnMiss[K, V]) Remove(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Remove(key)
}

// Get пытается получить значение из кэша
func (l *LRULoadOnMiss[K, V]) Get(key K) (V, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.cache.Get(key)
}
