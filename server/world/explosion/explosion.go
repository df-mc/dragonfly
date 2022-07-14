package explosion

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
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
	Explode(c Config, impact float64)
}

type ExplodableBlock interface {
	BlastResistance() float64
	Explode(pos cube.Pos, c Config)
}

type Config struct {
	World *world.World
	Pos   mgl64.Vec3
	Size  float64
	Rand  rand.Source
	Fire  bool

	Sound    world.Sound
	Particle world.Particle
}

func (c Config) Do() {
	if c.Rand == nil {
		c.Rand = rand.NewSource(time.Now().UnixNano())
	}
	r := rand.New(c.Rand)
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

	var affectedBlocks = make([]cube.Pos, 0, 32)
	for _, ray := range rays {
		pos := c.Pos
		for blastForce := c.Size * (0.7 + r.Float64()*0.6); blastForce > 0.0; blastForce -= 0.225 {
			current := cube.PosFromVec3(pos)
			if r, ok := c.World.Block(current).(interface{ BlastResistance() float64 }); ok {
				if blastForce -= (r.BlastResistance() + 0.3) * 0.3; blastForce > 0 {
					affectedBlocks = append(affectedBlocks, current)
				}
			}
			pos = pos.Add(ray)
		}
	}
	for _, pos := range affectedBlocks {
		if explodable, ok := c.World.Block(pos).(ExplodableBlock); ok {
			explodable.Explode(pos, c)
		}
	}
	if c.Fire {
		f := fire()
		for _, pos := range affectedBlocks {
			if rand.Intn(3) == 0 {
				if c.World.Block(pos) == air() {
					c.World.SetBlock(pos, f, nil)
				}
			}
		}
	}

	if c.Particle != nil {
		c.World.AddParticle(c.Pos, c.Particle)
	}
	if c.Sound != nil {
		c.World.PlaySound(c.Pos, c.Sound)
	}
}

// fire returns a fire block.
func fire() world.Block {
	f, ok := world.BlockByName("minecraft:fire", map[string]any{"age": int32(0)})
	if !ok {
		panic("could not find fire block")
	}
	return f
}

// air returns a air block.
func air() world.Block {
	f, ok := world.BlockByName("minecraft:air", map[string]any{})
	if !ok {
		panic("could not find air block")
	}
	return f
}

// TODO
func exposure(origin mgl64.Vec3, e world.Entity) float64 {
	w := e.World()
	pos := e.Position()
	bb := e.BBox().Translate(pos)
	min, max := bb.Min(), bb.Max()
	diff := max.Sub(min).Mul(2.0).Add(mgl64.Vec3{1, 1, 1})
	double4 := 1.0 / diff[0]
	double5 := 1.0 / diff[1]
	double6 := 1.0 / diff[2]
	double7 := (1.0 - math.Floor(1.0/double4)*double4) / 2.0
	double8 := (1.0 - math.Floor(1.0/double6)*double6) / 2.0
	if double4 < 0.0 || double5 < 0.0 || double6 < 0.0 {
		return 0.0
	}
	integer14 := 0.0
	integer15 := 0.0
	for float16 := 0.0; float16 <= 1.0; float16 += double4 {
		for float17 := 0.0; float17 <= 1.0; float17 += double5 {
			for float18 := 0.0; float18 <= 1.0; float18 += double6 {
				dck2 := mgl64.Vec3{
					lerp(float16, min[0], max[0]) + double7,
					lerp(float17, min[1], max[1]),
					lerp(float18, min[2], max[2]) + double8,
				}
				var collides bool
				trace.TraverseBlocks(dck2, origin, func(pos cube.Pos) (con bool) {
					air := w.Block(pos) == air()
					collides = !air
					return air
				})
				if collides {
					integer14++
				}
				integer15++
			}
		}
	}
	return integer14 / integer15
}

func lerp(v, v1, t float64) float64 {
	return (1-t)*v + t*v1
}
