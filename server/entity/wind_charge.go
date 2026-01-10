package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// NewWindCharge creates a wind charge entity at a position with an owner entity.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	conf.Owner = owner.H()
	return opts.New(WindChargeType, conf)
}

var windChargeConf = ProjectileBehaviourConfig{
	Gravity:               0.00,
	Drag:                  0.01,
	Particle:              particle.WindCharge{},
	SurviveBlockCollision: false,
	Sound:                 sound.WindCharge{},
	Hit:                   windChargeHit,
}

// windChargeHit knocks nearby entities back with a force proportional to the distance from the charge.
func windChargeHit(e *Ent, tx *world.Tx, target trace.Result) {
	pos := target.Position()

	const (
		explosionRadius = 3.0
		knockbackForce  = 0.7
		// Tweakable modifiers. Change these to adjust behavior:
		verticalBias   = 1.7 // higher = more vertical lift
		horizontalDamp = 0.6 // lower = less horizontal displacement
	)

	nearby := tx.EntitiesWithin(e.H().Type().BBox(e).Translate(pos).Grow(explosionRadius))
	for victim := range nearby {
		victimPos := victim.Position()
		delta := victimPos.Sub(pos)
		distance := delta.Len()

		if distance > explosionRadius {
			continue
		}

		var falloff float64
		if distance == 0 {
			// Entity exactly at impact — treat as maximum falloff.
			falloff = 1.0
		} else {
			falloff = 1 - (distance / explosionRadius)
		}

		strength := knockbackForce * falloff

		horizontalForce := strength * horizontalDamp
		verticalHeight := strength * verticalBias

		if l, ok := victim.(Living); ok {
			// If the victim is jumping (either moving upwards or not on ground), boost the vertical
			// knockback so jumping + wind charge launches higher than being hit from the ground alone.
			vh := verticalHeight
			isJumping := false
			if l.Velocity()[1] > 0.05 {
				isJumping = true
			}
			if g, ok := l.(interface{ OnGround() bool }); ok {
				if !g.OnGround() {
					isJumping = true
				}
			}
			if isJumping {
				vh *= 1.5
			}
			l.KnockBack(pos, horizontalForce, vh)
		}

		if setter, ok := victim.(interface{ SetLaunchY(float64) }); ok {
			setter.SetLaunchY(victimPos[1])
		}
	}

	// Interact with blocks (buttons, doors, etc.)
	for x := int(math.Floor(pos[0] - explosionRadius)); x <= int(math.Ceil(pos[0]+explosionRadius)); x++ {
		for y := int(math.Floor(pos[1] - explosionRadius)); y <= int(math.Ceil(pos[1]+explosionRadius)); y++ {
			for z := int(math.Floor(pos[2] - explosionRadius)); z <= int(math.Ceil(pos[2]+explosionRadius)); z++ {
				bpos := cube.Pos{x, y, z}
				if bpos.Vec3().Sub(pos).Len() <= explosionRadius {
					b := tx.Block(bpos)
					switch b := b.(type) {
					case block.WoodDoor:
						b.Open = !b.Open
						tx.SetBlock(bpos, b, nil)
					case block.WoodTrapdoor:
						b.Open = !b.Open
						tx.SetBlock(bpos, b, nil)
					case block.CopperDoor:
						b.Open = !b.Open
						tx.SetBlock(bpos, b, nil)
					case block.CopperTrapdoor:
						b.Open = !b.Open
						tx.SetBlock(bpos, b, nil)
					case block.WoodFenceGate:
						b.Open = !b.Open
						tx.SetBlock(bpos, b, nil)
					}
				}
			}
		}
	}
}

// WindChargeType is a world.EntityType implementation for wind charges.
var WindChargeType windChargeType

type windChargeType struct{}

func (t windChargeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (windChargeType) EncodeEntity() string { return "minecraft:wind_charge_projectile" }
func (windChargeType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3125, 0, -0.3125, 0.3125, 0.3125, 0.3125)
}

func (windChargeType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = windChargeConf.New()
}
func (windChargeType) EncodeNBT(*world.EntityData) map[string]any { return nil }
