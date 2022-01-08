package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// blockPosFromNBT returns a position from the X, Y and Z components stored in the NBT data map passed. The
// map is assumed to have an 'x', 'y' and 'z' key.
//noinspection GoCommentLeadingSpace
func blockPosFromNBT(data map[string]interface{}) cube.Pos {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	xInterface, _ := data["x"]
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	yInterface, _ := data["y"]
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	zInterface, _ := data["z"]
	x, _ := xInterface.(int32)
	y, _ := yInterface.(int32)
	z, _ := zInterface.(int32)
	return cube.Pos{int(x), int(y), int(z)}
}

// ChunkPos holds the position of a chunk. The type is provided as a utility struct for keeping track of a
// chunk's position. Chunks do not themselves keep track of that. Chunk positions are different from block
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

// chunkPosFromVec3 returns a chunk position from the Vec3 passed. The coordinates of the chunk position are
// those of the Vec3 divided by 16, then rounded down.
func chunkPosFromVec3(vec3 mgl64.Vec3) ChunkPos {
	return ChunkPos{
		int32(math.Floor(vec3[0])) >> 4,
		int32(math.Floor(vec3[2])) >> 4,
	}
}

// chunkPosFromBlockPos returns a chunk position of the chunk that a block at this position would be in.
func chunkPosFromBlockPos(p cube.Pos) ChunkPos {
	return ChunkPos{int32(p[0] >> 4), int32(p[2] >> 4)}
}

// Distance returns the distance between two vectors.
func Distance(a, b mgl64.Vec3) float64 {
	return b.Sub(a).Len()
}
