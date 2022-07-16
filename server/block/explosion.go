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
	// World is the world in which the explosion should take place.
	World *world.World
	// Pos is the center position of the explosion.
	Pos mgl64.Vec3
	// Size is the size of the explosion, it is effectively the radius which entities/blocks will be affected within.
	Size float64
	// RandSource is the source to use for the explosion "randomness".
	RandSource rand.Source
	// SpawnFire will cause the explosion to randomly start fires in 1/3 of all destroyed air blocks that are
	// above opaque blocks.
	SpawnFire bool

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
	Explode(c ExplosionConfig, impact float64)
}

// Explodable represents a block that can be exploded.
type Explodable interface {
	// Explode is called when an explosion occurs. The block can react to the explosion using the configuration passed.
	Explode(pos cube.Pos, c ExplosionConfig)
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

// Do performs the explosion as specified by the configuration.
func (c ExplosionConfig) Do() {
	if c.Sound == nil {
		c.Sound = sound.Explosion{}
	}
	if c.Particle == nil {
		c.Particle = particle.HugeExplosion{}
	}
	if c.RandSource == nil {
		c.RandSource = rand.NewSource(time.Now().UnixNano())
	}

	r, d := rand.New(c.RandSource), c.Size*2
	box := cube.Box(
		math.Floor(c.Pos[0]-d-1),
		math.Floor(c.Pos[1]-d-1),
		math.Floor(c.Pos[2]-d-1),
		math.Ceil(c.Pos[0]+d+1),
		math.Ceil(c.Pos[1]+d+1),
		math.Ceil(c.Pos[2]+d+1),
	)

	for _, e := range c.World.EntitiesWithin(box.Grow(2), nil) {
		pos := e.Position()
		if !e.BBox().Translate(pos).IntersectsWith(box) {
			continue
		}
		dist := pos.Sub(c.Pos).Len()
		if dist > d {
			continue
		}
		if explodable, ok := e.(ExplodableEntity); ok {
			// TODO: exposure
			explodable.Explode(c, 1-dist/d)
		}
	}

	affectedBlocks := make([]cube.Pos, 0, 32)
	for _, ray := range rays {
		pos := c.Pos
		for blastForce := c.Size * (0.7 + r.Float64()*0.6); blastForce > 0.0; blastForce -= 0.225 {
			current := cube.PosFromVec3(pos)
			if r, ok := c.World.Block(current).(Breakable); ok {
				if blastForce -= (r.BreakInfo().BlastResistance + 0.3) * 0.3; blastForce > 0 {
					affectedBlocks = append(affectedBlocks, current)
				}
			}
			pos = pos.Add(ray)
		}
	}
	for _, pos := range affectedBlocks {
		bl := c.World.Block(pos)
		if explodable, ok := bl.(Explodable); ok {
			explodable.Explode(pos, c)
		} else if breakable, ok := bl.(Breakable); ok {
			c.World.SetBlock(pos, nil, nil)
			if 1/c.Size > r.Float64() {
				for _, drop := range breakable.BreakInfo().Drops(item.ToolNone{}, nil) {
					dropItem(c.World, drop, pos.Vec3Centre())
				}
			}
		}
	}
	if c.SpawnFire {
		for _, pos := range affectedBlocks {
			if r.Intn(3) == 0 {
				if _, ok := c.World.Block(pos).(Air); ok && c.World.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos, cube.FaceUp, c.World) {
					c.World.SetBlock(pos, Fire{}, nil)
				}
			}
		}
	}

	c.World.AddParticle(c.Pos, c.Particle)
	c.World.PlaySound(c.Pos, c.Sound)
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

	double7 := (1.0 - math.Floor(diff[0])/diff[0]) / 2.0
	double8 := (1.0 - math.Floor(diff[2])/diff[2]) / 2.0

	var collisions, checks float64
	for x := 0.0; x <= 1.0; x += step[0] {
		for y := 0.0; y <= 1.0; y += step[1] {
			for z := 0.0; z <= 1.0; z += step[2] {
				dck2 := mgl64.Vec3{
					lerp(x, boxMin[0], boxMax[0]) + double7,
					lerp(y, boxMin[1], boxMax[1]),
					lerp(z, boxMin[2], boxMax[2]) + double8,
				}

				var collides bool
				trace.TraverseBlocks(origin, dck2, func(pos cube.Pos) (con bool) {
					_, air := w.Block(pos).(Air)
					collides = !air
					return air
				})
				if collides {
					collisions++
				}
				checks++
			}
		}
	}
	return collisions / checks
}

// lerp returns the linear interpolation between a and b at t.
func lerp(a, b, t float64) float64 {
	return (1-t)*a + t*b
}
