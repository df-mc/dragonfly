package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Bed struct {
	transparent
	sourceWaterDisplacer

	// Colour is the colour of the bed.
	Colour item.Colour
	// Facing is the direction that the bed is facing.
	Facing cube.Direction
	// HeadSide is true if the bed is the head side.
	HeadSide bool
}

// MaxCount always returns 1.
func (Bed) MaxCount() int {
	return 1
}

// Model ...
func (Bed) Model() world.BlockModel {
	return model.Bed{}
}

// SideClosed ...
func (Bed) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (b Bed) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, nothingEffective, oneOf(b))
}

// UseOnBlock ...
func (b Bed) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	if pos, _, used = firstReplaceable(w, pos, face, b); !used {
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
		return
	}

	b.Facing = user.Rotation().Direction()

	side, sidePos := b, pos.Side(b.Facing.Face())
	side.HeadSide = true

	if !replaceableWith(w, sidePos, side) {
		return
	}
	if _, ok := w.Block(sidePos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
		return
	}

	ctx.IgnoreBBox = true
	place(w, sidePos, side, user, ctx)
	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// EntityLand ...
func (b Bed) EntityLand(_ cube.Pos, _ *world.World, e world.Entity, distance *float64) {
	if s, ok := e.(sneakingEntity); ok && s.Sneaking() {
		// If the entity is sneaking, the fall distance and velocity stay the same.
		return
	}
	if _, ok := e.(fallDistanceEntity); ok {
		*distance *= 0.5
	}
	if v, ok := e.(velocityEntity); ok {
		vel := v.Velocity()
		vel[1] = vel[1] * -3 / 4
		v.SetVelocity(vel)
	}
}

// sneakingEntity represents an entity that can sneak.
type sneakingEntity interface {
	// Sneaking returns true if the entity is currently sneaking.
	Sneaking() bool
}

// velocityEntity represents an entity that can maintain a velocity.
type velocityEntity interface {
	// Velocity returns the current velocity of the entity.
	Velocity() mgl64.Vec3
	// SetVelocity sets the velocity of the entity.
	SetVelocity(mgl64.Vec3)
}

// NeighbourUpdateTick ...
func (b Bed) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, _, ok := b.Side(pos, w); !ok {
		w.SetBlock(pos, nil, nil)
	}
}

// EncodeItem ...
func (b Bed) EncodeItem() (name string, meta int16) {
	return "minecraft:bed", int16(b.Colour.Uint8())
}

// EncodeBlock ...
func (b Bed) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:bed", map[string]interface{}{
		"facing_bit":   int32(horizontalDirection(b.Facing)),
		"occupied_bit": boolByte(false),
		"head_bit":     boolByte(b.HeadSide),
	}
}

// EncodeNBT ...
func (b Bed) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{
		"id":    "Bed",
		"color": b.Colour.Uint8(),
	}
}

// DecodeNBT ...
func (b Bed) DecodeNBT(data map[string]interface{}) interface{} {
	b.Colour = item.Colours()[nbtconv.Uint8(data, "color")]
	return b
}

// Head returns the head side of the bed. If neither side is a head side, the third return value is false.
func (b Bed) Head(pos cube.Pos, w *world.World) (Bed, cube.Pos, bool) {
	headSide, headPos, ok := b.Side(pos, w)
	if !ok {
		return Bed{}, cube.Pos{}, false
	}
	if b.HeadSide {
		headSide, headPos = b, pos
	}
	return headSide, headPos, true
}

// Side returns the other side of the bed. If the other side is not a bed, the third return value is false.
func (b Bed) Side(pos cube.Pos, w *world.World) (Bed, cube.Pos, bool) {
	face := b.Facing.Face()
	if b.HeadSide {
		face = face.Opposite()
	}

	sidePos := pos.Side(face)
	o, ok := w.Block(sidePos).(Bed)
	return o, sidePos, ok
}

// allBeds returns all possible beds.
func allBeds() (beds []world.Block) {
	for _, d := range cube.Directions() {
		beds = append(beds, Bed{Facing: d})
		beds = append(beds, Bed{Facing: d, HeadSide: true})
	}
	return
}
