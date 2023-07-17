package animation

// Animation represents an animation & controller that may be attached to an entity.
// Animations and controllers must be defined in a resource pack
type Animation struct {
	name, state, controller string
	stopCondition           string
}

// New returns a new animation that can be attached to an entity. By default no controller or state is sent to the viewer.
// To add a state and controller use WithController and WithState respectively.
func New(animation string) Animation {
	return Animation{
		name:          animation,
		state:         "",
		controller:    "",
		stopCondition: "",
	}
}

// Name returns the name of the animation to be played
func (a Animation) Name() string {
	return a.name
}

// WithController sets the controller with the specified state.
// The controller must be added in a resource pack
func (a Animation) WithController(controller string) Animation {
	a.controller = controller
	return a
}

// Controller returns the name of the controller being used. Controller returns an empty string if
// no controller was previously set
func (a Animation) Controller() string {
	return a.controller
}

// WithState sets the state to transition to as defined in the controller.
func (a Animation) WithState(state string) Animation {
	a.state = state
	return a
}

// State returns the current state being played. State returns an empty string if
// no controller was previously set
func (a Animation) State() string {
	return a.state
}

// WithStopCondition takes the molang expression and stops the animation if the query passes.
func (a Animation) WithStopCondition(condition string) Animation {
	a.stopCondition = condition
	return a
}

// StopCondition returns the stop condition. StopCondition returns an empty string if
// no molang expression was set
func (a Animation) StopCondition() string {
	return a.stopCondition
}
