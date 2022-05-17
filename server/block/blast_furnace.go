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

// BlastFurnace is a block that smelts ores, raw metals, iron and gold armor and tools, similar to a furnace, but at
// twice the speed. It also serves as an armorer's job site block.
type BlastFurnace struct {
	solid
	bassDrum
	*smelter

	// Facing is the direction the blast furnace is facing.
	Facing cube.Face
	// Lit is true if the blast furnace is lit.
	Lit bool
}

// NewBlastFurnace creates a new initialised blast furnace. The smelter is properly initialised.
func NewBlastFurnace(face cube.Face) BlastFurnace {
	return BlastFurnace{
		Facing:  face,
		smelter: newSmelter(),
	}
}

// Tick is called to check if the blast furnace should update and start or stop smelting.
func (b BlastFurnace) Tick(_ int64, pos cube.Pos, w *world.World) {
	if b.Lit && rand.Float64() <= 0.016 { // Every three or so seconds.
		w.PlaySound(pos.Vec3Centre(), sound.BlastFurnaceCrackle{})
	}
	if lit := b.smelter.tickSmelting(2, b.Lit, func(i item.SmeltInfo) bool {
		return i.Ores
	}); b.Lit != lit {
		b.Lit = lit
		w.SetBlock(pos, b, nil)
	}
}

// EncodeItem ...
func (b BlastFurnace) EncodeItem() (name string, meta int16) {
	return "minecraft:blast_furnace", 0
}

// EncodeBlock ...
func (b BlastFurnace) EncodeBlock() (name string, properties map[string]interface{}) {
	if b.Lit {
		return "minecraft:lit_blast_furnace", map[string]interface{}{"facing_direction": int32(b.Facing)}
	}
	return "minecraft:blast_furnace", map[string]interface{}{"facing_direction": int32(b.Facing)}
}

// UseOnBlock ...
func (b BlastFurnace) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}

	place(w, pos, NewBlastFurnace(user.Facing().Face().Opposite()), user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b BlastFurnace) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3.5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// Activate ...
func (b BlastFurnace) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// EncodeNBT ...
func (b BlastFurnace) EncodeNBT() map[string]interface{} {
	if b.smelter == nil {
		//noinspection GoAssignmentToReceiver
		b = NewBlastFurnace(b.Facing)
	}
	return map[string]interface{}{
		"BurnTime": int32(b.remainingDuration.Milliseconds() / 50),
		"CookTime": int32(b.cookDuration.Milliseconds() / 50),
		"MaxTime":  int32(b.maxDuration.Milliseconds() / 50),
		"Items":    nbtconv.InvToNBT(b.inventory),
		"id":       "BlastFurnace",
	}
}

// DecodeNBT ...
func (b BlastFurnace) DecodeNBT(data map[string]interface{}) interface{} {
	facing, lit := b.Facing, b.Lit

	//noinspection GoAssignmentToReceiver
	b = NewBlastFurnace(facing)
	b.Lit = lit

	b.remainingDuration = time.Duration(nbtconv.Map[int32](data, "BurnTime")) * time.Millisecond * 50
	b.cookDuration = time.Duration(nbtconv.Map[int32](data, "CookTime")) * time.Millisecond * 50
	b.maxDuration = time.Duration(nbtconv.Map[int32](data, "MaxTime")) * time.Millisecond * 50
	nbtconv.InvFromNBT(b.inventory, nbtconv.Map[[]any](data, "Items"))
	return b
}

// allBlastFurnaces ...
func allBlastFurnaces() (furnaces []world.Block) {
	for _, face := range cube.Faces() {
		furnaces = append(furnaces, BlastFurnace{Facing: face})
		furnaces = append(furnaces, BlastFurnace{Facing: face, Lit: true})
	}
	return
}
