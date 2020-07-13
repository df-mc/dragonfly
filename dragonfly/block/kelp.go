package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Kelp is an underwater block which can grow on top of solids underwater.
type Kelp struct {

	// Age is the age of the kelp block which can be 0-15. If age is 15, kelp won't grow any further.
	Age int
}

// BreakInfo ...
func (k Kelp) BreakInfo() BreakInfo { // Kelp can be instantly destroyed.
	return BreakInfo{
		Hardness:    0.0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(Kelp{}, 1)),
	}
}

func (Kelp) EncodeItem() (id int32, meta int16) {
	return 335, 0
}

func (k Kelp) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:kelp", map[string]interface{}{"age": int32(k.Age)}
}

func (Kelp) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water // Kelp can waterlog.
}

func (Kelp) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false // Kelp can always be flowed through.
}

func (Kelp) AABB(world.BlockPos, *world.World) []physics.AABB {
	return nil // Kelp can be placed even if someone is standing on its placement position.
}

// SetRandomAge ...
func (k Kelp) setRandomAge() Kelp {
	// In Java Edition, the age value can be up to 25, but MCPE limits it to 15.
	k.Age = rand.Intn(14)
	return k
}

func (k Kelp) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, k)
	if !used {
		return
	}

	switch w.Block(pos.Add(world.BlockPos{0, -1})).(type) {
	// Kelp blocks must be placed on a solid or another kelp block, TODO: Replace this to check for a solid in the future when a Solid interface exists.
	case Air, Water:
		return false
	}

	liquid, liquidExists := w.Liquid(pos)                                            // A check for existent water at this position, kelp cannot be placed if none is found.
	if !liquidExists || liquid.LiquidType() != "water" || liquid.LiquidDepth() < 8 { // Water must be a source block to plant kelp.
		return false
	}

	// When first placed, kelp gets a random age between 0 and 14 in MCBE.
	place(w, pos, k.setRandomAge(), user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (k Kelp) NeighbourUpdateTick(pos, changed world.BlockPos, w *world.World) {
	if changed.Y()-1 == pos.Y() { // When a kelp block is broken above, the kelp block underneath it gets a new random age.
		w.PlaceBlock(pos, k.setRandomAge())
	}

	switch w.Block(pos.Add(world.BlockPos{0, -1})).(type) {
	// Kelp blocks can only exist on top of a solid or another kelp block, TODO: Replace this to check for a solid in the future when a Solid interface exists.
	case Air, Water:
		w.ScheduleBlockUpdate(pos, time.Second/20)
	}
}

// ScheduledTick ...
func (Kelp) ScheduledTick(pos world.BlockPos, w *world.World) {
	// Kelp blocks can only exist on top of a solid or another kelp block, TODO: Replace this to check for a solid in the future when a Solid interface exists.
	switch w.Block(pos.Add(world.BlockPos{0, -1})).(type) { // As of now, the breaking logic has to be in here as well to avoid issues.
	case Air, Water:
		w.BreakBlock(pos)
	}
}

// RandomTick ...
func (k Kelp) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if r.Intn(100) < 15 && k.Age < 15 { // Every random tick, there's a 14% chance for Kelp to grow if its age is below 15.
		abovePos := pos.Add(world.BlockPos{0, 1})

		liquid, liquidAboveExists := w.Liquid(abovePos)

		// For kelp to grow, there must be only water above.
		if liquidAboveExists && liquid.LiquidType() == "water" {
			switch w.Block(abovePos).(type) {
			case Air, Water:
				w.PlaceBlock(abovePos, Kelp{Age: k.Age + 1})
				if liquid.LiquidDepth() < 8 { // When kelp grows into a water block, the water block becomes a source block.
					w.SetLiquid(abovePos, Water{true, 8, false})
				}
			}
		}
	}
	w.ScheduleBlockUpdate(pos, time.Second/20)
}

// allKelp returns all possible states of a kelp block.
func allKelp() (b []world.Block) {
	for i := 15; i >= 0; i-- {
		b = append(b, Kelp{Age: i})
	}
	return
}
