package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Bed is a block, allowing players to sleep to set their spawns and skip the night.
type Bed struct {
	transparent

	// Colour is the colour of the bed.
	Colour item.Colour
	// Facing is the direction that the bed is facing.
	Facing cube.Direction
	// Head is true if the bed is the head side or the bottom side.
	Head bool
	// User is the user that is using the bed.
	User item.User
}

// Model ...
func (Bed) Model() world.BlockModel {
	return model.Bed{}
}

// BreakInfo ...
func (b Bed) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, axeEffective, oneOf(b))
}

// UseOnBlock ...
func (b Bed) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	if pos, _, used = firstReplaceable(w, pos, face, b); !used {
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
		return
	}

	b.Facing = user.Facing()

	side, sidePos := b, pos.Side(b.Facing.Face())
	side.Head = true

	if !replaceableWith(w, sidePos, side) {
		return
	}
	if _, ok := w.Block(sidePos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
		return
	}

	place(w, sidePos, side, user, ctx)
	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// sleeper represents an entity that can sleep on beds.
type sleeper interface {
	item.User
	UUID() uuid.UUID
	Message(a ...any)
	Sleep(pos cube.Pos)
	Wake()
}

// Activate ...
func (b Bed) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	s, ok := u.(sleeper)
	if !ok {
		return false
	}

	if w.Dimension() != world.Overworld {
		w.SetBlock(pos, nil, nil)
		ExplosionConfig{
			Size:      5,
			SpawnFire: true,
		}.Explode(w, pos.Vec3Centre())
		return true
	}

	_, sidePos, ok := b.side(pos, w)
	if !ok {
		return false
	}

	userPos := s.Position()
	if sidePos.Vec3Middle().Sub(userPos).Len() > 4 && pos.Vec3Middle().Sub(userPos).Len() > 4 {
		s.Message(text.Colourf("<grey>Bed is too far away</grey>"))
		return true
	}

	headSide, headPos, ok := b.head(pos, w)
	if !ok {
		return false
	}

	w.SetPlayerSpawn(s.UUID(), headPos)
	s.Message(text.Colourf("<grey>Respawn point set</grey>"))

	time := w.Time() % world.TimeFull
	if !(time >= world.TimeNight && time < world.TimeSunrise) {
		s.Message(text.Colourf("<grey>You can only sleep at night</grey>"))
		return true
	}
	if headSide.User != nil {
		s.Message(text.Colourf("<grey>This bed is occupied</grey>"))
		return true
	}

	s.Sleep(headPos)
	return true
}

// PostBreak ...
func (b Bed) PostBreak(pos cube.Pos, w *world.World, _ item.User) {
	headSide, _, ok := b.head(pos, w)
	if !ok {
		return
	}
	if s, ok := headSide.User.(sleeper); ok {
		s.Wake()
	}
}

// EntityLand ...
func (b Bed) EntityLand(_ cube.Pos, _ *world.World, e world.Entity) {
	if s, ok := e.(sneakingEntity); ok && s.Sneaking() {
		// If the entity is sneaking, the fall distance and velocity stay the same.
		return
	}
	if f, ok := e.(fallDistanceEntity); ok {
		f.SetFallDistance(f.FallDistance() * 0.5)
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
	if _, _, ok := b.side(pos, w); !ok {
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
		"occupied_bit": boolByte(b.User != nil),
		"head_bit":     boolByte(b.Head),
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
	b.Colour = item.Colours()[nbtconv.Map[uint8](data, "color")]
	return b
}

// head returns the head side of the bed. If neither side is a head side, the third return value is false.
func (b Bed) head(pos cube.Pos, w *world.World) (Bed, cube.Pos, bool) {
	headSide, headPos, ok := b.side(pos, w)
	if !ok {
		return Bed{}, cube.Pos{}, false
	}
	if b.Head {
		headSide, headPos = b, pos
	}
	return headSide, headPos, true
}

// side returns the other side of the bed. If the other side is not a bed, the third return value is false.
func (b Bed) side(pos cube.Pos, w *world.World) (Bed, cube.Pos, bool) {
	face := b.Facing.Face()
	if b.Head {
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
		beds = append(beds, Bed{Facing: d, Head: true})
	}
	return
}
