package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// ExperienceOrbBehaviourConfig holds optional parameters for the creation of
// an ExperienceOrbBehaviour.
type ExperienceOrbBehaviourConfig struct {
	// Gravity is the amount of Y velocity subtracted every tick.
	Gravity float64
	// Drag is used to reduce all axes of the velocity every tick. Velocity is
	// multiplied with (1-Drag) every tick.
	Drag float64
	// ExistenceDuration specifies how long the experience orb should last. The
	// default is time.Minute * 5.
	ExistenceDuration time.Duration
	// Experience is the amount of experience held by the orb. Default is 1.
	Experience int
}

// New creates an ExperienceOrbBehaviour using the parameters in conf.
func (conf ExperienceOrbBehaviourConfig) New() *ExperienceOrbBehaviour {
	if conf.Experience == 0 {
		conf.Experience = 1
	}
	if conf.ExistenceDuration == 0 {
		conf.ExistenceDuration = time.Minute * 5
	}
	b := &ExperienceOrbBehaviour{conf: conf, lastSearch: time.Now()}
	b.passive = PassiveBehaviourConfig{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		ExistenceDuration: conf.ExistenceDuration,
		Tick:              b.tick,
	}.New()
	return b
}

// ExperienceOrbBehaviour implements Behaviour for an experience orb entity.
type ExperienceOrbBehaviour struct {
	conf ExperienceOrbBehaviourConfig

	passive *PassiveBehaviour

	lastSearch time.Time
	target     experienceCollector
}

// Experience returns the amount of experience the orb carries.
func (exp *ExperienceOrbBehaviour) Experience() int {
	return exp.conf.Experience
}

// Tick finds a target for the experience orb and moves the orb towards it.
func (exp *ExperienceOrbBehaviour) Tick(e *Ent) *Movement {
	return exp.passive.Tick(e)
}

// followBox is the bounding box used to search for collectors to follow for experience orbs.
var followBox = cube.Box(-8, -8, -8, 8, 8, 8)

// tick finds a target for the experience orb and moves the orb towards it.
func (exp *ExperienceOrbBehaviour) tick(e *Ent) {
	w, pos := e.World(), e.Position()
	if exp.target != nil && (exp.target.Dead() || exp.target.World() != w || pos.Sub(exp.target.Position()).Len() > 8) {
		exp.target = nil
	}

	if time.Since(exp.lastSearch) >= time.Second {
		exp.findTarget(w, pos)
	}
	if exp.target != nil {
		exp.moveToTarget(e)
	}
}

// findTarget attempts to find a target for an experience orb in w around pos.
func (exp *ExperienceOrbBehaviour) findTarget(w *world.World, pos mgl64.Vec3) {
	if exp.target == nil {
		collectors := w.EntitiesWithin(followBox.Translate(pos), func(o world.Entity) bool {
			_, ok := o.(experienceCollector)
			return !ok
		})
		if len(collectors) > 0 {
			exp.target = collectors[0].(experienceCollector)
		}
	}
	exp.lastSearch = time.Now()
}

// moveToTarget applies velocity to the experience orb so that it moves towards
// its current target. If it intersects with the target, the orb is collected.
func (exp *ExperienceOrbBehaviour) moveToTarget(e *Ent) {
	pos, dst := e.Position(), exp.target.Position()
	if o, ok := exp.target.(Eyed); ok {
		dst[1] += o.EyeHeight() / 2
	}
	diff := dst.Sub(pos).Mul(0.125)
	if dist := diff.LenSqr(); dist < 1 {
		e.SetVelocity(e.Velocity().Add(diff.Normalize().Mul(0.2 * math.Pow(1-math.Sqrt(dist), 2))))
	}

	if e.Type().BBox(e).Translate(pos).IntersectsWith(exp.target.Type().BBox(exp.target).Translate(exp.target.Position())) && exp.target.CollectExperience(exp.conf.Experience) {
		_ = e.Close()
	}
}

// experienceCollector represents an entity that can collect experience orbs.
type experienceCollector interface {
	Living
	// CollectExperience makes the player collect the experience points passed,
	// adding it to the experience manager. A bool is returned indicating
	// whether the player was able to collect the experience, or not, due to the
	// 100ms delay between experience collection.
	CollectExperience(value int) bool
}
