package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// BambooSapling is the initial stage of bamboo growth. It appears as a small shoot
// and grows into a bamboo stalk over time.
type BambooSapling struct {
	empty
	transparent
	Age bool
}

var _ item.BoneMealAffected = BambooSapling{}

// UseOnBlock places a bamboo sapling on valid soil.
func (b BambooSapling) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if !supportsVegetation(Bamboo{}, tx.Block(below)) && !isBambooSupport(tx.Block(below)) {
		return false
	}
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick breaks the sapling if it loses support.
func (b BambooSapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	below := pos.Side(cube.FaceDown)
	if !supportsVegetation(Bamboo{}, tx.Block(below)) && !isBambooSupport(tx.Block(below)) {
		breakBlock(b, pos, tx)
		tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
}

// RandomTick grows the sapling into a bamboo stalk.
// RandomTick seeds the ScheduledTick growth chain for world-generated saplings.
// For player-placed saplings the chain is started directly in Bamboo.UseOnBlock.
func (b BambooSapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Age {
		return
	}
	tx.ScheduleBlockUpdate(pos, b, bambooGrowthDelay(r))
}

// ScheduledTick converts the sapling to a 2-block bamboo stalk once there is
// enough light (>= 9) and space above. If not, it re-schedules and waits.
func (b BambooSapling) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Age {
		return
	}
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return
	}
	delay := bambooGrowthDelay(r)
	t := tx.World().Time() % 24000
	if t >= 13000 && t < 23000 {
		// Nighttime — wait and retry.
		tx.ScheduleBlockUpdate(pos, b, delay)
		return
	}
	// Convert: bottom = aged bamboo, top = fresh growable bamboo.
	tx.SetBlock(pos, Bamboo{Age: true, LeafSize: bambooNoLeaves, Thick: false}, nil)
	newTop := Bamboo{Age: false, LeafSize: SmallLeaves, Thick: false}
	tx.SetBlock(above, newTop, nil)
	// Seed the growth chain on the new bamboo top.
	tx.ScheduleBlockUpdate(above, newTop, delay)
}

// BoneMeal grows the sapling into a bamboo stalk immediately.
func (b BambooSapling) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return false
	}
	// Bottom becomes aged (age_bit=1), top is fresh growable (age_bit=0).
	tx.SetBlock(pos, Bamboo{Age: true, LeafSize: bambooNoLeaves, Thick: false}, nil)
	tx.SetBlock(above, Bamboo{Age: false, LeafSize: SmallLeaves, Thick: false}, nil)
	return true
}

// BreakInfo ...
func (BambooSapling) BreakInfo() BreakInfo {
	b := Bamboo{}
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       oneOf(b),
		BreakHandler: func(pos cube.Pos, tx *world.Tx, u item.User) {
			tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
		},
	}
}

// HasLiquidDrops ...
func (BambooSapling) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (BambooSapling) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// CompostChance ...
func (BambooSapling) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (BambooSapling) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo", 0
}

// EncodeBlock ...
func (b BambooSapling) EncodeBlock() (string, map[string]any) {
	return "minecraft:bamboo_sapling", map[string]any{"age_bit": boolByte(b.Age)}
}

// allBambooSapling returns all bamboo sapling block states.
func allBambooSapling() (blocks []world.Block) {
	for _, age := range []bool{false, true} {
		blocks = append(blocks, BambooSapling{Age: age})
	}
	return
}

// isBambooSupport checks if a block can support bamboo or bamboo sapling.
func isBambooSupport(b world.Block) bool {
	switch b.(type) {
	case Bamboo, BambooSapling:
		return true
	}
	return false
}
