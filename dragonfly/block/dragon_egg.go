package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/particle"
	"math/rand"
)

// DragonEgg is a decorative block or a "trophy item", and the rarest item in the game.
type DragonEgg struct {
	noNBT
	solid
	transparent
	gravityAffected
}

// NeighbourUpdateTick ...
func (d DragonEgg) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	d.fall(d, pos, w)
}

// CanDisplace ...
func (d DragonEgg) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (d DragonEgg) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// teleport ...
func (d DragonEgg) teleport(pos world.BlockPos, w *world.World) {
	for i := 0; i < 1000; i++ {
		newPos := pos.Add(world.BlockPos{rand.Intn(16) - rand.Intn(16), max(0-pos.Y(), min(255-pos.Y(), rand.Intn(8)-rand.Intn(8))), rand.Intn(16) - rand.Intn(16)})

		if _, ok := w.Block(newPos).(Air); ok {
			w.PlaceBlock(newPos, d)
			w.BreakBlockWithoutParticles(pos)
			w.AddParticle(pos.Vec3(), particle.DragonEggTeleport{Diff: pos.Subtract(newPos)})
			return
		}
	}
}

// LightEmissionLevel ...
func (d DragonEgg) LightEmissionLevel() uint8 {
	return 1
}

// Punch ...
func (d DragonEgg) Punch(pos world.BlockPos, _ world.Face, w *world.World, _ item.User) {
	d.teleport(pos, w)
}

// Activate ...
func (d DragonEgg) Activate(pos world.BlockPos, _ world.Face, w *world.World, _ item.User) {
	d.teleport(pos, w)
}

// BreakInfo ...
func (d DragonEgg) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(d, 1)),
	}
}

// EncodeItem ...
func (d DragonEgg) EncodeItem() (id int32, meta int16) {
	return 122, 0
}

// EncodeBlock ...
func (d DragonEgg) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:dragon_egg", nil
}

// Hash ...
func (d DragonEgg) Hash() uint64 {
	return hashDragonEgg
}
