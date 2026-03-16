package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// windChargeBurstRadius is the radius within which entities are knocked back
// by a wind charge burst.
const windChargeBurstRadius = 2.0

// NewWindCharge creates a wind charge entity at a position with an owner
// entity. Wind charges fly in a straight line (no gravity) and create a burst
// of wind on impact that knocks back nearby entities and toggles certain
// interactive blocks.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	conf.Owner = owner.H()
	return opts.New(WindChargeType, conf)
}

var windChargeConf = ProjectileBehaviourConfig{
	Gravity:  0,
	Drag:     0.01,
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

	// Knock back all living entities within the burst radius, including the
	// directly-hit entity which receives both damage and burst knockback.
	box := e.H().Type().BBox(e).Translate(pos).Grow(windChargeBurstRadius)
	for other := range tx.EntitiesWithin(box) {
		if other.H() == e.H() {
			continue
		}
		l, ok := other.(Living)
		if !ok || other.Position().Sub(pos).LenSqr() > windChargeBurstRadius*windChargeBurstRadius {
			continue
		}
		l.KnockBack(pos, 0.45, 0.3608)
	}

	// Toggle interactive blocks at the impact point.
	if r, ok := target.(trace.BlockResult); ok {
		toggleWindChargeBlock(r.BlockPosition(), r.Face(), tx)
	}
}

// toggleWindChargeBlock toggles a block at pos if it is a door, trapdoor or
// fence gate. Other activatable blocks (chests, crafting tables, etc.) are not
// affected.
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
