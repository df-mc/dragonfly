package particle

// Custom is custom identifier of the particle, defined in particle description in resource pack.
type Custom struct {
	particle
	Identifier string

	MoLangVariables map[string]any
}
