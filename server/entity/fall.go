package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"sync"
)

// FallManager handles entities that can fall.
type FallManager struct {
	mu           sync.Mutex
	e            fallEntity
	fallDistance float64
	onGround     bool
}

// fallEntity is an entity that can fall.
type fallEntity interface {
	world.Entity
	OnGround() bool
}

// entityLander represents a block that reacts to an entity landing on it after falling.
type entityLander interface {
	// EntityLand is called when an entity lands on the block.
	EntityLand(pos cube.Pos, w *world.World, e world.Entity)
}

// NewFallManager returns a new fall manager.
func NewFallManager(e fallEntity) *FallManager {
	return &FallManager{e: e}
}

// SetFallDistance sets the fall distance of the entity.
func (f *FallManager) SetFallDistance(distance float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallDistance = distance
}

// FallDistance returns the entity's fall distance.
func (f *FallManager) FallDistance() float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fallDistance
}

// ResetFallDistance resets the player's fall distance.
func (f *FallManager) ResetFallDistance() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallDistance = 0
}

// OnGround returns whether the entity is currently considered to be on the ground.
func (f *FallManager) OnGround() bool {
	if false { // todo: if f.e is a player
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.onGround
	}
	return f.e.OnGround()
}

// UpdateFallState is called to update the entities falling state.
func (f *FallManager) UpdateFallState(distanceThisTick float64) {
	f.mu.Lock()
	fallDistance := f.fallDistance
	f.mu.Unlock()
	if f.OnGround() {
		if fallDistance > 0 {
			f.fall(fallDistance)
			f.ResetFallDistance()
		}
	} else if distanceThisTick < fallDistance {
		f.mu.Lock()
		f.fallDistance -= distanceThisTick
		f.mu.Unlock()
	} else {
		f.ResetFallDistance()
	}
}

// fall is called when a falling entity hits the ground.
func (f *FallManager) fall(distance float64) {
	var (
		w   = f.e.World()
		pos = cube.PosFromVec3(f.e.Position())
		b   = w.Block(pos)
		dmg = distance - 3
	)
	if len(b.Model().BBox(pos, w)) == 0 {
		pos = pos.Sub(cube.Pos{0, 1})
		b = w.Block(pos)
	}
	if h, ok := b.(entityLander); ok {
		h.EntityLand(pos, w, f.e)
	}

	if p, ok := f.e.(Living); ok {
		if boost, ok := p.Effect(effect.JumpBoost{}); ok {
			dmg -= float64(boost.Level())
		}
		if dmg < 0.5 {
			return
		}
		p.Hurt(math.Ceil(dmg), damage.SourceFall{})
	}
}

// CheckOnGround checks if the entity is currently considered to be on the ground.
func (f *FallManager) CheckOnGround(w *world.World) bool {
	box := f.e.BBox().Translate(f.e.Position())

	b := box.Grow(1)

	min, max := cube.PosFromVec3(b.Min()), cube.PosFromVec3(b.Max())
	for x := min[0]; x <= max[0]; x++ {
		for z := min[2]; z <= max[2]; z++ {
			for y := min[1]; y < max[1]; y++ {
				pos := cube.Pos{x, y, z}
				boxList := w.Block(pos).Model().BBox(pos, w)
				for _, bb := range boxList {
					if bb.GrowVec3(mgl64.Vec3{0, 0.05}).Translate(pos.Vec3()).IntersectsWith(box) {
						f.mu.Lock()
						f.onGround = true
						f.mu.Unlock()
						return true
					}
				}
			}
		}
	}
	f.mu.Lock()
	f.onGround = false
	f.mu.Unlock()
	return false
}
