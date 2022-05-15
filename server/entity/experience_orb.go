package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/slices"
	"math"
	"time"
)

// ExperienceOrb is an entity that carries a varying amount of experience. These can be collected by nearby players, and
// are then added to the player's own experience.
type ExperienceOrb struct {
	transform
	age, xp    int
	lastSearch time.Time
	target     experienceCollector
	c          *MovementComputer
}

// orbSplitSizes contains split sizes used for dropping experience orbs.
var orbSplitSizes = []int{2477, 1237, 617, 307, 149, 73, 37, 17, 7, 3, 1}

// NewExperienceOrbs takes in a position and an amount and automatically splits the amount into multiple orbs, returning
// a slice of the created orbs.
func NewExperienceOrbs(pos mgl64.Vec3, amount int) (orbs []*ExperienceOrb) {
	for amount > 0 {
		size := orbSplitSizes[slices.IndexFunc(orbSplitSizes, func(value int) bool {
			return amount >= value
		})]

		orbs = append(orbs, NewExperienceOrb(pos, size))
		amount -= size
	}
	return
}

// NewExperienceOrb creates a new experience orb and returns it.
func NewExperienceOrb(pos mgl64.Vec3, xp int) *ExperienceOrb {
	o := &ExperienceOrb{
		xp:         xp,
		lastSearch: time.Now(),
		c: &MovementComputer{
			Gravity:           0.04,
			Drag:              0.02,
			DragBeforeGravity: true,
		},
	}
	o.transform = newTransform(o, pos)
	return o
}

// Name ...
func (*ExperienceOrb) Name() string {
	return "Experience Orb"
}

// EncodeEntity ...
func (*ExperienceOrb) EncodeEntity() string {
	return "minecraft:xp_orb"
}

// BBox ...
func (*ExperienceOrb) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Experience returns the amount of experience the orb carries.
func (e *ExperienceOrb) Experience() int {
	return e.xp
}

// experienceCollector represents an entity that can collect experience orbs.
type experienceCollector interface {
	Living
	// CollectExperience makes the player collect the experience points passed, adding it to the experience manager. A bool
	// is returned indicating whether the player was able to collect the experience, or not, due to the 100ms delay between
	// experience collection.
	CollectExperience(value int) bool
}

// followBox is the bounding box used to search for collectors to follow for experience orbs.
var followBox = cube.Box(-8, -8, -8, 8, 8, 8)

// Tick ...
func (e *ExperienceOrb) Tick(w *world.World, current int64) {
	e.mu.Lock()
	m := e.c.TickMovement(e, e.pos, e.vel, 0, 0)
	e.pos, e.vel = m.pos, m.vel
	e.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		_ = e.Close()
		return
	}
	if e.age++; e.age > 6000 {
		_ = e.Close()
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.target != nil && (e.target.Dead() || e.target.World() != w || e.pos.Sub(e.target.Position()).Len() > 8) {
		e.target = nil
	}

	if time.Since(e.lastSearch) >= time.Second {
		if e.target == nil {
			if collectors := w.EntitiesWithin(followBox.Translate(e.pos), func(o world.Entity) bool {
				_, ok := o.(experienceCollector)
				return !ok
			}); len(collectors) > 0 {
				e.target = collectors[0].(experienceCollector)
			}
		}
		e.lastSearch = time.Now()
	}

	if e.target != nil {
		vec := e.target.Position()
		if o, ok := e.target.(Eyed); ok {
			vec[1] += o.EyeHeight() / 2
		}
		vec = vec.Sub(e.pos).Mul(0.125)
		if dist := vec.LenSqr(); dist < 1 {
			e.vel = e.vel.Add(vec.Normalize().Mul(0.2 * math.Pow(1-math.Sqrt(dist), 2)))
		}

		if e.BBox().Translate(e.pos).IntersectsWith(e.target.BBox().Translate(e.target.Position())) && e.target.CollectExperience(e.xp) {
			_ = e.Close()
		}
	}
}

// DecodeNBT decodes the properties in a map to an Item and returns a new Item entity.
func (e *ExperienceOrb) DecodeNBT(data map[string]any) any {
	o := NewExperienceOrb(nbtconv.MapVec3(data, "Pos"), int(nbtconv.Map[int32](data, "Value")))
	o.SetVelocity(nbtconv.MapVec3(data, "Motion"))
	o.age = int(nbtconv.Map[int16](data, "Age"))
	return o
}

// EncodeNBT encodes the Item entity's properties as a map and returns it.
func (e *ExperienceOrb) EncodeNBT() map[string]any {
	return map[string]any{
		"Age":    int16(e.age),
		"Value":  int32(e.xp),
		"Pos":    nbtconv.Vec3ToFloat32Slice(e.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(e.Velocity()),
	}
}
