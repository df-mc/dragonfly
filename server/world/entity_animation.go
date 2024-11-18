package world

// EntityAnimation represents an animation that may be played on an entity from an active resource pack on
// the client.
type EntityAnimation struct {
	name          string
	nextState     string
	controller    string
	stopCondition string
}

// NewEntityAnimation returns a new animation that can be played on an entity. If no controller or stop
// condition is set, the animation will play for its full duration, including looping. Controllers can be set
// to manage multiple states of animations. It is also possible to use vanilla animations/controllers if they
// work for your entity, i.e. "animation.pig.baby_transform".
func NewEntityAnimation(name string) EntityAnimation {
	return EntityAnimation{name: name}
}

// Name returns the name of the animation to be played.
func (a EntityAnimation) Name() string {
	return a.name
}

// Controller returns the name of the controller to be used for the animation.
func (a EntityAnimation) Controller() string {
	return a.controller
}

// WithController returns a copy of the EntityAnimation with the provided animation controller. An animation
// controller with the same name must be defined in a resource pack for it to work.
func (a EntityAnimation) WithController(controller string) EntityAnimation {
	a.controller = controller
	return a
}

// NextState returns the state to transition to after the animation has finished playing within the
// animation controller.
func (a EntityAnimation) NextState() string {
	return a.nextState
}

// WithNextState returns a copy of the EntityAnimation with the provided state to transition to after the
// animation has finished playing within the animation controller.
func (a EntityAnimation) WithNextState(state string) EntityAnimation {
	a.nextState = state
	return a
}

// StopCondition returns the condition that must be met for the animation to stop playing. This is often
// a Molang expression that can be used to query various entity properties to determine when the animation
// should stop playing.
func (a EntityAnimation) StopCondition() string {
	return a.stopCondition
}

// WithStopCondition returns a copy of the EntityAnimation with the provided stop condition. The stop condition
// is a Molang expression that can be used to query various entity properties to determine when the animation
// should stop playing.
func (a EntityAnimation) WithStopCondition(condition string) EntityAnimation {
	a.stopCondition = condition
	return a
}
