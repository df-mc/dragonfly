package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// FireCharge is an item that can be used to place fire when used on a block, or shot from a dispenser to create a small
// fireball.
type FireCharge struct{}

// EncodeItem ...
func (f FireCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:fire_charge", 0
}

// UseOnBlock ...
func (f FireCharge) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if l, ok := tx.Block(pos).(ignitable); ok && l.Ignite(pos, tx, user) {
		ctx.SubtractFromCount(1)
		tx.PlaySound(pos.Vec3Centre(), sound.FireCharge{})
		return true
	} else if s := pos.Side(face); tx.Block(s) == air() {
		ctx.SubtractFromCount(1)
		tx.PlaySound(s.Vec3Centre(), sound.FireCharge{})

		flame := fire()
		tx.SetBlock(s, flame, nil)
		tx.ScheduleBlockUpdate(s, flame, time.Duration(30+rand.IntN(10))*time.Second/20)
		return true
	}
	return false
}
