package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Snowball ...
type Snowball struct {
	transform
	yaw, pitch float64

	ticksLived int

	owner Living

	c *projectileComputer
}

func NewSnowball(pos mgl64.Vec3, yaw, pitch float64, owner Living) *Snowball {
	s := &Snowball{
		yaw:   yaw,
		pitch: pitch,
		c: &projectileComputer{c: &MovementComputer{
			Gravity:           0.03,
			DragBeforeGravity: true,
			Drag:              0.01,
		}},
		owner: owner,
	}
	s.transform = newTransform(s, pos)

	return s
}

func (s *Snowball) Tick(current int64) {
	var result trace.Result
	s.mu.Lock()
	if s.ticksLived < 5 {
		s.pos, s.vel, s.yaw, s.pitch, result = s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.owner)
	} else {
		s.pos, s.vel, s.yaw, s.pitch, result = s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch)
	}
	pos := s.pos
	s.mu.Unlock()

	if pos[1] < cube.MinY && current%10 == 0 {
		_ = s.Close()
		return
	}

	if result != nil {
		w := s.World()
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.SnowballPoof{})
		}

		_ = s.Close()
	}
}

// Name ...
func (s *Snowball) Name() string {
	return "Snowball"
}

// EncodeEntity ...
func (s *Snowball) EncodeEntity() string {
	return "minecraft:snowball"
}

// AABB ...
func (s *Snowball) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{0.25, 0.25, 0.25})
}
