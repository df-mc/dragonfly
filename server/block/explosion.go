package block

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// ExplosionConfig is the configuration for an explosion. The sound, particle, item drop chance and more can all be
// configured through this configuration. The position and size come from the world.ExplosionSource passed to Explode.
type ExplosionConfig struct {
	// RandSource is the source to use for the explosion "randomness". If set
	// to nil, RandSource defaults to a `rand.PCG`source seeded with
	// `time.Now().UnixNano()`.
	RandSource rand.Source
	// SpawnFire will cause the explosion to randomly start fires in 1/3 of all destroyed air blocks that are
	// above opaque blocks.
	SpawnFire bool
	// SuppressUnderwaterImpact prevents the explosion from affecting entities through liquid layers. Bedrock Edition
	// applies this to every explosion.
	SuppressUnderwaterImpact bool
	// ItemDropChance specifies how item drops should be handled. By default,
	// the item drop chance is 1/Size. If negative, no items will be dropped by
	// the explosion. If set to 1 or higher, all items are dropped.
	ItemDropChance float64

	// Sound is the sound to play when the explosion is created. If set to nil, this will default to the sound of a
	// regular explosion.
	Sound world.Sound
	// Particle is the particle to spawn when the explosion is created. If set to nil, this will default to the particle
	// of a regular huge explosion.
	Particle world.Particle
}

// ExplodableEntity represents an entity that can be exploded.
type ExplodableEntity interface {
	// Explode is called when an explosion occurs. The entity can react using the source and impact provided.
	Explode(src world.ExplosionSource, impact float64)
}

// Explodable represents a block that can be exploded.
type Explodable interface {
	// Explode is called when an explosion occurs. The block can react using the source passed.
	Explode(src world.ExplosionSource, pos cube.Pos, tx *world.Tx)
}

// rays ...
var rays = make([]mgl64.Vec3, 0, 1352)

// init ...
func init() {
	for x := 0.0; x < 16; x++ {
		for y := 0.0; y < 16; y++ {
			for z := 0.0; z < 16; z++ {
				if x != 0 && x != 15 && y != 0 && y != 15 && z != 0 && z != 15 {
					continue
				}
				rays = append(rays, mgl64.Vec3{x/15*2 - 1, y/15*2 - 1, z/15*2 - 1}.Normalize().Mul(0.3))
			}
		}
	}
}

