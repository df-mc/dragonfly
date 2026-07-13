package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	windChargeExplosionPower      = 1.2
	windChargeExplosionDiameter   = windChargeExplosionPower * 2
	windChargeKnockbackMultiplier = 1.22
	windChargeBlockHitOffset      = 0.25
)

// NewWindCharge creates a wind charge entity at a position with an owner
// entity. Wind charges fly in a straight line (no gravity or drag) and create
// a burst of wind on impact.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	conf.Owner = owner.H()
	return opts.New(WindChargeType, conf)
}

var windChargeConf = ProjectileBehaviourConfig{
	Gravity:               0,
	Drag:                  0,
	Damage:                -1,
	Hit:                   windChargeBurst,
	EntityCollisionFilter: windChargeCanHit,
}

// windChargeBurst is called when a wind charge hits a target. It deals 1 HP
// damage on a direct entity hit, applies Java-style explosion knockback and
// toggles interactive blocks at the impact point.
func windChargeBurst(e *Ent, tx *world.Tx, target trace.Result) {
	pos := target.Position()
	if r, ok := target.(trace.BlockResult); ok {
		pos = windChargeBlockExplosionPosition(pos, r.Face())
	}
	tx.AddParticle(pos, particle.WindExplosion{})
	tx.PlaySound(pos, sound.WindChargeBurst{})

	var owner world.Entity
	if h := e.Behaviour().(*ProjectileBehaviour).Owner(); h != nil {
		owner, _ = h.Entity(tx)
	}
	if er, ok := target.(trace.EntityResult); ok {
		if l, ok := er.Entity().(Living); ok {
			l.Hurt(1, ProjectileDamageSource{Projectile: e, Owner: owner})
		}
	}

	d := windChargeExplosionDiameter
	box := cube.Box(
		math.Floor(pos[0]-d-1),
		math.Floor(pos[1]-d-1),
		math.Floor(pos[2]-d-1),
		math.Floor(pos[0]+d+1),
		math.Floor(pos[1]+d+1),
		math.Floor(pos[2]+d+1),
	)
	for other := range tx.EntitiesWithin(box) {
		if other.H() == e.H() || !windChargeCanHitType(other.H().Type().EncodeEntity()) {
			continue
		}
		moving, ok := other.(interface {
			Velocity() mgl64.Vec3
			SetVelocity(mgl64.Vec3)
		})
		if !ok {
			continue
		}
		velocity := moving.Velocity()
		knockedBack := windChargeKnockback(
			pos,
			other.Position(),
			EyePosition(other),
			velocity,
			block.ExplosionExposure(tx, pos, other),
		)
		if knockedBack != velocity {
			moving.SetVelocity(knockedBack)
		}
	}

	if r, ok := target.(trace.BlockResult); ok {
		blockPos := r.BlockPosition()
		if b, ok := tx.Block(blockPos).(block.WindChargeAffected); ok {
			b.Activate(blockPos, r.Face(), tx, nil, nil)
		}
	}
}

func windChargeKnockback(burst, position, eye, velocity mgl64.Vec3, exposure float64) mgl64.Vec3 {
	distance := position.Sub(burst).Len() / windChargeExplosionDiameter
	if distance > 1 || exposure == 0 {
		return velocity
	}
	direction := eye.Sub(burst)
	if direction.LenSqr() == 0 {
		return velocity
	}
	strength := (1 - distance) * exposure * windChargeKnockbackMultiplier
	return velocity.Add(direction.Normalize().Mul(strength))
}

func windChargeBlockExplosionPosition(hit mgl64.Vec3, face cube.Face) mgl64.Vec3 {
	var direction mgl64.Vec3
	switch face {
	case cube.FaceDown:
		direction[1] = -1
	case cube.FaceUp:
		direction[1] = 1
	case cube.FaceNorth:
		direction[2] = -1
	case cube.FaceSouth:
		direction[2] = 1
	case cube.FaceWest:
		direction[0] = -1
	case cube.FaceEast:
		direction[0] = 1
	}
	return hit.Add(direction.Mul(windChargeBlockHitOffset))
}

func windChargeCanHit(e world.Entity) bool {
	return windChargeCanHitType(e.H().Type().EncodeEntity())
}

func windChargeCanHitType(name string) bool {
	return name != "minecraft:wind_charge_projectile" &&
		name != "minecraft:end_crystal" &&
		name != "minecraft:ender_crystal"
}

// WindChargeType is a world.EntityType implementation for WindCharge.
var WindChargeType windChargeType

type windChargeType struct{}

func (windChargeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (windChargeType) EncodeEntity() string { return "minecraft:wind_charge_projectile" }
func (windChargeType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.15625, -0.15, -0.15625, 0.15625, 0.1625, 0.15625)
}

func (windChargeType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = windChargeConf.New()
}
func (windChargeType) EncodeNBT(*world.EntityData) map[string]any { return nil }
