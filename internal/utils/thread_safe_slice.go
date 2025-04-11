package utils

import (
	"sync"
)

type ThreadSafeSlice[T any] struct {
	mu sync.Mutex
	s  []T
}

func (s *ThreadSafeSlice[T]) Append(v T) {
	s.mu.Lock()
	s.s = append(s.s, v)
	s.mu.Unlock()
}

func (s *ThreadSafeSlice[T]) Get() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Возвращаем копию
	return append([]T(nil), s.s...)
}

func (s *ThreadSafeSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.s)
}

func (s *ThreadSafeSlice[T]) ToSlice() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]T(nil), s.s...)
}
