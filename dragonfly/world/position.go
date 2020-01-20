package world

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"github.com/go-gl/mathgl/mgl32"
	"math"
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

// Position returns a block position that represents the corner of the chunk, where the X and Z of the chunk
// position are multiplied by 16. The y value of the block position returned is always 0.
func (p ChunkPos) BlockPos() block.Position {
	return block.Position{int(p[0] << 4), 0, int(p[1] << 4)}
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

// chunkPosFromHash returns a chunk position from the hash produced using ChunkPos.Hash. It panics if the
// length of the hash is not 8.
func chunkPosFromHash(hash string) ChunkPos {
	if len(hash) != 8 {
		panic("length of hash must be exactly 8 bytes long")
	}
	return ChunkPos{
		int32(hash[3]) | int32(hash[2])<<8 | int32(hash[1])<<16 | int32(hash[0])<<24,
		int32(hash[7]) | int32(hash[6])<<8 | int32(hash[5])<<16 | int32(hash[4])<<24,
	}
}

// chunkPosFromVec3 returns a chunk position from the Vec3 passed. The coordinates of the chunk position are
// those of the Vec3 divided by 16, then rounded down.
func chunkPosFromVec3(vec3 mgl32.Vec3) ChunkPos {
	return ChunkPos{
		int32(math.Floor(float64(vec3[0]) / 16)),
		int32(math.Floor(float64(vec3[2]) / 16)),
	}
}

// chunkPosFromBlockPos returns a chunk position of the chunk that a block at this position would be in.
func chunkPosFromBlockPos(p block.Position) ChunkPos {
	return ChunkPos{int32(p[0] >> 4), int32(p[2] >> 4)}
}

// Distance returns the distance between two vectors.
func Distance(a, b mgl32.Vec3) float32 {
	return float32(math.Sqrt(
		math.Pow(float64(b[0]-a[0]), 2) +
			math.Pow(float64(b[1]-a[1]), 2) +
			math.Pow(float64(b[2]-a[2]), 2),
	))
}