// Explode performs the explosion as specified by the configuration.
func (c ExplosionConfig) Explode(tx *world.Tx, src world.ExplosionSource) {
	if c.Sound == nil {
		c.Sound = sound.Explosion{}
	}
	if c.Particle == nil {
		c.Particle = particle.HugeExplosion{}
	}
	if c.RandSource == nil {
		t := uint64(time.Now().UnixNano())
		c.RandSource = rand.NewPCG(t, t)
	}
	size, explosionPos := src.Size(), src.Position()
	if c.ItemDropChance == 0 {
		c.ItemDropChance = 1.0 / size
	}

	r, d := rand.New(c.RandSource), size*2
	box := cube.Box(
		math.Floor(explosionPos[0]-d-1),
		math.Floor(explosionPos[1]-d-1),
		math.Floor(explosionPos[2]-d-1),
		math.Ceil(explosionPos[0]+d+1),
		math.Ceil(explosionPos[1]+d+1),
		math.Ceil(explosionPos[2]+d+1),
	)

	affectedEntities := make([]world.Entity, 0, 32)
	for e := range tx.EntitiesWithin(box.Grow(2)) {
		pos := e.Position()
		dist := pos.Sub(explosionPos).Len()
		if dist > d || dist == 0 {
			continue
		}

		affectedEntities = append(affectedEntities, e)
	}

	affectedBlocks, seen := make([]cube.Pos, 0, 32), make(map[cube.Pos]struct{}, 32)
	for _, ray := range rays {
		pos := explosionPos
		for blastForce := size * (0.7 + r.Float64()*0.6); blastForce > 0.0; blastForce -= 0.225 {
			current := cube.PosFromVec3(pos)
			currentBlock := tx.Block(current)

			resistance, resists := 0.0, false
			if l, ok := tx.Liquid(current); ok {
				resistance, resists = l.BlastResistance(), true
			} else if i, ok := currentBlock.(Breakable); ok {
				resistance, resists = i.BreakInfo().BlastResistance, true
			} else if _, ok = currentBlock.(Air); !ok {
				// Completely stop the ray if the current block is not air and unbreakable.
				break
			}

			pos = pos.Add(ray)
			// Air offers no resistance to the ray, only blocks and liquids reduce its force beyond the step decay.
			if resists {
				blastForce -= (resistance + 0.3) * 0.3
			}
			if blastForce > 0 {
				if _, ok := seen[current]; !ok {
					seen[current] = struct{}{}
					affectedBlocks = append(affectedBlocks, current)
				}
			}
		}
	}

	ctx := tx.Event()
	spawnFire := c.SpawnFire
	itemDropChance := c.ItemDropChance
	if tx.World().Handler().HandleExplosion(ctx, src, &affectedEntities, &affectedBlocks, &itemDropChance, &spawnFire); ctx.Cancelled() {
		return
	}

	for _, e := range affectedEntities {
		explodable, ok := e.(ExplodableEntity)
		if !ok {
			continue
		}
		impact := (1 - e.Position().Sub(explosionPos).Len()/d) * c.exposure(tx, explosionPos, e)
		if c.SuppressUnderwaterImpact && impact <= 0 {
			// The blast never reached the entity. Skip the call entirely, as entities with a constant damage term,
			// such as players, would otherwise still be hurt through the liquid that blocked it.
			continue
		}
		explodable.Explode(src, impact)
	}

	for _, pos := range affectedBlocks {
		bl := tx.Block(pos)
		if explodable, ok := bl.(Explodable); ok {
			explodable.Explode(src, pos, tx)
		} else if breakable, ok := bl.(Breakable); ok {
			// Clear the block first so break handlers see the post-break world, this is required by things such as redstone updates.
			tx.SetBlock(pos, nil, nil)
			breakHandler := breakable.BreakInfo().BreakHandler
			if breakHandler != nil {
				breakHandler(pos, tx, nil)
			}
			if itemDropChance > r.Float64() {
				for _, drop := range breakable.BreakInfo().Drops(item.ToolNone{}, nil) {
					dropItem(tx, drop, pos.Vec3Centre())
				}
			}
		}
	}

	if spawnFire {
		for _, pos := range affectedBlocks {
			if r.IntN(3) == 0 {
				if _, ok := tx.Block(pos).(Air); ok && tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos, cube.FaceUp, tx) {
					Fire{}.Start(tx, pos)
				}
			}
		}
	}

	tx.AddParticle(explosionPos, c.Particle)
	tx.PlaySound(explosionPos, c.Sound)
}

// exposure returns the exposure of an explosion to an entity, used to calculate the impact of an explosion.
func (c ExplosionConfig) exposure(tx *world.Tx, origin mgl64.Vec3, e world.Entity) float64 {
	pos := e.Position()
	box := e.H().Type().BBox(e).Translate(pos)

	boxMin, boxMax := box.Min(), box.Max()
	diff := boxMax.Sub(boxMin).Mul(2.0).Add(mgl64.Vec3{1, 1, 1})

	step := mgl64.Vec3{1.0 / diff[0], 1.0 / diff[1], 1.0 / diff[2]}
	if step[0] < 0.0 || step[1] < 0.0 || step[2] < 0.0 {
		return 0.0
	}

	xOffset := (1.0 - math.Floor(diff[0])/diff[0]) / 2.0
	zOffset := (1.0 - math.Floor(diff[2])/diff[2]) / 2.0

	var checks, misses float64
	for x := 0.0; x <= 1.0; x += step[0] {
		for y := 0.0; y <= 1.0; y += step[1] {
			for z := 0.0; z <= 1.0; z += step[2] {
				point := mgl64.Vec3{
					lerp(x, boxMin[0], boxMax[0]) + xOffset,
					lerp(y, boxMin[1], boxMax[1]),
					lerp(z, boxMin[2], boxMax[2]) + zOffset,
				}
				var collided bool
				trace.TraverseBlocks(origin, point, func(pos cube.Pos) (cont bool) {
					if c.SuppressUnderwaterImpact {
						if _, liquid := tx.Liquid(pos); liquid {
							collided = true
							return false
						}
					}
					_, collided = trace.BlockIntercept(pos, tx, tx.Block(pos), origin, point)
					return !collided
				})

				if !collided {
					misses++
				}
				checks++
			}
		}
	}
	return misses / checks
}

// lerp returns the linear interpolation between a and b at t.
func lerp(a, b, t float64) float64 {
	return b + a*(t-b)
}
