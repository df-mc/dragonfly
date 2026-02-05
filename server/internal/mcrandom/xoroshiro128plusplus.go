package mcrandom

import "math/bits"

// Xoroshiro128PlusPlus is a member of the Xor-Shift-Rotate family of generators. Memory
// footprint is 128 bits and the period is (2^128)-1.
type Xoroshiro128PlusPlus struct {
	seed0, seed1 uint64
}

// NewXoroshiro128PlusPlus ...
func NewXoroshiro128PlusPlus(seed0, seed1 uint64) *Xoroshiro128PlusPlus {
	return &Xoroshiro128PlusPlus{seed0, seed1}
}

// Next ...
func (x *Xoroshiro128PlusPlus) Next() uint64 {
	s0 := x.seed0
	s1 := x.seed1
	result := bits.RotateLeft64(s0+s1, 17) + s0
	s1 ^= s0
	x.seed0 = bits.RotateLeft64(s0, 49) ^ s1 ^ (s1 << 21)
	x.seed1 = bits.RotateLeft64(s1, 28)
	return result
}
