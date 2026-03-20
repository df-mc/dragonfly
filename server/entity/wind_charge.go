package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// windChargeBurstRadius is the maximum radius within which entities are
// knocked back by a wind charge burst.
const windChargeBurstRadius = 2.5

// NewWindCharge creates a wind charge entity at a position with an owner
// entity. Wind charges fly in a straight line (no gravity) and create a burst
// of wind on impact that knocks back nearby entities and toggles certain
// interactive blocks.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	conf.Owner = owner.H()
	return opts.New(WindChargeType, conf)
}

// TODO: Wind charges should have increased drag when travelling through water
// or lava, but per-medium drag is not yet supported by ProjectileBehaviour.
var windChargeConf = ProjectileBehaviourConfig{
	Gravity:  0,
	Drag:     0,
	Damage:   -1,
	Particle: particle.WindExplosion{},
	Sound:    sound.WindChargeBurst{},
	Hit:      windChargeBurst,
}

// windChargeBurst is called when a wind charge hits a target. It deals 1 HP
// damage on a direct entity hit, knocks back all living entities within the
// burst radius, and toggles interactive blocks at the impact point.
func windChargeBurst(e *Ent, tx *world.Tx, target trace.Result) {
	pos := target.Position()
	owner, _ := e.Behaviour().(*ProjectileBehaviour).Owner().Entity(tx)

	// Deal flat 1 HP damage to the directly-hit entity.
	if er, ok := target.(trace.EntityResult); ok {
		if l, ok := er.Entity().(Living); ok {
			l.Hurt(1, ProjectileDamageSource{Projectile: e, Owner: owner})
		}
	}

	// Apply knockback to all living entities within the burst radius. Impact
	// scales with distance (closer = stronger) and is split into horizontal
	// and vertical components.
	box := e.H().Type().BBox(e).Translate(pos).Grow(windChargeBurstRadius)
	for other := range tx.EntitiesWithin(box) {
		if other.H() == e.H() {
			continue
		}
		l, ok := other.(Living)
		if !ok {
			continue
		}
		entityPos := other.Position()
		dist := entityPos.Sub(pos).Len()
		impact := 1.3 - dist/windChargeBurstRadius
		if impact <= 0 {
			continue
		}

		vel := l.Velocity()
		// If the entity is directly above the impact, apply a flat upward
		// boost. Otherwise split into horizontal and vertical components.
		dx := entityPos[0] - pos[0]
		dz := entityPos[2] - pos[2]
		if dx*dx+dz*dz < 0.01 {
			vel[1] += 1.1
		} else {
			dir := entityPos.Sub(pos)
			dir[1] = 0
			dir = dir.Normalize()
			vel = vel.Add(dir.Mul(impact))
			vel[1] += impact * 0.4
		}
		l.SetVelocity(vel)
	}

	// Toggle interactive blocks at the impact point.
	if r, ok := target.(trace.BlockResult); ok {
		toggleWindChargeBlock(r.BlockPosition(), r.Face(), tx)
	}
}

// toggleWindChargeBlock toggles a block at pos if it is a door, trapdoor or
// fence gate.
// TODO: Buttons, levers, bells and candles should also be toggled.
func toggleWindChargeBlock(pos cube.Pos, face cube.Face, tx *world.Tx) {
	b := tx.Block(pos)
	switch b.(type) {
	case block.WoodDoor, block.CopperDoor, block.WoodTrapdoor, block.CopperTrapdoor, block.WoodFenceGate:
		b.(block.Activatable).Activate(pos, face, tx, nil, nil)
	}
}

// WindChargeType is a world.EntityType implementation for WindCharge.
var WindChargeType windChargeType

type windChargeType struct{}

func (windChargeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (windChargeType) EncodeEntity() string { return "minecraft:wind_charge_projectile" }
func (windChargeType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.15625, 0, -0.15625, 0.15625, 0.3125, 0.15625)
}

func (windChargeType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = windChargeConf.New()
}
func (windChargeType) EncodeNBT(*world.EntityData) map[string]any { return nil }
