package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sync/atomic"
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
	return ChunkPos{
		int32(hash[3]) | int32(hash[2])<<8 | int32(hash[1])<<16 | int32(hash[0])<<24,
		int32(hash[7]) | int32(hash[6])<<8 | int32(hash[5])<<16 | int32(hash[4])<<24,
	}
}

// ChunkPosFromVec3 returns a chunk position from the Vec3 passed. The coordinates of the chunk position are
// those of the Vec3 divided by 16, then rounded down.
func ChunkPosFromVec3(vec3 mgl32.Vec3) ChunkPos {
	return ChunkPos{
		int32(math.Floor(float64(vec3[0]) / 16)),
		int32(math.Floor(float64(vec3[2]) / 16)),
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

// Vec3 returns a vec3 holding the same coordinates as the block position.
func (p BlockPos) Vec3() mgl32.Vec3 {
	return mgl32.Vec3{float32(p[0]), float32(p[1]), float32(p[2])}
}

// BlockPosFromVec3 returns a block position from the Vec3 passed. The coordinates are all rounded down to the
// nearest full number.
func BlockPosFromVec3(vec3 mgl32.Vec3) BlockPos {
	return BlockPos{
		int(math.Floor(float64(vec3[0]))),
		int(math.Floor(float64(vec3[1]))),
		int(math.Floor(float64(vec3[2]))),
	}
}

// Distance returns the distance between two vectors.
func Distance(a, b mgl32.Vec3) float32 {
	return float32(math.Sqrt(
		math.Pow(float64(b[0]-a[0]), 2) +
			math.Pow(float64(b[1]-a[1]), 2) +
			math.Pow(float64(b[2]-a[2]), 2),
	))
}

// Pos holds the position base of the entity. It implements the entity.Entity interface and creates the base
// of the entity that implements position management.
// Entities must embed this struct to be able to use functions in the entity package.
type Pos struct {
	pos, yaw, pitch atomic.Value
}

// Position returns the current position of the entity. It may be changed as the entity moves or is moved
// around the world.
func (pos *Pos) Position() mgl32.Vec3 {
	v := pos.pos.Load()
	if v == nil {
		return mgl32.Vec3{}
	}
	return v.(mgl32.Vec3)
}

// Yaw returns the yaw of the entity. This is horizontal rotation (rotation around the vertical axis), and
// is 0 when the entity faces forward.
func (pos *Pos) Yaw() float32 {
	v := pos.yaw.Load()
	if v == nil {
		return 0
	}
	return v.(float32)
}

// Pitch returns the pitch of the entity. This is vertical rotation (rotation around the horizontal axis),
// and is 0 when the entity faces forward.
func (pos *Pos) Pitch() float32 {
	v := pos.pitch.Load()
	if v == nil {
		return 0
	}
	return v.(float32)
}

// setYaw sets the yaw of the entity to the new yaw passed. It merely sets the field of the struct and does
// not take care of sending it to viewers.
func (pos *Pos) setYaw(new float32) {
	pos.yaw.Store(new)
}

// setPitch sets the pitch of the entity to the new pitch passed. It merely sets the field of the struct and
// does not take care of sending it to viewers.
func (pos *Pos) setPitch(new float32) {
	pos.pitch.Store(new)
}

// setPosition sets the position of the entity to a new position passed. It merely sets the field of the
// struct and does not take care sending it to viewers.
func (pos *Pos) setPosition(new mgl32.Vec3) {
	pos.pos.Store(new)
}
