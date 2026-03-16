package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// SweetBerryBush is a plant block that grows sweet berries and pricks entities moving through it.
type SweetBerryBush struct {
	empty
	transparent
	sourceWaterDisplacer

	// Age is the growth stage of the bush. A value of 3 is fully grown.
	Age int
}

// Activate harvests berries from a grown bush or defers to bone meal when appropriate.
func (s SweetBerryBush) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Item().(item.BoneMeal); ok {
		creative := false
		if g, ok := u.(interface{ GameMode() world.GameMode }); ok {
			creative = g.GameMode().CreativeInventory()
		}
		if s.Age < 3 {
			// In creative, the bush grows instantly.
			if !creative {
				return false
			}
			s.Age = 3
			tx.SetBlock(pos, s, nil)
			ctx.SubtractFromCount(1)
			tx.AddParticle(pos.Vec3(), particle.BoneMeal{})
			return true
		}
		// Grow by one age stage in other gamemodes.
		ctx.SubtractFromCount(1)
		tx.AddParticle(pos.Vec3(), particle.BoneMeal{})
		dropItem(tx, item.NewStack(SweetBerries{}, sweetBerryDropCount(s.Age)), pos.Vec3Centre())
		return true
	}
	if s.Age < 2 {
		return true
	}

	dropItem(tx, item.NewStack(SweetBerries{}, sweetBerryDropCount(s.Age)), pos.Vec3Centre())

	s.Age = 1
	tx.SetBlock(pos, s, nil)
	return true
}

// BoneMeal grows the bush by one age stage, up to age 3.
func (s SweetBerryBush) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if s.Age == 3 {
		return false
	}
	s.Age++
	tx.SetBlock(pos, s, nil)
	return true
}

// BreakInfo ...
func (s SweetBerryBush) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(_ item.Tool, enchantments []item.Enchantment) []item.Stack {
		if s.Age < 2 {
			return nil
		}
		fortune := fortuneLevel(enchantments)
		count := 1 + rand.IntN(2+fortune)
		if s.Age == 3 {
			count++
		}
		return []item.Stack{item.NewStack(SweetBerries{}, count)}
	})
}

// EntityInside damages moving, non-sneaking entities that brush through the bush.
func (s SweetBerryBush) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if s.Age == 0 {
		return
	}
	v, ok := e.(velocityEntity)
	if !ok {
		return
	}

	vel := v.Velocity()
	slowed := vel
	slowed[0] *= 0.8
	slowed[1] *= 0.75
	slowed[2] *= 0.8
	v.SetVelocity(slowed)

	if sneaking, ok := e.(interface{ Sneaking() bool }); ok && sneaking.Sneaking() {
		return
	}

	vel[1] = 0
	if mgl64.FloatEqualThreshold(vel.Len(), 0, 0.003) || rand.IntN(20) != 0 {
		return
	}
	if living, ok := e.(livingEntity); ok {
		living.Hurt(1, DamageSource{Block: s})
	}
}

// FlammabilityInfo ...
func (SweetBerryBush) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, false)
}

// HasLiquidDrops ...
func (SweetBerryBush) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (s SweetBerryBush) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsSweetBerryBush(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(s, pos, tx)
	}
}

// Pick ...
func (SweetBerryBush) Pick() item.Stack {
	return item.NewStack(SweetBerries{}, 1)
}

// RandomTick ...
func (s SweetBerryBush) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if s.Age < 3 && r.IntN(5) == 0 && tx.Light(pos.Side(cube.FaceUp)) >= 9 {
		s.Age++
		tx.SetBlock(pos, s, nil)
	}
}

// EncodeBlock ...
func (s SweetBerryBush) EncodeBlock() (string, map[string]any) {
	return "minecraft:sweet_berry_bush", map[string]any{"growth": int32(s.Age)}
}

// allSweetBerryBushes ...
func allSweetBerryBushes() (b []world.Block) {
	for i := 0; i <= 3; i++ {
		b = append(b, SweetBerryBush{Age: i})
	}
	return
}

func sweetBerryDropCount(age int) int {
	if age <= 1 {
		return 0
	}
	amount := 1 + rand.IntN(2)
	if age == 3 {
		amount++
	}
	return amount
}

func supportsSweetBerryBush(b world.Block) bool {
	switch b.(type) {
	case Dirt, Grass, Podzol, Farmland:
		return true
	default:
		return false
	}
}
