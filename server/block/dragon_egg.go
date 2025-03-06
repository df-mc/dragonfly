package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// DragonEgg is a decorative block or a "trophy item", and the rarest item in the game.
type DragonEgg struct {
	solid
	transparent
	gravityAffected
	sourceWaterDisplacer
}

// NeighbourUpdateTick ...
func (d DragonEgg) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	d.fall(d, pos, tx)
}

// SideClosed ...
func (d DragonEgg) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// teleport ...
func (d DragonEgg) teleport(pos cube.Pos, tx *world.Tx) {
	for i := 0; i < 1000; i++ {
		newPos := pos.Add(cube.Pos{rand.IntN(31) - 15, max(tx.Range()[0]-pos.Y(), min(tx.Range()[1]-pos.Y(), rand.IntN(15)-7)), rand.IntN(31) - 15})

		if _, ok := tx.Block(newPos).(Air); ok {
			tx.SetBlock(newPos, d, nil)
			tx.SetBlock(pos, nil, nil)
			tx.AddParticle(pos.Vec3(), particle.DragonEggTeleport{Diff: pos.Sub(newPos)})
			return
		}
	}
}

// LightEmissionLevel ...
func (d DragonEgg) LightEmissionLevel() uint8 {
	return 1
}

// Punch ...
func (d DragonEgg) Punch(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User) {
	if gm, ok := u.(interface{ GameMode() world.GameMode }); ok && gm.GameMode().CreativeInventory() {
		return
	}
	d.teleport(pos, tx)
}

// Activate ...
func (d DragonEgg) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	d.teleport(pos, tx)
	return true
}

// BreakInfo ...
func (d DragonEgg) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(45)
}

// EncodeItem ...
func (DragonEgg) EncodeItem() (name string, meta int16) {
	return "minecraft:dragon_egg", 0
}

// EncodeBlock ...
func (DragonEgg) EncodeBlock() (string, map[string]any) {
	return "minecraft:dragon_egg", nil
}
