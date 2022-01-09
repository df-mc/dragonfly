package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TODO: Dragon Heads can be powered by redstone

// Skull is a decorative block. There are six types of skulls: player, zombie, skeleton, wither skeleton, creeper,
// and dragon.
type Skull struct {
	transparent

	// Type is the type of the skull.
	Type SkullType
	// Direction is the direction the skull is facing. For skulls placed on the floor, this is cube.FaceUp.
	Direction cube.Face
	// Rotation is the number of rotations for skulls placed on the floor. There are a total of 16 rotations.
	Rotation cube.Orientation
}

// Helmet ...
func (Skull) Helmet() bool {
	return true
}

// DefencePoints ...
func (Skull) DefencePoints() float64 {
	return 0
}

// KnockBackResistance ...
func (Skull) KnockBackResistance() float64 {
	return 0
}

// Model ...
func (s Skull) Model() world.BlockModel {
	return model.Skull{Direction: s.Direction}
}

// UseOnBlock ...
func (s Skull) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used || face == cube.FaceDown {
		return false
	}

	s.Direction = face
	if face == cube.FaceUp {
		yaw, _ := user.Rotation()
		s.Rotation = cube.OrientationFromYaw(yaw)
	}
	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// CanDisplace ...
func (Skull) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Skull) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// HasLiquidDrops ...
func (Skull) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (s Skull) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, nothingEffective, oneOf(Skull{Type: s.Type}))
}

// EncodeItem ...
func (s Skull) EncodeItem() (name string, meta int16) {
	return "minecraft:skull", int16(s.Type.Uint8())
}

// DecodeNBT ...
func (s Skull) DecodeNBT(data map[string]interface{}) interface{} {
	s.Type = SkullType{skull(nbtconv.MapByte(data, "SkullType"))}
	s.Rotation = cube.Orientation(nbtconv.MapByte(data, "Rot"))
	return s
}

// EncodeNBT ...
func (s Skull) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"id": "Skull", "SkullType": s.Type.Uint8(), "Rot": byte(s.Rotation)}
}

// EncodeBlock ...
func (s Skull) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:skull", map[string]interface{}{"facing_direction": int32(s.Direction), "no_drop_bit": uint8(0)}
}

// allSkulls ...
func allSkulls() (skulls []world.Block) {
	for _, f := range cube.Faces() {
		skulls = append(skulls, Skull{Direction: f})
	}
	return
}
