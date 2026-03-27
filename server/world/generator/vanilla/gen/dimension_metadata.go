package gen

type DimensionMetadata struct {
	MinY            int
	Height          int
	SizeHorizontal  int
	SizeVertical    int
	SeaLevel        int
	AquifersEnabled bool
	OreVeinsEnabled bool
	DefaultBlock    DimensionBlockState
	DefaultFluid    DimensionBlockState
}

type DimensionBlockState struct {
	Name       string
	Properties map[string]string
}
