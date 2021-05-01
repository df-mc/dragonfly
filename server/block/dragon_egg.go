package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"math/rand"
)

// DragonEgg is a decorative block or a "trophy item", and the rarest item in the game.
type DragonEgg struct {
	solid
	transparent
	gravityAffected
}

// NeighbourUpdateTick ...
func (d DragonEgg) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	d.fall(d, pos, w)
}

// CanDisplace ...
func (d DragonEgg) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (d DragonEgg) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// teleport ...
func (d DragonEgg) teleport(pos cube.Pos, w *world.World) {
	for i := 0; i < 1000; i++ {
		newPos := pos.Add(cube.Pos{rand.Intn(31) - 15, max(0-pos.Y(), min(255-pos.Y(), rand.Intn(15)-7)), rand.Intn(31) - 15})

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
func (d DragonEgg) Punch(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
	d.teleport(pos, w)
}

// Activate ...
func (d DragonEgg) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
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
func (DragonEgg) EncodeItem() (id int32, name string, meta int16) {
	return 122, "minecraft:dragon_egg", 0
}

// EncodeBlock ...
func (DragonEgg) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:dragon_egg", nil
}
