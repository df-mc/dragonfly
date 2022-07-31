package world

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// ChunkPos holds the position of a chunk. The type is provided as a utility struct for keeping track of a
// chunk's position. Chunks do not themselves keep track of that. Chunk positions are different from block
// positions in the way that increasing the X/Z by one means increasing the absolute value on the X/Z axis in
// terms of blocks by 16.
type ChunkPos [2]int32

// String implements fmt.Stringer and returns (x, z).
func (p ChunkPos) String() string {
	return fmt.Sprintf("(%v, %v)", p[0], p[1])
}

// X returns the X coordinate of the chunk position.
func (p ChunkPos) X() int32 {
	return p[0]
}

// Z returns the Z coordinate of the chunk position.
func (p ChunkPos) Z() int32 {
	return p[1]
}

// SubChunkPos holds the position of a sub-chunk. The type is provided as a utility struct for keeping track of a
// sub-chunk's position. Sub-chunks do not themselves keep track of that. Sub-chunk positions are different from
// block positions in the way that increasing the X/Y/Z by one means increasing the absolute value on the X/Y/Z axis in
// terms of blocks by 16.
type SubChunkPos [3]int32

// String implements fmt.Stringer and returns (x, y, z).
func (p SubChunkPos) String() string {
	return fmt.Sprintf("(%v, %v, %v)", p[0], p[1], p[2])
}

// X returns the X coordinate of the sub-chunk position.
func (p SubChunkPos) X() int32 {
	return p[0]
}

// Y returns the Y coordinate of the sub-chunk position.
func (p SubChunkPos) Y() int32 {
	return p[1]
}

// Z returns the Z coordinate of the sub-chunk position.
func (p SubChunkPos) Z() int32 {
	return p[2]
}

// blockPosFromNBT returns a position from the X, Y and Z components stored in the NBT data map passed. The
// map is assumed to have an 'x', 'y' and 'z' key.
func blockPosFromNBT(data map[string]any) cube.Pos {
	x, _ := data["x"].(int32)
	y, _ := data["y"].(int32)
	z, _ := data["z"].(int32)
	return cube.Pos{int(x), int(y), int(z)}
}

// chunkPosFromVec3 returns a chunk position from the Vec3 passed. The coordinates of the chunk position are
// those of the Vec3 divided by 16, then rounded down.
func chunkPosFromVec3(vec3 mgl64.Vec3) ChunkPos {
	return ChunkPos{
		int32(math.Floor(vec3[0])) >> 4,
		int32(math.Floor(vec3[2])) >> 4,
	}
}

// chunkPosFromBlockPos returns the ChunkPos of the chunk that a block at a cube.Pos is in.
func chunkPosFromBlockPos(p cube.Pos) ChunkPos {
	return ChunkPos{int32(p[0] >> 4), int32(p[2] >> 4)}
}
