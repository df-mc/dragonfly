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

// Skull is a decorative block. There are seven types of skulls: player, zombie, skeleton, wither skeleton, creeper,
// dragon, and piglin.
type Skull struct {
	transparent
	sourceWaterDisplacer

	// Type is the type of the skull.
	Type SkullType

	// Attach is the attachment of the Skull. It is either of the type WallAttachment or StandingAttachment.
	//blockhash:facing_only
	Attach Attachment
}

// Helmet ...
func (Skull) Helmet() bool {
	return true
}

// DefencePoints ...
func (Skull) DefencePoints() float64 {
	return 0
}

// Toughness ...
func (Skull) Toughness() float64 {
	return 0
}

// KnockBackResistance ...
func (Skull) KnockBackResistance() float64 {
	return 0
}

// Model ...
func (s Skull) Model() world.BlockModel {
	return model.Skull{Direction: s.Attach.facing.Face(), Hanging: s.Attach.hanging}
}

// UseOnBlock ...
func (s Skull) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, s)
	if !used || face == cube.FaceDown {
		return false
	}

	if face == cube.FaceUp {
		s.Attach = StandingAttachment(user.Rotation().Orientation())
	} else {
		s.Attach = WallAttachment(face.Direction())
	}
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// SideClosed ...
func (Skull) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
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
	return "minecraft:" + s.Type.String(), 0
}

// DecodeNBT ...
func (s Skull) DecodeNBT(data map[string]interface{}) interface{} {
	if t := skull(nbtconv.Uint8(data, "SkullType")); t != 255 {
		// Used to upgrade pre-1.21.40 skulls after their flattening. Any skull placed since will set
		// SkullType to 255.
		s.Type = SkullType{t}
	}
	s.Attach.o = cube.OrientationFromYaw(float64(nbtconv.Float32(data, "Rotation")))
	if s.Attach.facing >= 0 {
		s.Attach.hanging = true
	}
	return s
}

// EncodeNBT ...
func (s Skull) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"id": "Skull", "SkullType": uint8(255), "Rotation": float32(s.Attach.o.Yaw())}
}

// EncodeBlock ...
func (s Skull) EncodeBlock() (string, map[string]interface{}) {
	if s.Attach.hanging {
		if s.Attach.facing == unknownDirection {
			return "minecraft:" + s.Type.String(), map[string]interface{}{"facing_direction": int32(0)}
		}
		return "minecraft:" + s.Type.String(), map[string]interface{}{"facing_direction": int32(s.Attach.facing) + 2}
	}
	return "minecraft:" + s.Type.String(), map[string]interface{}{"facing_direction": int32(1)}
}

// allSkulls ...
func allSkulls() (skulls []world.Block) {
	for _, t := range SkullTypes() {
		for _, d := range append(cube.Directions(), unknownDirection) {
			skulls = append(skulls, Skull{Type: t, Attach: WallAttachment(d)})
		}
		skulls = append(skulls, Skull{Type: t, Attach: StandingAttachment(0)})
	}
	return
}
