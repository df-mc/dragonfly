package entity

import (
	"math"
	"time"

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
	windChargeReflectImmunity     = 500 * time.Millisecond
)

// NewWindCharge creates a wind charge entity at a position with an owner
// entity. Wind charges fly in a straight line (no gravity or drag) and create
// a burst of wind on impact.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	if owner != nil {
		conf.Owner = owner.H()
	}
	return opts.New(WindChargeType, conf)
}

var windChargeConf = ProjectileBehaviourConfig{
	Damage:                -1,
	Hit:                   windChargeBurst,
	EntityCollisionFilter: windChargeCanHit,
	Attack:                reflectWindCharge,
	IgnitedByBlocks:       true,
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
			windChargeHitLiving(e, l, owner)
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
		if other.H() == e.H() {
			continue
		}
		moving, ok := other.(interface {
			Velocity() mgl64.Vec3
			SetVelocity(mgl64.Vec3)
		})
		if !ok {
			continue
		}
		// Entities beyond the burst radius are never knocked back, so skip the
		// exposure ray casts that windChargeKnockback would discard anyway.
		if other.Position().Sub(pos).Len() > windChargeExplosionDiameter {
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
		if receiver, ok := other.(interface {
			WindChargeKnockbackMultiplier() float64
		}); ok {
			knockedBack = velocity.Add(knockedBack.Sub(velocity).Mul(receiver.WindChargeKnockbackMultiplier()))
		}
		if knockedBack != velocity {
			applyWindChargeKnockback(other, moving, velocity, knockedBack)
		}
	}

	windChargeAffectBlocks(tx, pos)
}

func applyWindChargeKnockback(other world.Entity, moving interface {
	SetVelocity(mgl64.Vec3)
}, velocity, knockedBack mgl64.Vec3) {
	if receiver, ok := other.(interface {
		NegateFallDamageFromWindCharge(float64)
	}); ok && knockedBack.Y() > velocity.Y() {
		receiver.NegateFallDamageFromWindCharge(other.Position().Y())
	}
	moving.SetVelocity(knockedBack)
}

func windChargeHitLiving(e *Ent, target Living, owner world.Entity) {
	_, vulnerable := target.Hurt(1, ProjectileDamageSource{Projectile: e, Owner: owner})
	if flammable, ok := target.(Flammable); ok && vulnerable && e.OnFireDuration() > 0 {
		flammable.SetOnFire(5 * time.Second)
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
	return hit.Add(cube.Pos{}.Side(face).Vec3().Mul(windChargeBlockHitOffset))
}

func windChargeCanHit(e world.Entity) bool {
	if _, immune := e.H().Type().(WindChargeCollisionImmune); immune {
		return false
	}
	return windChargeCanHitType(e.H().Type().EncodeEntity())
}

func windChargeCanHitType(identifier string) bool {
	switch identifier {
	case "minecraft:end_crystal", "minecraft:ender_crystal",
		"minecraft:wind_charge_projectile", "minecraft:breeze_wind_charge_projectile":
		return false
	default:
		return true
	}
}

func windChargeAffectBlocks(tx *world.Tx, burst mgl64.Vec3) {
	radius := windChargeExplosionPower
	radiusSq := radius * radius
	min := cube.PosFromVec3(burst.Sub(mgl64.Vec3{radius, radius, radius}))
	max := cube.PosFromVec3(burst.Add(mgl64.Vec3{radius, radius, radius}))
	affected := make(map[cube.Pos]struct{})
	for pos := range cube.Range3D(min, max) {
		if pos.Vec3Centre().Sub(burst).LenSqr() > radiusSq {
			continue
		}
		b, ok := tx.Block(pos).(block.WindChargeAffected)
		if !ok {
			continue
		}
		interactionPos := b.WindChargeInteractionPos(pos)
		if _, ok := affected[interactionPos]; ok {
			continue
		}
		affected[interactionPos] = struct{}{}
		// Most blocks interact in place; only multi-block structures such as
		// doors redirect to another position, which must interact using its
		// own block value rather than the one found here.
		if interactionPos == pos {
			b.WindChargeInteract(pos, tx)
			continue
		}
		if target, ok := tx.Block(interactionPos).(block.WindChargeAffected); ok {
			target.WindChargeInteract(interactionPos, tx)
		}
	}
}

func reflectWindCharge(e *Ent, attacker world.Entity) bool {
	if e.Age() < windChargeReflectImmunity {
		return true
	}
	direction := attacker.Rotation().Vec3()
	owner := attacker.H()
	if projectile, ok := attacker.(*Ent); ok {
		if behaviour, ok := projectile.Behaviour().(*ProjectileBehaviour); ok {
			direction = projectile.Velocity()
			owner = behaviour.Owner()
			if owner != nil {
				if ownerEntity, ok := owner.Entity(e.tx); ok {
					direction = ownerEntity.Rotation().Vec3()
				}
			}
		}
	}
	if direction.LenSqr() == 0 {
		return true
	}
	speed := e.Velocity().Len()
	if speed == 0 {
		speed = 1.5
	}
	e.SetVelocity(direction.Normalize().Mul(speed))
	e.Behaviour().(*ProjectileBehaviour).SetOwner(owner)
	return true
}

// WindChargeType is a world.EntityType implementation for WindCharge.
var WindChargeType windChargeType

type windChargeType struct{}

// WindChargeCollisionImmune is implemented by entity types that wind charges
// pass through without bursting.
type WindChargeCollisionImmune interface {
	WindChargeCollisionImmune()
}

func (windChargeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (windChargeType) EncodeEntity() string       { return "minecraft:wind_charge_projectile" }
func (windChargeType) WindChargeCollisionImmune() {}
func (windChargeType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.15625, -0.15, -0.15625, 0.15625, 0.1625, 0.15625)
}

func (windChargeType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = windChargeConf.New()
}
func (windChargeType) EncodeNBT(*world.EntityData) map[string]any { return nil }
