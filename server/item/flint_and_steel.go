package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// FlintAndSteel is an item used to light blocks on fire.
type FlintAndSteel struct{}

// MaxCount ...
func (f FlintAndSteel) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (f FlintAndSteel) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 65,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// ignitable represents a block that can be lit by a fire emitter, such as flint and steel.
type ignitable interface {
	// Ignite is called when the block is lit by flint and steel.
	Ignite(pos cube.Pos, tx *world.Tx, igniter world.Entity) bool
}

// UseOnBlock ...
func (f FlintAndSteel) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	ctx.DamageItem(1)
	if l, ok := tx.Block(pos).(ignitable); ok {
		return l.Ignite(pos, tx, user)
	}
	if s := pos.Side(face); tx.Block(s) == air() {
		tx.PlaySound(s.Vec3Centre(), sound.Ignite{})
		if portal.ActivateNetherPortal(tx, s) {
			return true
		}

		flame := fire(s, tx)
		tx.SetBlock(s, flame, nil)
		tx.ScheduleBlockUpdate(s, flame, time.Duration(30+rand.IntN(10))*time.Second/20)
		return true
	}
	return false
}

// EncodeItem ...
func (f FlintAndSteel) EncodeItem() (name string, meta int16) {
	return "minecraft:flint_and_steel", 0
}

// air returns an air block.
func air() world.Block {
	a, ok := world.BlockByName("minecraft:air", nil)
	if !ok {
		panic("could not find air block")
	}
	return a
}

// fire returns a fire block to place at the position passed, with its state
// derived from the blocks around that position.
func fire(pos cube.Pos, tx *world.Tx) world.Block {
	f, ok := world.BlockByName("minecraft:fire", map[string]any{
		"age":                        int32(0),
		"minecraft:connection_north": false,
		"minecraft:connection_east":  false,
		"minecraft:connection_south": false,
		"minecraft:connection_west":  false,
		"minecraft:connection_up":    false,
	})
	if !ok {
		panic("could not find fire block")
	}
	if deriver, ok := f.(world.StateDeriver); ok {
		return deriver.DeriveState(pos, tx)
	}
	return f
}
