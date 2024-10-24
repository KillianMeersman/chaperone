package datastructures

import "context"

type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable](elements ...T) *Set[T] {
	set := &Set[T]{
		elements: make(map[T]struct{}),
	}

	for _, el := range elements {
		set.Add(el)
	}
	return set
}

func (s *Set[T]) Add(el T) {
	s.elements[el] = struct{}{}
}

func (s *Set[T]) Remove(el T) {
	delete(s.elements, el)
}

func (s *Set[T]) Contains(el T) bool {
	_, ok := s.elements[el]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.elements)
}

func (s *Set[T]) Iter(ctx context.Context) <-chan T {
	c := make(chan T)
	go func() {
	outer:
		for el := range s.elements {
			select {
			case c <- el:
			case <-ctx.Done():
				break outer
			}

		}
		close(c)
	}()
	return c
}

func (s *Set[T]) List() []T {
	list := make([]T, s.Len())

	i := 0
	for e := range s.elements {
		list[i] = e
		i++
	}
	return list
}

// Returns the elements that are in both sets.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	intersection := NewSet[T]()

	for el := range s.elements {
		if other.Contains(el) {
			intersection.Add(el)
		}
	}

	return intersection
}

// Returns the elements contained in the set that are not in the other.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	difference := NewSet[T]()

	for el := range s.elements {
		if !other.Contains(el) {
			difference.Add(el)
		}
	}

	return difference
}

// Returns the union of both sets. e.a. all elements of both sets.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	union := NewSet[T]()

	for el := range s.elements {
		union.Add(el)
	}

	for el := range other.elements {
		union.Add(el)
	}

	return union
}
