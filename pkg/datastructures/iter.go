package datastructures

import "context"

type Iterable[T any] interface {
	Iter(ctx context.Context) <-chan T
}

// Run f on every element, changing it in-place.
func Map[T any](l []T, f func(el T) T) {
	for i, el := range l {
		l[i] = f(el)
	}
}

// Run f on every element and store the result in a new list.
func MapCopy[T any](l []T, f func(el T) T) []T {
	newList := make([]T, len(l))

	for i, el := range l {
		newList[i] = f(el)
	}

	return newList
}

// Run f on every element, expecting a list as return value.
func Expand[T any](l []T, f func(el T) []T) []T {
	expanded := make([]T, 0, len(l))

	for _, el := range l {
		expanded = append(expanded, f(el)...)
	}

	return expanded
}

// Reduce a list to a single value.
// Using a function taking the current element and an accumulator, returning the new valie for the accumulator.
// After the last element is processed, the accumulator is returned.
func Reduce[T any, A any](l []T, f func(el T, acc A) A, initial A) A {
	acc := initial
	for _, el := range l {
		acc = f(el, acc)
	}

	return acc
}

// Returns true if any of the elements match the predicate.
func Any[T any](l []T, f func(el T) bool) bool {
	for _, el := range l {
		if f(el) {
			return true
		}
	}

	return false
}

// Returns true if all of the elements match the predicate.
func All[T any](l []T, f func(el T) bool) bool {
	for _, el := range l {
		if !f(el) {
			return false
		}
	}

	return true
}

// Returns the first instance that passes the given predicate.
func Find[T any](l []T, f func(el T) bool) *T {
	for _, el := range l {
		if f(el) {
			return &el
		}
	}

	return nil
}

// Returns a copy with instances that do not pass the filter function removed.
func Filter[T any](l []T, f func(el T) bool) []T {
	c := make([]T, 0, cap(l))
	for _, el := range l {
		if f(el) {
			c = append(c, el)
		}
	}

	return c
}
