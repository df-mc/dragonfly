package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/state"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"go.uber.org/atomic"
)

type PrimedTNT struct {
	world.Entity
	Pos cube.Pos
	W   *world.World

	exploding atomic.Bool
}

func (*PrimedTNT) Close() error {
	return nil
}

func (*PrimedTNT) EncodeEntity() string {
	return "minecraft:tnt"
}

func (*PrimedTNT) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 0.5, 0.5})
}

//func (p *PrimedTNT) Tick(current int64) {
//	return
//}

func (p *PrimedTNT) World() *world.World {
	return p.W
}

func (p *PrimedTNT) Position() mgl64.Vec3 {
	return p.Pos.Vec3()
}

func (PrimedTNT) Rotation() (yaw, pitch float64) {
	return 0, 0
}

func (p PrimedTNT) State() (s []state.State) {
	if p.exploding.Load() {
		return []state.State{state.Primed{Time: 80}}
	} else {
		return nil
	}
}

//func (PrimedTNT) Velocity() mgl64.Vec3 {
//	return mgl64.Vec3{0, 1, 0}
//}
