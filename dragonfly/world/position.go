package world

import (
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"unsafe"
)

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

// OutOfBounds checks if the Y value is either bigger than 255 or smaller than 0.
func (p BlockPos) OutOfBounds() bool {
	y := p[1]
	return y > 255 || y < 0
}

// Add adds two block positions together and returns a new one with the combined values.
func (p BlockPos) Add(pos BlockPos) BlockPos {
	return BlockPos{p[0] + pos[0], p[1] + pos[1], p[2] + pos[2]}
}

// Vec3 returns a vec3 holding the same coordinates as the block position.
func (p BlockPos) Vec3() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]), float64(p[1]), float64(p[2])}
}

// Vec3Middle returns a Vec3 holding the coordinates of the block position with 0.5 added on both horizontal
// axes.
func (p BlockPos) Vec3Middle() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]) + 0.5, float64(p[1]), float64(p[2]) + 0.5}
}

// Vec3Centre returns a Vec3 holding the coordinates of the block position with 0.5 added on all axes.
func (p BlockPos) Vec3Centre() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]) + 0.5, float64(p[1]) + 0.5, float64(p[2]) + 0.5}
}

// Side returns the position on the side of this block position, at a specific face.
func (p BlockPos) Side(face Face) BlockPos {
	switch face {
	case FaceUp:
		p[1]++
	case FaceDown:
		p[1]--
	case FaceNorth:
		p[2]--
	case FaceSouth:
		p[2]++
	case FaceWest:
		p[0]--
	case FaceEast:
		p[0]++
	}
	return p
}

// Neighbours calls the function passed for each of the block position's neighbours. If the Y value is below
// 0 or above 255, the function will not be called for that position.
func (p BlockPos) Neighbours(f func(neighbour BlockPos)) {
	y := p[1]
	if y > 255 || y < 0 {
		return
	}
	p[0]++
	f(p)
	p[0] -= 2
	f(p)
	p[0]++
	p[1]++
	if p[1] <= 255 {
		f(p)
	}
	p[1] -= 2
	if p[1] >= 0 {
		f(p)
	}
	p[1]++
	p[2]++
	f(p)
	p[2] -= 2
	f(p)
}

// blockPosFromNBT returns a position from the X, Y and Z components stored in the NBT data map passed. The
// map is assumed to have an 'x', 'y' and 'z' key.
//noinspection GoCommentLeadingSpace
func blockPosFromNBT(data map[string]interface{}) BlockPos {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	xInterface, _ := data["x"]
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	yInterface, _ := data["y"]
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	zInterface, _ := data["z"]
	x, _ := xInterface.(int32)
	y, _ := yInterface.(int32)
	z, _ := zInterface.(int32)
	return BlockPos{int(x), int(y), int(z)}
}

// BlockPosFromVec3 returns a block position by a Vec3, rounding the values down adequately.
func BlockPosFromVec3(vec3 mgl64.Vec3) BlockPos {
	return BlockPos{int(math.Floor(vec3[0])), int(math.Floor(vec3[1])), int(math.Floor(vec3[2]))}
}

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

// chunkPosFromVec3 returns a chunk position from the Vec3 passed. The coordinates of the chunk position are
// those of the Vec3 divided by 16, then rounded down.
func chunkPosFromVec3(vec3 mgl64.Vec3) ChunkPos {
	return ChunkPos{
		int32(math.Floor(vec3[0] / 16)),
		int32(math.Floor(vec3[2] / 16)),
	}
}

// chunkPosFromBlockPos returns a chunk position of the chunk that a block at this position would be in.
func chunkPosFromBlockPos(p BlockPos) ChunkPos {
	return ChunkPos{int32(p[0] >> 4), int32(p[2] >> 4)}
}

// Distance returns the distance between two vectors.
func Distance(a, b mgl64.Vec3) float64 {
	return math.Sqrt(
		math.Pow(b[0]-a[0], 2) +
			math.Pow(b[1]-a[1], 2) +
			math.Pow(b[2]-a[2], 2),
	)
}
