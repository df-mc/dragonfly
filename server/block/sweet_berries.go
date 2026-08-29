package block

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// SweetBerries are edible berries that can also be planted to grow a sweet berry bush.
type SweetBerries struct {
	transparent
	empty

	// Growth is the current growth stage of the bush when placed as a block. 3 is fully grown.
	Growth int
}

// AlwaysConsumable ...
func (SweetBerries) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (SweetBerries) ConsumeDuration() time.Duration {
	return item.DefaultConsumeDuration
}

// Consume ...
func (SweetBerries) Consume(_ *world.Tx, c item.Consumer) item.Stack {
	c.Saturate(2, 1.2)
	return item.Stack{}
}

// CompostChance ...
func (SweetBerries) CompostChance() float64 {
	return 0.3
}

// UseOnBlock ...
func (s SweetBerries) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used || !supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	s.Growth = 0
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BoneMeal ...
func (s SweetBerries) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	if s.Growth >= 3 {
		return item.BoneMealResultNone
	}
	s.Growth++
	tx.SetBlock(pos, s, nil)
	return item.BoneMealResultSmall
}

// FlammabilityInfo ...
func (SweetBerries) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// HasLiquidDrops ...
func (SweetBerries) HasLiquidDrops() bool {
	return true
}

// Activate ...
func (s SweetBerries) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, user item.User, _ *item.UseContext) bool {
	if s.Growth < 2 {
		return false
	}
	if s.Growth < 3 && user != nil {
		held, _ := user.HeldItems()
		if _, ok := held.Item().(item.BoneMeal); ok {
			return false
		}
	}

	count := rand.IntN(2) + 1
	if s.Growth >= 3 {
		count++
	}
	dropItem(tx, item.NewStack(SweetBerries{}, count), pos.Vec3Centre())

	s.Growth = 1
	tx.SetBlock(pos, s, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.SweetBerryBushPick{})
	return true
}

// EntityInside ...
func (s SweetBerries) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if h := e.H(); h != nil && h.Type().EncodeEntity() == "minecraft:item" {
		return
	}
	if fall, ok := e.(fallDistanceEntity); ok {
		fall.ResetFallDistance()
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

	if s.Growth < 1 {
		return
	}
	living, ok := e.(livingEntity)
	if !ok {
		return
	}
	if math.Abs(movement[0]) >= 0.003 || math.Abs(movement[2]) >= 0.003 {
		if _, vulnerable := living.Hurt(1, DamageSource{Block: s}); vulnerable {
			tx.PlaySound(pos.Vec3Centre(), sound.SweetBerryBushHurt{})
		}
	}
}

// NeighbourUpdateTick ...
func (s SweetBerries) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(s, pos, tx)
	}
}

// RandomTick ...
func (s SweetBerries) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	abovePos := pos.Side(cube.FaceUp)
	if s.Growth >= 3 || sweetBerryGrowthBlocked(abovePos, tx) || tx.Light(abovePos) < 9 || r.IntN(5) != 0 {
		return
	}
	s.Growth++
	tx.SetBlock(pos, s, nil)
}

// sweetBerryGrowthBlocked reports whether an opaque full block prevents a sweet berry bush from growing.
func sweetBerryGrowthBlocked(pos cube.Pos, tx *world.Tx) bool {
	above := tx.Block(pos)
	boxes := above.Model().BBox(pos, tx)
	if len(boxes) != 1 || boxes[0] != cube.Box(0, 0, 0, 1, 1, 1) {
		return false
	}
	diffuser, ok := above.(LightDiffuser)
	return !ok || diffuser.LightDiffusionLevel() == 15
}

// BreakInfo ...
func (s SweetBerries) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		var count int
		switch {
		case s.Growth >= 3:
			count = rand.IntN(2+fortuneLevel(enchantments)) + 2
		case s.Growth >= 2:
			count = rand.IntN(2+fortuneLevel(enchantments)) + 1
		default:
			return nil
		}
		return []item.Stack{item.NewStack(SweetBerries{}, count)}
	})
}

// EncodeItem ...
func (SweetBerries) EncodeItem() (name string, meta int16) {
	return "minecraft:sweet_berries", 0
}

// EncodeBlock ...
func (s SweetBerries) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:sweet_berry_bush", map[string]any{"growth": int32(s.Growth)}
}

// allSweetBerryBushes ...
func allSweetBerryBushes() (bushes []world.Block) {
	for growth := 0; growth <= 7; growth++ {
		bushes = append(bushes, SweetBerries{Growth: growth})
	}
	return
}
