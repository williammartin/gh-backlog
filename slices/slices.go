package slices

import (
	"slices"
)

// Map iterates the slice of elements applying the provided function
// and collecting the results of that application in a new slice.
func Map[A any, B any](a []A, f func(A) B) []B {
	n := make([]B, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

func Sort[T any](slice []T, cmp func(i, j T) int) []T {
	n := make([]T, len(slice))
	copy(n, slice)
	slices.SortFunc(n, cmp)
	return n
}
