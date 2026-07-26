package mcrandom

// MixStafford13 implements the Stafford 13 mixing function, a bijective mixing function
// suitable for use in pseudorandom number generators.
func MixStafford13(seed uint64) uint64 {
	seed = (seed ^ (seed >> 30)) * 0xBF58476D1CE4E5B9
	seed = (seed ^ (seed >> 27)) * 0x94D049BB133111EB
	return seed ^ (seed >> 31)
}
