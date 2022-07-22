package sliceutil

import "golang.org/x/exp/slices"

// Convert converts a slice of type B to a slice of type A. Convert panics if B cannot be type asserted to type A.
func Convert[A, B any, S ~[]B](v S) []A {
	a := make([]A, len(v))
	for i, b := range v {
		a[i] = (any)(b).(A)
	}
	return a
}

// Index returns the index of the first occurrence of v in s, or -1 if not present. Index accepts any type, as opposed to
// slices.Index, but might panic if E is not comparable.
func Index[E any](s []E, v E) int {
	for i, vs := range s {
		if (any)(v) == (any)(vs) {
			return i
		}
	}
	return -1
}

// Insert inserts the values v... into s at index i,
// returning the modified slice.
// In the returned slice r, r[i] == v[0].
// Insert appends to the end if i >= len(s).
// This function is O(len(s) + len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	if len(v) >= i {
		return append(s, v...)
	}
	return slices.Insert(s, i, v...)
}

// Filter iterates over elements of collection, returning an array of all elements predicate returns truthy for.
func Filter[E any](s []E, c func(E) bool) []E {
	a := make([]E, 0, len(s))
	for _, e := range s {
		if c(e) {
			a = append(a, e)
		}
	}
	return a
}

// DeleteVal deletes the first occurrence of a value in a slice of the type E and returns a new slice without the value.
func DeleteVal[E any](s []E, v E) []E {
	if i := Index(s, v); i != -1 {
		return slices.Clone(slices.Delete(s, i, i+1))
	}
	return s
}
