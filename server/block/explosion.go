package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
	"time"
)

// ExplosionConfig is the configuration for an explosion. The world, position, size, sound, particle, and more can all
// be configured through this configuration.
type ExplosionConfig struct {
	// Size is the size of the explosion, it is effectively the radius which entities/blocks will be affected within.
	Size float64
	// Rand is the source to use for the explosion "randomness".
	Rand rand.Source
	// SpawnFire will cause the explosion to randomly start fires in 1/3 of all destroyed air blocks that are
	// above opaque blocks.
	SpawnFire bool
	// DisableItemDrops, when set to true, will prevent any item entities from dropping as a result of blocks being
	// destroyed.
	DisableItemDrops bool

	// Sound is the sound to play when the explosion is created. If set to nil, this will default to the sound of a
	// regular explosion.
	Sound world.Sound
	// Particle is the particle to spawn when the explosion is created. If set to nil, this will default to the particle
	// of a regular huge explosion.
	Particle world.Particle
}

// ExplodableEntity represents an entity that can be exploded.
type ExplodableEntity interface {
	// Explode is called when an explosion occurs. The entity can then react to the explosion using the configuration
	// and impact provided.
	Explode(explosionPos mgl64.Vec3, impact float64, c ExplosionConfig)
}

// Explodable represents a block that can be exploded.
type Explodable interface {
	// Explode is called when an explosion occurs. The block can react to the explosion using the configuration passed.
	Explode(explosionPos mgl64.Vec3, pos cube.Pos, w *world.World, c ExplosionConfig)
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
func (c ExplosionConfig) Explode(w *world.World, explosionPos mgl64.Vec3) {
	if c.Sound == nil {
		c.Sound = sound.Explosion{}
	}
	if c.Particle == nil {
		c.Particle = particle.HugeExplosion{}
	}
	if c.Rand == nil {
		c.Rand = rand.NewSource(time.Now().UnixNano())
	}
	if c.Size == 0 {
		c.Size = 4
	}

	r, d := rand.New(c.Rand), c.Size*2
	box := cube.Box(
		math.Floor(explosionPos[0]-d-1),
		math.Floor(explosionPos[1]-d-1),
		math.Floor(explosionPos[2]-d-1),
		math.Ceil(explosionPos[0]+d+1),
		math.Ceil(explosionPos[1]+d+1),
		math.Ceil(explosionPos[2]+d+1),
	)

	for _, e := range w.EntitiesWithin(box.Grow(2), nil) {
		pos := e.Position()
		if !e.BBox().Translate(pos).IntersectsWith(box) {
			continue
		}
		dist := pos.Sub(pos).Len()
		if dist >= d {
			continue
		}
		if explodable, ok := e.(ExplodableEntity); ok {
			impact := (1 - dist/d) * exposure(pos, e)
			explodable.Explode(explosionPos, impact, c)
		}
	}

	affectedBlocks := make([]cube.Pos, 0, 32)
	for _, ray := range rays {
		pos := explosionPos
		for blastForce := c.Size * (0.7 + r.Float64()*0.6); blastForce > 0.0; blastForce -= 0.225 {
			current := cube.PosFromVec3(pos)
			if r, ok := w.Block(current).(Breakable); ok {
				if blastForce -= (r.BreakInfo().BlastResistance/5 + 0.3) * 0.3; blastForce > 0 {
					affectedBlocks = append(affectedBlocks, current)
				}
			}
			pos = pos.Add(ray)
		}
	}
	for _, pos := range affectedBlocks {
		bl := w.Block(pos)
		if explodable, ok := bl.(Explodable); ok {
			explodable.Explode(explosionPos, pos, w, c)
		} else if breakable, ok := bl.(Breakable); ok {
			w.SetBlock(pos, nil, nil)
			if !c.DisableItemDrops && 1/c.Size > r.Float64() {
				for _, drop := range breakable.BreakInfo().Drops(item.ToolNone{}, nil) {
					dropItem(w, drop, pos.Vec3Centre())
				}
			}
		}
	}
	if c.SpawnFire {
		for _, pos := range affectedBlocks {
			if r.Intn(3) == 0 {
				if _, ok := w.Block(pos).(Air); ok && w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos, cube.FaceUp, w) {
					w.SetBlock(pos, Fire{}, nil)
				}
			}
		}
	}

	w.AddParticle(explosionPos, c.Particle)
	w.PlaySound(explosionPos, c.Sound)
}

// exposure returns the exposure of an explosion to an entity, used to calculate the impact of an explosion.
func exposure(origin mgl64.Vec3, e world.Entity) float64 {
	w := e.World()
	pos := e.Position()
	box := e.BBox().Translate(pos)

	boxMin, boxMax := box.Min(), box.Max()
	diff := boxMax.Sub(boxMin).Mul(2.0).Add(mgl64.Vec3{1, 1, 1})

	step := mgl64.Vec3{1.0 / diff[0], 1.0 / diff[1], 1.0 / diff[2]}
	if step[0] < 0.0 || step[1] < 0.0 || step[2] < 0.0 {
		return 0.0
	}

	xOffset := (1.0 - math.Floor(diff[0])/diff[0]) / 2.0
	zOffset := (1.0 - math.Floor(diff[2])/diff[2]) / 2.0

	var checks, misses int
	for x := 0.0; x <= 1.0; x += step[0] {
		for y := 0.0; y <= 1.0; y += step[1] {
			for z := 0.0; z <= 1.0; z += step[2] {
				point := mgl64.Vec3{
					lerp(x, boxMin[0], boxMax[0]) + xOffset,
					lerp(y, boxMin[1], boxMax[1]),
					lerp(z, boxMin[2], boxMax[2]) + zOffset,
				}

				var collided bool
				trace.TraverseBlocks(origin, point, func(pos cube.Pos) (con bool) {
					_, air := w.Block(pos).(Air)
					collided = !air
					return air
				})
				if !collided {
					misses++
				}
				checks++
			}
		}
	}
	return float64(misses) / float64(checks)
}

// lerp returns the linear interpolation between a and b at t.
func lerp(a, b, t float64) float64 {
	return b + a*(t-b)
}
