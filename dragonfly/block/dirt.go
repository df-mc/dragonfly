package block

type (
	Dirt       struct{}
	CoarseDirt struct{}
)

func (Dirt) Name() string {
	return "Dirt"
}

func (CoarseDirt) Name() string {
	return "Coarse Dirt"
}
