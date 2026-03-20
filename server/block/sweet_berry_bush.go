package block

import (
	"math"
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SweetBerryBush is a plant block that slows entities and can be harvested for sweet berries.
type SweetBerryBush struct {
	replaceable
	transparent
	empty

	// Growth is the current growth stage of the bush. 7 is fully grown.
	Growth int
}

// Pick ...
func (SweetBerryBush) Pick() item.Stack {
	return item.NewStack(SweetBerries{}, 1)
}

// BoneMeal ...
func (b SweetBerryBush) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if b.Growth == 7 {
		return false
	}
	b.Growth++
	tx.SetBlock(pos, b, nil)
	return true
}

// FlammabilityInfo ...
func (SweetBerryBush) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// HasLiquidDrops ...
func (SweetBerryBush) HasLiquidDrops() bool {
	return true
}

// Activate ...
func (b SweetBerryBush) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Item().(item.BoneMeal); ok && b.Growth < 7 {
		return false
	}
	if b.Growth < 4 {
		return false
	}

	count := rand.IntN(2) + 1
	if b.Growth == 7 {
		count++
	}
	dropItem(tx, item.NewStack(SweetBerries{}, count), pos.Vec3Centre())

	b.Growth = 2
	tx.SetBlock(pos, b, nil)
	return true
}

// EntityInside ...
func (b SweetBerryBush) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	living, ok := e.(livingEntity)
	if !ok {
		return
	}

	var movement mgl64.Vec3
	if v, ok := e.(velocityEntity); ok {
		movement = v.Velocity()
		vel := movement
		vel[0] *= 0.8
		vel[1] *= 0.75
		vel[2] *= 0.8
		v.SetVelocity(vel)
	}

	if b.Growth < 2 {
		return
	}
	if math.Abs(movement[0]) >= 0.003 || math.Abs(movement[2]) >= 0.003 {
		living.Hurt(0.5, DamageSource{Block: b})
	}
}

// NeighbourUpdateTick ...
func (b SweetBerryBush) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !b.canGrowOn(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(b, pos, tx)
	}
}

// RandomTick ...
func (b SweetBerryBush) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Growth == 7 || tx.Light(pos.Side(cube.FaceUp)) < 9 || r.IntN(5) != 0 {
		return
	}
	b.Growth++
	tx.SetBlock(pos, b, nil)
}

// BreakInfo ...
func (b SweetBerryBush) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, nothingEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		var count int
		switch {
		case b.Growth == 7:
			count = rand.IntN(2+fortuneLevel(enchantments)) + 2
		case b.Growth >= 4:
			count = rand.IntN(2+fortuneLevel(enchantments)) + 1
		default:
			return nil
		}
		return []item.Stack{item.NewStack(SweetBerries{}, count)}
	})
}

// EncodeItem ...
func (SweetBerryBush) EncodeItem() (name string, meta int16) {
	return "minecraft:sweet_berry_bush", 0
}

// EncodeBlock ...
func (b SweetBerryBush) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:sweet_berry_bush", map[string]any{"growth": int32(b.Growth)}
}

func (SweetBerryBush) canGrowOn(block world.Block) bool {
	switch block.(type) {
	case Farmland, Grass, Dirt, Podzol, Mud, MuddyMangroveRoots:
		return true
	}
	return false
}

// allSweetBerryBushes ...
func allSweetBerryBushes() (bushes []world.Block) {
	for growth := 0; growth <= 7; growth++ {
		bushes = append(bushes, SweetBerryBush{Growth: growth})
	}
	return
}
