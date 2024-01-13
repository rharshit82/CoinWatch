package main

import "sync"

type SafeMap struct {
	mu sync.RWMutex
	data  map[currency]string
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[currency]string),
	}
}

func (m *SafeMap) Set(key currency, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
}

func (m *SafeMap) Get(key currency) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.data[key]
	return val, ok
}
