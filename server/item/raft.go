package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Raft is an item used to sail across water. Bamboo wood uses a raft instead of a boat.
// Unlike regular wood boats, bamboo rafts have a flat design.
type Raft struct {
	// Chest specifies whether the raft has a chest in it.
	Chest bool
}

// UseOnBlock spawns a raft entity in the world at the adjacent block.
func (r Raft) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().Raft
	if create == nil {
		return false
	}

	spawnPos := pos.Side(face)
	rot := user.Rotation()
	rot[1] = 0

	opts := world.EntitySpawnOpts{
		Position: spawnPos.Vec3Centre(),
		Rotation: rot,
	}
	tx.AddEntity(create(opts, r.Chest))

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (r Raft) EncodeItem() (name string, meta int16) {
	if r.Chest {
		return "minecraft:bamboo_chest_raft", 0
	}
	return "minecraft:bamboo_raft", 0
}

// MaxCount ...
func (Raft) MaxCount() int {
	return 1
}
