package block

type (
	Dirt       struct{}
	CoarseDirt struct{}
)

// Name ...
func (Dirt) Name() string {
	return "Dirt"
}

// Name ...
func (CoarseDirt) Name() string {
	return "Coarse Dirt"
}
