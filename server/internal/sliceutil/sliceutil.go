package sliceutil

// Convert converts a slice of type B to a slice of type A. Convert panics if B cannot be type asserted to type A.
func Convert[A, B any, S ~[]B](v S) []A {
	a := make([]A, len(v))
	for i, b := range v {
		a[i] = (interface{})(b).(A)
	}
	return a
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present. Index accepts any type, as opposed to
// slices.Index, but might panic if E is not comparable.
func Index[E any](s []E, v E) int {
	for i, vs := range s {
		if (interface{})(v) == (interface{})(vs) {
			return i
		}
	}
	return -1
}
