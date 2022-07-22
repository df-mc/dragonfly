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

	yaw, pitch float64

	age   int
	close bool

	owner world.Entity

	t potion.Potion
	c *ProjectileComputer
}

// NewLingeringPotion ...
func NewLingeringPotion(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, t potion.Potion) *LingeringPotion {
	l := &LingeringPotion{
		yaw:   yaw,
		pitch: pitch,
		owner: owner,

		splashable: splashable{t: t},
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
	}
	l.transform = newTransform(l, pos)
	return l
}

// Name ...
func (l *LingeringPotion) Name() string {
	return "Lingering Potion"
}

// EncodeEntity ...
func (l *LingeringPotion) EncodeEntity() string {
	return "minecraft:lingering_potion"
}

// BBox ...
func (l *LingeringPotion) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Rotation ...
func (l *LingeringPotion) Rotation() (float64, float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.yaw, l.pitch
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
	m, result := l.c.TickMovement(l, l.pos, l.vel, l.yaw, l.pitch, l.ignores)
	l.pos, l.vel, l.yaw, l.pitch = m.pos, m.vel, m.yaw, m.pitch
	l.mu.Unlock()

	l.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		l.close = true
		return
	}

	if result != nil {
		l.splash(l, w, m.pos, result, l.BBox())
		w.AddEntity(NewDefaultAreaEffectCloud(m.pos, l.splashable.t))

		l.close = true
	}
}

// ignores returns whether the LingeringPotion should ignore collision with the entity passed.
func (l *LingeringPotion) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == l || (l.age < 5 && entity == l.owner)
}

// New creates a LingeringPotion with the position, velocity, yaw, and pitch provided. It doesn't spawn the
// LingeringPotion, only returns it.
func (l *LingeringPotion) New(pos, vel mgl64.Vec3, yaw, pitch float64, t potion.Potion) world.Entity {
	lingering := NewLingeringPotion(pos, yaw, pitch, nil, t)
	lingering.vel = vel
	return lingering
}

// Owner ...
func (l *LingeringPotion) Owner() world.Entity {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner
}

// Own ...
func (l *LingeringPotion) Own(owner world.Entity) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.owner = owner
}

// DecodeNBT decodes the properties in a map to a LingeringPotion and returns a new LingeringPotion entity.
func (l *LingeringPotion) DecodeNBT(data map[string]any) any {
	return l.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.Map[float32](data, "Yaw")),
		float64(nbtconv.Map[float32](data, "Pitch")),
		potion.From(nbtconv.Map[int32](data, "PotionId")),
	)
}

// EncodeNBT encodes the LingeringPotion entity's properties as a map and returns it.
func (l *LingeringPotion) EncodeNBT() map[string]any {
	yaw, pitch := l.Rotation()
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(l.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(l.Velocity()),
		"PotionId": l.t.Uint8(),
		"Yaw":      yaw,
		"Pitch":    pitch,
	}
}
