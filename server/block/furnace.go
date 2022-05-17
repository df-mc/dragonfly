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

// Furnace is a utility block used for the smelting of blocks and items.
type Furnace struct {
	solid
	bassDrum
	*smelter

	// Facing is the direction the furnace is facing.
	Facing cube.Face
	// Lit is true if the furnace is lit.
	Lit bool
}

// NewFurnace creates a new initialised furnace. The smelter is properly initialised.
func NewFurnace(face cube.Face) Furnace {
	return Furnace{
		Facing:  face,
		smelter: newSmelter(),
	}
}

// Tick is called to check if the furnace should update and start or stop smelting.
func (f Furnace) Tick(_ int64, pos cube.Pos, w *world.World) {
	if f.Lit && rand.Float64() <= 0.016 { // Every three or so seconds.
		w.PlaySound(pos.Vec3Centre(), sound.FurnaceCrackle{})
	}
	if lit := f.smelter.tickSmelting(1, f.Lit, func(item.SmeltInfo) bool {
		return true
	}); f.Lit != lit {
		f.Lit = lit
		w.SetBlock(pos, f, nil)
	}
}

// EncodeItem ...
func (f Furnace) EncodeItem() (name string, meta int16) {
	return "minecraft:furnace", 0
}

// EncodeBlock ...
func (f Furnace) EncodeBlock() (name string, properties map[string]interface{}) {
	if f.Lit {
		return "minecraft:lit_furnace", map[string]interface{}{"facing_direction": int32(f.Facing)}
	}
	return "minecraft:furnace", map[string]interface{}{"facing_direction": int32(f.Facing)}
}

// UseOnBlock ...
func (f Furnace) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}

	place(w, pos, NewFurnace(user.Facing().Face().Opposite()), user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (f Furnace) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3.5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(f, 1)),
	}
}

// Activate ...
func (f Furnace) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// EncodeNBT ...
func (f Furnace) EncodeNBT() map[string]interface{} {
	if f.smelter == nil {
		//noinspection GoAssignmentToReceiver
		f = NewFurnace(f.Facing)
	}
	return map[string]interface{}{
		"BurnTime": int32(f.remainingDuration.Milliseconds() / 50),
		"CookTime": int32(f.cookDuration.Milliseconds() / 50),
		"MaxTime":  int32(f.maxDuration.Milliseconds() / 50),
		"Items":    nbtconv.InvToNBT(f.inventory),
		"id":       "Furnace",
	}
}

// DecodeNBT ...
func (f Furnace) DecodeNBT(data map[string]interface{}) interface{} {
	//noinspection GoAssignmentToReceiver
	f = NewFurnace(f.Facing)
	f.remainingDuration = time.Duration(nbtconv.Map[int32](data, "BurnTime")) * time.Millisecond * 50
	f.cookDuration = time.Duration(nbtconv.Map[int32](data, "CookTime")) * time.Millisecond * 50
	f.maxDuration = time.Duration(nbtconv.Map[int32](data, "MaxTime")) * time.Millisecond * 50
	nbtconv.InvFromNBT(f.inventory, nbtconv.Map[[]any](data, "Items"))
	return f
}

// allFurnaces ...
func allFurnaces() (furnaces []world.Block) {
	for _, face := range cube.Faces() {
		furnaces = append(furnaces, Furnace{Facing: face})
		furnaces = append(furnaces, Furnace{Facing: face, Lit: true})
	}
	return
}