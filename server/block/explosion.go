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

var rays = make([]mgl64.Vec3, 0, 1352)

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

type ExplodableEntity interface {
	Explode(c ExplosionConfig, impact float64)
}

type Explodable interface {
	Explode(pos cube.Pos, c ExplosionConfig)
}

type ExplosionConfig struct {
	// World is the world that the explosion will take place.
	World *world.World
	// Pos is the position in the world that the explosion will take place at.
	Pos mgl64.Vec3
	// Size ...
	Size float64

	// SpawnFire will cause the explosion to randomly start fires in 1/3 of all destroyed air blocks that are
	// above opaque blocks.
	SpawnFire bool

	Sound    world.Sound
	Particle world.Particle

	RandSource rand.Source
}

// Do ...
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
	r := rand.New(c.RandSource)

	d := c.Size * 2
	bb := cube.Box(
		math.Floor(c.Pos[0]-d-1),
		math.Ceil(c.Pos[0]+d+1),
		math.Floor(c.Pos[1]-d-1),
		math.Ceil(c.Pos[1]+d+1),
		math.Floor(c.Pos[2]-d-1),
		math.Ceil(c.Pos[2]+d+1),
	)

	for _, e := range c.World.EntitiesWithin(bb.Grow(2), nil) {
		pos := e.Position()
		if !e.BBox().Translate(pos).IntersectsWith(bb) {
			continue
		}
		dist := pos.Sub(c.Pos).Len()
		if dist > d {
			continue
		}
		if explodable, ok := e.(ExplodableEntity); ok {
			explodable.Explode(c, (1-dist/d)*exposure(c.Pos, e))
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
				if _, ok := c.World.Block(pos).(Air); ok && !c.World.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos, cube.FaceUp, c.World) {
					c.World.SetBlock(pos, Fire{}, nil)
				}
			}
		}
	}

	c.World.AddParticle(c.Pos, c.Particle)
	c.World.PlaySound(c.Pos, c.Sound)
}

// TODO
func exposure(origin mgl64.Vec3, e world.Entity) float64 {
	w := e.World()
	pos := e.Position()
	bb := e.BBox().Translate(pos)
	min, max := bb.Min(), bb.Max()
	diff := max.Sub(min).Mul(2.0).Add(mgl64.Vec3{1, 1, 1})
	step := mgl64.Vec3{1.0 / diff[0], 1.0 / diff[1], 1.0 / diff[2]}
	if step[0] < 0.0 || step[1] < 0.0 || step[2] < 0.0 {
		return 0.0
	}
	double7 := (1.0 - math.Floor(diff[0])/diff[0]) / 2.0
	double8 := (1.0 - math.Floor(diff[2])/diff[2]) / 2.0
	collisions := 0.0
	checks := 0.0
	for x := 0.0; x <= 1.0; x += step[0] {
		for y := 0.0; y <= 1.0; y += step[1] {
			for z := 0.0; z <= 1.0; z += step[2] {
				dck2 := mgl64.Vec3{
					lerp(x, min[0], max[0]) + double7,
					lerp(y, min[1], max[1]),
					lerp(z, min[2], max[2]) + double8,
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

func lerp(v, v1, t float64) float64 {
	return (1-t)*v + t*v1
}
