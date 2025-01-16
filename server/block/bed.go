package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bed is a block, allowing players to sleep to set their spawns and skip the night.
type Bed struct {
	transparent
	sourceWaterDisplacer

	// Colour is the colour of the bed.
	Colour item.Colour
	// Facing is the direction that the bed is Facing.
	Facing cube.Direction
	// Head is true if the bed is the head side.
	Head bool
	// Sleeper is the user that is using the bed. It is only set for the Head part of the bed.
	Sleeper *world.EntityHandle
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
func (Bed) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (b Bed) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, nothingEffective, oneOf(b)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		headSide, _, ok := b.head(pos, tx)
		if !ok {
			return
		}

		sleeper := headSide.Sleeper
		if sleeper != nil {

			ent, ok := sleeper.Entity(tx)
			if ok {
				sleeper, ok := ent.(world.Sleeper)
				if ok {
					sleeper.Wake()
				}
			}
		}
	})
}

// UseOnBlock ...
func (b Bed) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	if pos, _, used = firstReplaceable(tx, pos, face, b); !used {
		return
	}
	if !supportedFromBelow(pos, tx) {
		return
	}

	b.Facing = user.Rotation().Direction()

	side, sidePos := b, pos.Side(b.Facing.Face())
	side.Head = true

	if !replaceableWith(tx, sidePos, side) {
		return
	}

	if !supportedFromBelow(sidePos, tx) {
		return
	}

	ctx.IgnoreBBox = true
	place(tx, sidePos, side, user, ctx)
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Activate ...
func (b Bed) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	s, ok := u.(world.Sleeper)
	if !ok {
		return false
	}

	w := tx.World()

	if w.Dimension() != world.Overworld {
		tx.SetBlock(pos, nil, nil)
		ExplosionConfig{
			Size:      5,
			SpawnFire: true,
		}.Explode(tx, pos.Vec3Centre())
		return true
	}

	_, sidePos, ok := b.side(pos, tx)
	if !ok {
		return false
	}

	userPos := s.Position()
	if sidePos.Vec3Middle().Sub(userPos).Len() > 4 && pos.Vec3Middle().Sub(userPos).Len() > 4 {
		s.Messaget(chat.MessageBedTooFar)
		return true
	}

	headSide, headPos, ok := b.head(pos, tx)
	if !ok {
		return false
	}
	if _, ok = tx.Liquid(headPos); ok {
		return false
	}

	previousSpawn := w.PlayerSpawn(s.UUID())
	if previousSpawn != pos && previousSpawn != sidePos {
		w.SetPlayerSpawn(s.UUID(), pos)
		s.Messaget(chat.MessageRespawnPointSet)
	}

	time := w.Time() % world.TimeFull
	if (time < world.TimeNight || time >= world.TimeSunrise) && !tx.ThunderingAt(pos) {
		s.Messaget(chat.MessageNoSleep)
		return true
	}
	if headSide.Sleeper != nil {
		s.Messaget(chat.MessageBedIsOccupied)
		return true
	}

	s.Sleep(headPos)
	return true
}

// EntityLand ...
func (b Bed) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance *= 0.5
	}
	if v, ok := e.(velocityEntity); ok {
		vel := v.Velocity()
		vel[1] = vel[1] * -2 / 3
		v.SetVelocity(vel)
	}
}

// velocityEntity represents an entity that can maintain a velocity.
type velocityEntity interface {
	// Velocity returns the current velocity of the entity.
	Velocity() mgl64.Vec3
	// SetVelocity sets the velocity of the entity.
	SetVelocity(mgl64.Vec3)
}

// NeighbourUpdateTick ...
func (b Bed) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, _, ok := b.side(pos, tx); !ok {
		breakBlockNoDrops(b, pos, tx)
	}
}

// EncodeItem ...
func (b Bed) EncodeItem() (name string, meta int16) {
	return "minecraft:bed", int16(b.Colour.Uint8())
}

// EncodeBlock ...
func (b Bed) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:bed", map[string]interface{}{
		"direction":    int32(horizontalDirection(b.Facing)),
		"occupied_bit": boolByte(b.Sleeper != nil),
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
	b.Colour = item.Colours()[nbtconv.Uint8(data, "color")]
	return b
}

// head returns the head side of the bed. If neither side is a head side, the third return value is false.
func (b Bed) head(pos cube.Pos, tx *world.Tx) (Bed, cube.Pos, bool) {
	headSide, headPos, ok := b.side(pos, tx)
	if !ok {
		return Bed{}, cube.Pos{}, false
	}
	if b.Head {
		return b, pos, true
	}
	return headSide, headPos, true
}

// side returns the other side of the bed. If the other side is not a bed, the third return value is false.
func (b Bed) side(pos cube.Pos, tx *world.Tx) (Bed, cube.Pos, bool) {
	face := b.Facing.Face()
	if b.Head {
		face = face.Opposite()
	}

	sidePos := pos.Side(face)
	o, ok := tx.Block(sidePos).(Bed)
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

func (Bed) CanRespawnOn() bool {
	return true
}

func (Bed) RespawnOn(pos cube.Pos, u item.User, tx *world.Tx) {}

// RespawnBlock represents a block using which player can set his spawn point.
type RespawnBlock interface {
	// CanRespawnOn defines if player can use this block to respawn.
	CanRespawnOn() bool
	// RespawnOn is called when a player decides to respawn using this block.
	RespawnOn(pos cube.Pos, u item.User, tx *world.Tx)
}

// supportedFromBelow ...
func supportedFromBelow(pos cube.Pos, tx *world.Tx) bool {
	below := pos.Side(cube.FaceDown)
	return tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
}
