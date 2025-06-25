package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand/v2"
	"time"
)

// Target is a block that provides redstone power based on how close a projectile hits its center.
type Target struct {
	solid

	// Power is the redstone power level emitted by the target block, ranging from 0 to 15.
	Power int
}

// BreakInfo ...
func (t Target) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, hoeEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateAroundRedstone(pos, tx)
	})
}

// RedstoneSource ...
func (t Target) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (t Target) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	return t.Power
}

// StrongPower ...
func (t Target) StrongPower(_ cube.Pos, _ cube.Face, _ *world.Tx, _ bool) int {
	return t.Power
}

// HitByProjectile handles when a projectile hits the target block.
func (t Target) HitByProjectile(pos mgl64.Vec3, blockPos cube.Pos, tx *world.Tx, delay time.Duration) {
	center := blockPos.Vec3Centre()
	distance := pos.Sub(center).Len()

	maxDistance := math.Sqrt(0.75)
	normalizedDistance := math.Min(distance/maxDistance, 1.0)

	var rawPower float64
	if normalizedDistance <= 0.58 {
		rawPower = 15
	} else if normalizedDistance > 0.9 {
		rawPower = 0
	} else {
		rawPower = 15 * (1 - math.Pow(normalizedDistance, 4.5))
	}

	t.Power = int(math.Max(math.Round(rawPower), 0))
	tx.SetBlock(blockPos, t, &world.SetOpts{DisableBlockUpdates: false})
	tx.PlaySound(blockPos.Vec3Centre(), sound.PowerOn{})

	tx.ScheduleBlockUpdate(blockPos, t, delay)
}

// ScheduledTick ...
func (t Target) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if t.Power > 0 {
		t.Power = 0
		tx.SetBlock(pos, t, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	}
}

// DecodeNBT ...
func (t Target) DecodeNBT(data map[string]any) any {
	t.Power = int(nbtconv.Int32(data, "Power"))
	return t
}

// EncodeNBT ...
func (t Target) EncodeNBT() map[string]any {
	m := map[string]any{
		"Power": int32(t.Power),
	}
	return m
}

// EncodeItem ...
func (t Target) EncodeItem() (name string, meta int16) {
	return "minecraft:target", 0
}

// EncodeBlock ...
func (t Target) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:target", nil
}
