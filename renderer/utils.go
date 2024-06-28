package renderer

import "sync"

type Stack[T any] struct {
	data []T
	lock sync.RWMutex
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		data: make([]T, 0),
	}
}
func (s *Stack[T]) Push(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = append(s.data, value)
}

// Pop a value from the stack
func (s *Stack[T]) Pop() (T, bool) {
	var v T
	var idx int

	s.lock.Lock()
	defer s.lock.Unlock()

	if v, idx = s.get(); idx >= 0 {
		s.data = s.data[:idx] // Pop
		return v, true
	}

	return v, false
}

func (s *Stack[T]) Depth() int {
	return len(s.data)
}

// Peek the item that would be returned from pop. Note that for LIFO situations the
// list consumer must mutex, otherwise the next Pop may yield a different element
func (s *Stack[T]) Peek() (T, bool) {
	var idx int
	var v T

	s.lock.RLock()
	defer s.lock.RUnlock()

	if v, idx = s.get(); idx >= 0 {
		return v, true
	}

	return v, false

}

func (s *Stack[T]) get() (T, int) {
	idx := -1
	var v T

	if len(s.data) > 0 {
		idx = len(s.data) - 1
		v = s.data[idx]
	}

	return v, idx
}
