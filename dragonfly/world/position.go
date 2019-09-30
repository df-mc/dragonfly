package world

import (
	"unsafe"
)

// ChunkPos holds the position of a chunk. The type is provided as a utility struct for keeping track of a
// chunk's position. Chunks do not themselves keep track of that. Chunk positions are different than block
// positions in the way that increasing the X/Z by one means increasing the absolute value on the X/Z axis in
// terms of blocks by 16.
type ChunkPos [2]int32

// X returns the X coordinate of the chunk position.
func (p ChunkPos) X() int32 {
	return p[0]
}

// Z returns the Z coordinate of the chunk position.
func (p ChunkPos) Z() int32 {
	return p[1]
}

// BlockPos returns a block position that represents the corner of the chunk, where the X and Z of the chunk
// position are multiplied by 16. The y value of the block position returned is always 0.
func (p ChunkPos) BlockPos() BlockPos {
	return BlockPos{int(p[0] << 4), 0, int(p[1] << 4)}
}

// Hash returns the hash of the chunk position. It is essentially the bytes of the X and Z values of the
// position following each other.
func (p ChunkPos) Hash() string {
	x, z := p[0], p[1]
	v := []byte{
		uint8(x >> 24), uint8(x >> 16), uint8(x >> 8), uint8(x),
		uint8(z >> 24), uint8(z >> 16), uint8(z >> 8), uint8(z),
	}
	// We can 'safely' unsafely turn the byte slice into a string here, as the byte slice will never be
	// changed. (It never leaves the method.)
	return *(*string)(unsafe.Pointer(&v))
}

// ChunkPosFromHash returns a chunk position from the hash produced using ChunkPos.Hash. It panics if the
// length of the hash is not 8.
func ChunkPosFromHash(hash string) ChunkPos {
	if len(hash) != 8 {
		panic("length of hash must be exactly 8 bytes long")
	}
	b := *(*[]byte)(unsafe.Pointer(&hash))
	return ChunkPos{
		int32(b[3]) | int32(b[2])<<8 | int32(b[1])<<16 | int32(b[0])<<24,
		int32(b[7]) | int32(b[6])<<8 | int32(b[5])<<16 | int32(b[4])<<24,
	}
}

// BlockPos holds the position of a block. The position is represented of an array with an x, y and z value,
// where the y value is positive.
type BlockPos [3]int

// X returns the X coordinate of the block position.
func (p BlockPos) X() int {
	return p[0]
}

// Y returns the Y coordinate of the block position.
func (p BlockPos) Y() int {
	return p[1]
}

// Z returns the Z coordinate of the block position.
func (p BlockPos) Z() int {
	return p[2]
}

// ChunkPos returns a chunk position of the chunk that a block at this position would be in.
func (p BlockPos) ChunkPos() ChunkPos {
	return ChunkPos{int32(p[0] >> 4), int32(p[2] >> 4)}
}
