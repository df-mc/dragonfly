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
func (s Skull) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used || face == cube.FaceDown {
		return false
	}

	if face == cube.FaceUp {
		yaw, _ := user.Rotation()
		s.Attach = StandingAttachment(cube.OrientationFromYaw(yaw))
	} else {
		s.Attach = WallAttachment(face.Direction())
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
func (s Skull) DecodeNBT(_ cube.Pos, _ *world.World, data map[string]any) any {
	s.Type = SkullType{skull(nbtconv.Map[byte](data, "SkullType"))}
	s.Attach.o = cube.Orientation(nbtconv.Map[byte](data, "Rot"))
	if s.Attach.facing >= 0 {
		s.Attach.hanging = true
	}
	return s
}

// EncodeNBT ...
func (s Skull) EncodeNBT(cube.Pos, *world.World) map[string]any {
	return map[string]any{"id": "Skull", "SkullType": s.Type.Uint8(), "Rot": byte(s.Attach.o)}
}

// EncodeBlock ...
func (s Skull) EncodeBlock() (string, map[string]any) {
	if s.Attach.hanging {
		if s.Attach.facing == unknownDirection {
			return "minecraft:skull", map[string]any{"facing_direction": int32(0)}
		}
		return "minecraft:skull", map[string]any{"facing_direction": int32(s.Attach.facing) + 2}
	}
	return "minecraft:skull", map[string]any{"facing_direction": int32(1)}
}

// allSkulls ...
func allSkulls() (skulls []world.Block) {
	for _, f := range cube.HorizontalFaces() {
		// A direction of -2 and -1 isn't actually valid, but when encoding the block these are encoded as 0 and 1. We
		// can't otherwise represent this properly in an Attachment type.
		skulls = append(skulls, Skull{Attach: WallAttachment(f.Direction())})
	}
	return append(skulls, Skull{Attach: StandingAttachment(0)})
}
