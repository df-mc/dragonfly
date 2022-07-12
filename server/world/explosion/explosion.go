package explosion

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
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

type (
	BlockFunc  func(c Config, pos cube.Pos)
	EntityFunc func(c Config, e world.Entity)
)

var (
	DefaultBlockFunc BlockFunc = func(c Config, pos cube.Pos) {
		// TODO
		c.World.SetBlock(pos, nil, nil)
	}
	DefaultEntityFunc EntityFunc = func(c Config, e world.Entity) {
		// TODO: Account for item entities etc.
		living, ok := e.(entity.Living)
		if !ok {
			return
		}
		pos := e.Position()
		diff := pos.Sub(c.Pos)

		impact := (1 - diff.Len()) * exposure(c.Pos, e)

		dmg := math.Floor(((impact*impact+impact)/2)*8*c.Size*2 + 1)

		living.Hurt(dmg, damage.SourceExplosion{})
		living.KnockBack(c.Pos, impact, diff.Normalize().Mul(impact)[0])
	}
)

type Config struct {
	World *world.World
	Pos   mgl64.Vec3
	Size  float64
	Rand  rand.Source
	// TODO: Fire spawning

	BlockFunc  BlockFunc
	EntityFunc EntityFunc
}

func (c Config) Do() {
	if c.Rand == nil {
		c.Rand = rand.NewSource(time.Now().UnixNano())
	}
	r := rand.New(c.Rand)
	if c.BlockFunc == nil {
		c.BlockFunc = DefaultBlockFunc
	}
	if c.EntityFunc == nil {
		c.EntityFunc = DefaultEntityFunc
	}
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
		entityPos := e.Position()
		if !e.BBox().Translate(entityPos).IntersectsWith(bb) {
			continue
		}
		if entityPos.Sub(c.Pos).Len() <= d {
			c.EntityFunc(c, e)
		}
	}

	var affectedBlocks []cube.Pos
	for _, ray := range rays {
		pos := c.Pos
		for blastForce := c.Size * (0.7 + r.Float64()*0.6); blastForce > 0.0; blastForce -= 0.225 {
			current := cube.PosFromVec3(pos)
			bl := c.World.Block(current)
			if r, ok := bl.(interface{ BlastResistance() float64 }); ok {
				if blastForce -= (r.BlastResistance() + 0.3) * 0.3; blastForce > 0 {
					affectedBlocks = append(affectedBlocks, current)
				}
			}
			pos = pos.Add(ray)
		}
	}
	for _, pos := range affectedBlocks {
		c.BlockFunc(c, pos)
	}

	c.World.AddParticle(c.Pos, particle.HugeExplosion{})
	c.World.PlaySound(c.Pos, sound.Explosion{})
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
					_, ok := w.Block(pos).(block.Air)
					collides = !ok
					return ok
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
