package particle

// Custom is custom identifier of the particle, defined in particle description in resource pack.
type Custom struct {
	particle
	Identifier string

	MoLangVariables []MoLangVariable
}

// MoLangVariable is implementation of MoLang variable. Thanks, seb.
type MoLangVariable struct {
	Name  string              `json:"name"`
	Value MoLangVariableValue `json:"value"`
}

type MoLangVariableValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}
