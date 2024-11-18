package animation

// Animation represents an animation that may be played on an entity from an active resource pack on
// the client.
type Animation struct {
	name          string
	nextState     string
	controller    string
	stopCondition string
}

// New returns a new animation that can be played on an entity. If no controller or stop condition is set,
// the animation will play for its full duration, including looping. Controllers can be set to manage
// multiple states of animations. It is also possible to use vanilla animations/controllers if they work
// for your entity, i.e. "animation.pig.baby_transform".
func New(name string) Animation {
	return Animation{name: name}
}

// Name returns the name of the animation to be played.
func (a Animation) Name() string {
	return a.name
}

// Controller returns the name of the controller to be used for the animation.
func (a Animation) Controller() string {
	return a.controller
}

// WithController returns a copy of the Animation with the provided animation controller. An animation
// controller with the same name must be defined in a resource pack for it to work.
func (a Animation) WithController(controller string) Animation {
	a.controller = controller
	return a
}

// NextState returns the state to transition to after the animation has finished playing within the
// animation controller.
func (a Animation) NextState() string {
	return a.nextState
}

// WithNextState returns a copy of the Animation with the provided state to transition to after the
// animation has finished playing within the animation controller.
func (a Animation) WithNextState(state string) Animation {
	a.nextState = state
	return a
}

// StopCondition returns the condition that must be met for the animation to stop playing. This is often
// a Molang expression that can be used to query various entity properties to determine when the animation
// should stop playing.
func (a Animation) StopCondition() string {
	return a.stopCondition
}

// WithStopCondition returns a copy of the Animation with the provided stop condition. The stop condition
// is a Molang expression that can be used to query various entity properties to determine when the animation
// should stop playing.
func (a Animation) WithStopCondition(condition string) Animation {
	a.stopCondition = condition
	return a
}
