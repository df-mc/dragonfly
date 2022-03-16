package sliceutil

// Convert converts a slice of type A to a slice of type B. Convert panics if A cannot be type asserted to type B.
func Convert[A any, B any, S ~[]A](v S) []B {
	b := make([]B, len(v))
	for i, a := range v {
		b[i] = (interface{})(a).(B)
	}
	return b
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
