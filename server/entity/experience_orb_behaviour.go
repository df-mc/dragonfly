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

func (conf ExperienceOrbBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
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
	target     *world.EntityHandle
}

// Experience returns the amount of experience the orb carries.
func (exp *ExperienceOrbBehaviour) Experience() int {
	return exp.conf.Experience
}

// Tick finds a target for the experience orb and moves the orb towards it.
func (exp *ExperienceOrbBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	return exp.passive.Tick(e, tx)
}

// followBox is the bounding box used to search for collectors to follow for experience orbs.
var followBox = cube.Box(-8, -8, -8, 8, 8, 8)

// tick finds a target for the experience orb and moves the orb towards it.
func (exp *ExperienceOrbBehaviour) tick(e *Ent, tx *world.Tx) {
	targetEnt, ok := exp.target.Entity(tx)
	target, _ := targetEnt.(experienceCollector)

	pos := e.Position()
	hasTarget := ok && !target.Dead() && pos.Sub(target.Position()).Len() <= 8
	if !hasTarget && time.Since(exp.lastSearch) >= time.Second {
		exp.findTarget(tx, pos)
	} else if hasTarget {
		exp.moveToTarget(e, target)
	}
}

// findTarget attempts to find a target for an experience orb in w around pos.
func (exp *ExperienceOrbBehaviour) findTarget(tx *world.Tx, pos mgl64.Vec3) {
	exp.target = nil
	for o := range tx.EntitiesWithin(followBox.Translate(pos)) {
		if _, ok := o.(experienceCollector); ok {
			exp.target = o.H()
			break
		}
	}
	exp.lastSearch = time.Now()
}

// moveToTarget applies velocity to the experience orb so that it moves towards
// its current target. If it intersects with the target, the orb is collected.
func (exp *ExperienceOrbBehaviour) moveToTarget(e *Ent, target experienceCollector) {
	pos, dst := e.Position(), target.Position()
	if o, ok := target.(Eyed); ok {
		dst[1] += o.EyeHeight() / 2
	}
	diff := dst.Sub(pos).Mul(0.125)
	if dist := diff.LenSqr(); dist < 1 {
		e.SetVelocity(e.Velocity().Add(diff.Normalize().Mul(0.2 * math.Pow(1-math.Sqrt(dist), 2))))
	}

	if e.H().Type().BBox(e).Translate(pos).IntersectsWith(target.H().Type().BBox(target).Translate(target.Position())) && target.CollectExperience(exp.conf.Experience) {
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
