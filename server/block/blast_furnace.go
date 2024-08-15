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
// The empty value of BlastFurnace is not valid. It must be created using block.NewBlastFurnace(cube.Face).
type BlastFurnace struct {
	solid
	bassDrum
	*smelter

	// Facing is the direction the blast furnace is facing.
	Facing cube.Direction
	// Lit is true if the blast furnace is lit.
	Lit bool
}

// NewBlastFurnace creates a new initialised blast furnace. The smelter is properly initialised.
func NewBlastFurnace(face cube.Direction) BlastFurnace {
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
	if lit := b.smelter.tickSmelting(time.Second*5, time.Millisecond*200, b.Lit, func(i item.SmeltInfo) bool {
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
		return "minecraft:lit_blast_furnace", map[string]interface{}{"minecraft:cardinal_direction": b.Facing.String()}
	}
	return "minecraft:blast_furnace", map[string]interface{}{"minecraft:cardinal_direction": b.Facing.String()}
}

// UseOnBlock ...
func (b BlastFurnace) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}

	place(w, pos, NewBlastFurnace(user.Rotation().Direction().Opposite()), user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b BlastFurnace) BreakInfo() BreakInfo {
	xp := b.Experience()
	return newBreakInfo(3.5, alwaysHarvestable, pickaxeEffective, oneOf(b)).withXPDropRange(xp, xp).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		for _, i := range b.Inventory(w, pos).Clear() {
			dropItem(w, i, pos.Vec3())
		}
	})
}

// Activate ...
func (b BlastFurnace) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
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
	remaining, maximum, cook := b.Durations()
	return map[string]interface{}{
		"BurnTime":     int16(remaining.Milliseconds() / 50),
		"CookTime":     int16(cook.Milliseconds() / 50),
		"BurnDuration": int16(maximum.Milliseconds() / 50),
		"StoredXPInt":  int16(b.Experience()),
		"Items":        nbtconv.InvToNBT(b.inventory),
		"id":           "BlastFurnace",
	}
}

// DecodeNBT ...
func (b BlastFurnace) DecodeNBT(data map[string]interface{}) interface{} {
	remaining := nbtconv.TickDuration[int16](data, "BurnTime")
	maximum := nbtconv.TickDuration[int16](data, "BurnDuration")
	cook := nbtconv.TickDuration[int16](data, "CookTime")

	xp := int(nbtconv.Int16(data, "StoredXPInt"))
	lit := b.Lit

	//noinspection GoAssignmentToReceiver
	b = NewBlastFurnace(b.Facing)
	b.Lit = lit
	b.setExperience(xp)
	b.setDurations(remaining, maximum, cook)
	nbtconv.InvFromNBT(b.inventory, nbtconv.Slice(data, "Items"))
	return b
}

// allBlastFurnaces ...
func allBlastFurnaces() (furnaces []world.Block) {
	for _, face := range cube.Directions() {
		furnaces = append(furnaces, BlastFurnace{Facing: face})
		furnaces = append(furnaces, BlastFurnace{Facing: face, Lit: true})
	}
	return
}
