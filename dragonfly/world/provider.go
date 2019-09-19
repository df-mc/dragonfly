package world

// Provider represents a value that may provide world data to a World value. It usually does the reading and
// writing of the world data so that the World may use it.
type Provider interface {
	// WorldName returns the name of the world that the provider provides for. When setting the provider of a
	// World, the World will replace its current name with this one.
	WorldName() string
}

// NoIOProvider implements a Provider while not performing any disk I/O. It generates values on the run and
// dynamically, instead of reading and writing data.
type NoIOProvider struct{}

// WorldName ...
func (p NoIOProvider) WorldName() string {
	return "World"
}
