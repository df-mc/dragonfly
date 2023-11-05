package sliceutil

import "golang.org/x/exp/slices"

// Convert converts a slice of type B to a slice of type A. Convert panics if B
// cannot be type asserted to type A.
func Convert[A, B any, S ~[]B](v S) []A {
	a := make([]A, len(v))
	for i, b := range v {
		a[i] = (any)(b).(A)
	}
	return a
}

// Index returns the index of the first occurrence of v in s, or -1 if not
// present. Index accepts any type, as opposed to slices.Index, but might panic
// if E is not comparable.
func Index[E any](s []E, v E) int {
	for i, vs := range s {
		if (any)(v) == (any)(vs) {
			return i
		}
	}
	return -1
}

// SearchValue iterates through slice v, calling function f for every element.
// If true is returned in this function, the respective element is returned and
// ok is true. If the function f does not return true for any element, false is
// returned.
func SearchValue[A any, S ~[]A](v S, f func(a A) bool) (a A, ok bool) {
	for _, val := range v {
		if f(val) {
			return val, true
		}
	}
	return
}

// Filter iterates over elements of collection, returning an array of all
// elements function c returns true for.
func Filter[E any](s []E, c func(E) bool) []E {
	a := make([]E, 0, len(s))
	for _, e := range s {
		if c(e) {
			a = append(a, e)
		}
	}
	return a
}

// DeleteVal deletes the first occurrence of a value in a slice of the type E
// and returns a new slice without the value.
func DeleteVal[E any](s []E, v E) []E {
	if i := Index(s, v); i != -1 {
		return slices.Clone(slices.Delete(s, i, i+1))
	}
	return s
}
