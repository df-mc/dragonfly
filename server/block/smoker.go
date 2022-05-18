package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Smoker is a type of furnace that cooks food items, similar to a furnace, but twice as fast. It also serves as a
// butcher's job site block.
type Smoker struct {
	solid
	bassDrum
	*smelter

	// Facing is the direction the smoker is facing.
	Facing cube.Face
	// Lit is true if the smoker is lit.
	Lit bool
}

// NewSmoker creates a new initialised smoker. The smelter is properly initialised.
func NewSmoker(face cube.Face) Smoker {
	return Smoker{
		Facing:  face,
		smelter: newSmelter(),
	}
}

// Tick is called to check if the smoker should update and start or stop smelting.
func (s Smoker) Tick(_ int64, pos cube.Pos, w *world.World) {
	if s.Lit && rand.Float64() <= 0.016 { // Every three or so seconds.
		w.PlaySound(pos.Vec3Centre(), sound.SmokerCrackle{})
	}
	if lit := s.smelter.tickSmelting(time.Second*5, time.Millisecond*200, s.Lit, func(i item.SmeltInfo) bool {
		return i.Food
	}); s.Lit != lit {
		s.Lit = lit
		w.SetBlock(pos, s, nil)
	}
}

// EncodeItem ...
func (s Smoker) EncodeItem() (name string, meta int16) {
	return "minecraft:smoker", 0
}

// EncodeBlock ...
func (s Smoker) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Lit {
		return "minecraft:lit_smoker", map[string]interface{}{"facing_direction": int32(s.Facing)}
	}
	return "minecraft:smoker", map[string]interface{}{"facing_direction": int32(s.Facing)}
}

// UseOnBlock ...
func (s Smoker) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, s)
	if !used {
		return false
	}

	place(w, pos, NewSmoker(user.Facing().Face().Opposite()), user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s Smoker) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3.5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// Activate ...
func (s Smoker) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// EncodeNBT ...
func (s Smoker) EncodeNBT() map[string]interface{} {
	if s.smelter == nil {
		//noinspection GoAssignmentToReceiver
		s = NewSmoker(s.Facing)
	}
	remaining, maximum, cook := s.Durations()
	return map[string]interface{}{
		"BurnTime":     int16(remaining.Milliseconds() / 50),
		"CookTime":     int16(cook.Milliseconds() / 50),
		"BurnDuration": int16(maximum.Milliseconds() / 50),
		"StoredXPInt":  int16(s.Experience()),
		"Items":        nbtconv.InvToNBT(s.Inventory()),
		"id":           "Smoker",
	}
}

// DecodeNBT ...
func (s Smoker) DecodeNBT(data map[string]interface{}) interface{} {
	facing, lit := s.Facing, s.Lit

	//noinspection GoAssignmentToReceiver
	s = NewSmoker(facing)
	s.Lit = lit

	remaining := time.Duration(nbtconv.Map[int16](data, "BurnTime")) * time.Millisecond * 50
	maximum := time.Duration(nbtconv.Map[int16](data, "BurnDuration")) * time.Millisecond * 50
	cook := time.Duration(nbtconv.Map[int16](data, "CookTime")) * time.Millisecond * 50
	s.UpdateDurations(remaining, maximum, cook)

	nbtconv.InvFromNBT(s.Inventory(), nbtconv.Map[[]any](data, "Items"))
	return s
}

// allSmokers ...
func allSmokers() (smokers []world.Block) {
	for _, face := range cube.Faces() {
		smokers = append(smokers, Smoker{Facing: face})
		smokers = append(smokers, Smoker{Facing: face, Lit: true})
	}
	return
}
