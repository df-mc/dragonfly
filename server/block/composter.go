package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
	"time"
)

// Composter is a block that can turn biological matter in to compost which can then produce bone meal. It is also the
// work station for a farming villager.
type Composter struct {
	bass
	transparent
	sourceWaterDisplacer

	// Level is the level of compost inside the composter. At level 8 it can be collected in the form of bone meal.
	Level int
}

// Model ...
func (c Composter) Model() world.BlockModel {
	return model.Composter{Level: c.Level}
}

// FuelInfo ...
func (c Composter) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// FlammabilityInfo ...
func (c Composter) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 20, true)
}

// SideClosed ...
func (c Composter) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (c Composter) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(c)).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		if c.Level == 8 {
			dropItem(w, item.NewStack(item.BoneMeal{}, 1), pos.Side(cube.FaceUp).Vec3Middle())
		}
	})
}

// Activate ...
func (c Composter) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if c.Level >= 7 {
		if c.Level == 8 {
			c.Level = 0
			w.SetBlock(pos, c, nil)
			dropItem(w, item.NewStack(item.BoneMeal{}, 1), pos.Side(cube.FaceUp).Vec3Middle())
			w.PlaySound(pos.Vec3(), sound.ComposterEmpty{})
		}
		return false
	}
	it, _ := u.HeldItems()
	compostable, ok := it.Item().(item.Compostable)
	if !ok {
		return false
	}
	ctx.SubtractFromCount(1)
	w.AddParticle(pos.Vec3(), particle.BoneMeal{})
	if rand.Float64() > compostable.CompostChance() {
		w.PlaySound(pos.Vec3(), sound.ComposterFill{})
		return true
	}
	c.Level++
	w.SetBlock(pos, c, nil)
	w.PlaySound(pos.Vec3(), sound.ComposterFillLayer{})
	if c.Level == 7 {
		w.ScheduleBlockUpdate(pos, time.Second)
	}
	return true
}

// ScheduledTick ...
func (c Composter) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if c.Level == 7 {
		c.Level = 8
		w.SetBlock(pos, c, nil)
		w.PlaySound(pos.Vec3(), sound.ComposterReady{})
	}
}

// EncodeItem ...
func (c Composter) EncodeItem() (name string, meta int16) {
	return "minecraft:composter", 0
}

// EncodeBlock ...
func (c Composter) EncodeBlock() (string, map[string]any) {
	return "minecraft:composter", map[string]any{"composter_fill_level": int32(c.Level)}
}

// allComposters ...
func allComposters() (all []world.Block) {
	for i := 0; i < 9; i++ {
		all = append(all, Composter{Level: i})
	}
	return
}
