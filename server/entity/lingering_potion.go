package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// LingeringPotion is a variant of a splash potion that can be thrown to leave clouds with status effects that linger on
// the ground in an area.
type LingeringPotion struct {
	splashable
	transform

	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewLingeringPotion ...
func NewLingeringPotion(pos mgl64.Vec3, owner world.Entity, t potion.Potion) *LingeringPotion {
	l := &LingeringPotion{
		owner:      owner,
		splashable: splashable{t: t, m: 0.25},
		c:          newProjectileComputer(0.05, 0.01),
	}
	l.transform = newTransform(l, pos)
	return l
}

// Type returns LingeringPotionType.
func (*LingeringPotion) Type() world.EntityType {
	return LingeringPotionType{}
}

// Lingers always returns true.
func (l *LingeringPotion) Lingers() bool {
	return true
}

// Tick ...
func (l *LingeringPotion) Tick(w *world.World, current int64) {
	if l.close {
		_ = l.Close()
		return
	}
	l.mu.Lock()
	m, result := l.c.TickMovement(l, l.pos, l.vel, 0, 0)
	l.pos, l.vel = m.pos, m.vel
	l.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		l.close = true
		return
	}

	if result != nil {
		l.splash(l, w, m.pos, result, l.Type().BBox(l))
		w.AddEntity(NewAreaEffectCloud(m.pos, l.t))

		l.close = true
	}
}

// New creates a LingeringPotion with the position and velocity provided. It doesn't spawn the
// LingeringPotion, only returns it.
func (l *LingeringPotion) New(pos, vel mgl64.Vec3, t potion.Potion, owner world.Entity) world.Entity {
	lingering := NewLingeringPotion(pos, nil, t)
	lingering.vel = vel
	lingering.owner = owner
	return lingering
}

// Owner ...
func (l *LingeringPotion) Owner() world.Entity {
	return l.owner
}

// LingeringPotionType is a world.EntityType implementation for LingeringPotion.
type LingeringPotionType struct{}

func (LingeringPotionType) EncodeEntity() string {
	return "minecraft:lingering_potion"
}
func (LingeringPotionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (LingeringPotionType) DecodeNBT(m map[string]any) world.Entity {
	pot := NewLingeringPotion(nbtconv.Vec3(m, "Pos"), nil, potion.From(nbtconv.Int32(m, "PotionId")))
	pot.vel = nbtconv.Vec3(m, "Motion")
	return pot
}

func (LingeringPotionType) EncodeNBT(e world.Entity) map[string]any {
	pot := e.(*LingeringPotion)
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(pot.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(pot.Velocity()),
		"PotionId": pot.t.Uint8(),
	}
}
