package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

type Kelp struct {

	//0-15, defines the age of the kelp block. If age is 15, kelp won't grow any further.
	Age int
}

// BreakInfo ...
func (k Kelp) BreakInfo() BreakInfo { //Kelp can be instantly destroyed.
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

func (Kelp) LightLevelRequired() uint8 {
	return 0 //Kelp grows without any light.
}

func (Kelp) RequiresHydration() bool {
	return false //Kelp does not need hydration but rather just has to be placed in water.
}

func (Kelp) RequiresFarmland() bool {
	return false //Kelp doesn't require a farmland, literally grows underwater.
}

func (Kelp) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water //Kelp can waterlog.
}

func (Kelp) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false //Kelp can always be flowed through.
}

func (Kelp) AABB(world.BlockPos, *world.World) []physics.AABB {
	return nil //Kelp can be placed even if someone is standing on its placement position.
}

// SetRandomAge ...
func (k Kelp) SetRandomAge() Kelp {
	//In Java Edition, the age value can be up to 25, but MCPE limits it to 15.
	k.Age = rand.Intn(14)
	return k
}

// SetAge ...
func (k Kelp) SetAge(age int) Kelp { //The age must be 0-15.
	k.Age = age
	return k
}

func (k Kelp) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, k)
	if !used {
		return
	}

	blockBelow, _ := w.Block(pos.Add(world.BlockPos{0, -1})).EncodeBlock()
	if blockBelow == "minecraft:air" || blockBelow == "minecraft:water" || blockBelow == "minecraft:flowing_water" { //Kelp cannot be placed on top of some blocks.
		return
	}

	liquid, liquidExists := w.Liquid(pos) //A check for existent water at this position, kelp cannot be placed if none is found.
	if !liquidExists || liquid.LiquidType() != "water" || liquid.LiquidDepth() < 8 { //Water must be a source block to plant kelp.
		return
	}

	//When first placed, kelp gets a random age between 0 and 14 in MCBE.
	place(w, pos, k.SetRandomAge(), user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (k Kelp) NeighbourUpdateTick(pos, changed world.BlockPos, w *world.World) {
	if changed.Y()-1 == pos.Y() { //When a kelp block is broken above, the kelp block underneath it gets a new random age.
		w.SetBlock(pos, k.SetRandomAge())
	}

	blockBelow, _ := w.Block(pos.Add(world.BlockPos{0, -1})).EncodeBlock()

	if blockBelow == "minecraft:air" || blockBelow == "minecraft:water" || blockBelow == "minecraft:flowing_water" {
		w.ScheduleBlockUpdate(pos, time.Second/20)
	}
}

// ScheduledTick ...
func (Kelp) ScheduledTick(pos world.BlockPos, w *world.World) {
	w.BreakBlock(pos)
}

// RandomTick ...
func (k Kelp) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if rand.Intn(100) < 15 && k.Age < 15 { //Every random tick, there's a 14% chance for Kelp to grow if its age is below 15.
		abovePos := pos.Add(world.BlockPos{0, 1})

		liquid, liquidAboveExists := w.Liquid(abovePos)
		block, _ := w.Block(abovePos).EncodeBlock()

		//For kelp to grow, there must be only water above.
		if liquidAboveExists && liquid.LiquidType() == "water" && (block == "minecraft:air" || block == "minecraft:water" || block == "minecraft:flowing_water") {
			w.SetBlock(abovePos, Kelp{}.SetAge(k.Age+1))
			if liquid.LiquidDepth() < 8 { //When kelp grows into a water block, the water block becomes a source block.
				w.SetLiquid(abovePos, Water{true, 8, false})
			}
		}
	}
}