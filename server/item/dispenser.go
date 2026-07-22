package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func dispenseProjectile(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext, played world.Sound, create func(world.EntitySpawnOpts) *world.EntityHandle) DispenseResult {
	direction := cube.Pos{}.Side(face).Vec3()
	r := ctx.Rand
	opts := world.EntitySpawnOpts{
		Position: pos.Vec3Centre().Add(direction.Mul(0.7)),
		Velocity: direction.Mul(1.1).Add(mgl64.Vec3{r.Float64()*0.1 - 0.05, r.Float64()*0.1 - 0.05, r.Float64()*0.1 - 0.05}),
	}
	ctx.SubtractFromCount(1)
	tx.AddEntity(create(opts))
	tx.PlaySound(pos.Vec3Centre(), played)
	return DispenseSuccess
}
