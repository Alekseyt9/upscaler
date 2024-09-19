package lrulom

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

// LRULoadOnMiss обертка для LRU-кэша с автоматической подгрузкой
type LRULoadOnMiss[K comparable, V any] struct {
	Cache    *lru.Cache[K, V]
	loadFunc func(key K) (V, error) // Функция загрузки, задается при создании
}

// New создает новый экземпляр LRULoadOnMiss с заданным размером и функцией загрузки
func New[K comparable, V any](size int, loadFunc func(key K) (V, error)) (*LRULoadOnMiss[K, V], error) {
	cache, err := lru.New[K, V](size)
	if err != nil {
		return nil, err
	}
	return &LRULoadOnMiss[K, V]{
		Cache:    cache,
		loadFunc: loadFunc,
	}, nil
}

// GetOrLoad пытается получить значение из кэша или загружает его, если значение отсутствует
func (l *LRULoadOnMiss[K, V]) GetOrLoad(key K) (V, error) {
	if value, ok := l.Cache.Get(key); ok {
		return value, nil
	}

	value, err := l.loadFunc(key)
	if err != nil {
		var zeroValue V
		return zeroValue, err
	}

	l.Cache.Add(key, value)
	return value, nil
}
