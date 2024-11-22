package block

// OxidationType represents a type of oxidation.
type OxidationType struct {
	oxidation
}

type oxidation uint8

// UnoxidisedOxidation is the normal variant of oxidation.
func UnoxidisedOxidation() OxidationType {
	return OxidationType{0}
}

// ExposedOxidation is the exposed variant of oxidation.
func ExposedOxidation() OxidationType {
	return OxidationType{1}
}

// WeatheredOxidation is the weathered variant of oxidation.
func WeatheredOxidation() OxidationType {
	return OxidationType{2}
}

// OxidisedOxidation is the oxidised variant of oxidation.
func OxidisedOxidation() OxidationType {
	return OxidationType{3}
}

// Uint8 returns the oxidation as a uint8.
func (s oxidation) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s oxidation) Name() string {
	switch s {
	case 0:
		return ""
	case 1:
		return "Exposed"
	case 2:
		return "Weathered"
	case 3:
		return "Oxidized"
	}
	panic("unknown oxidation type")
}

// Decrease attempts to decrease the oxidation level by one. It returns the new oxidation level and if the
// decrease was successful.
func (s oxidation) Decrease() (OxidationType, bool) {
	if s > 0 {
		return OxidationType{s - 1}, true
	}
	return UnoxidisedOxidation(), false
}

// Increase attempts to increase the oxidation level by one. It returns the new oxidation level and if the
// increase was successful.
func (s oxidation) Increase() (OxidationType, bool) {
	if s < 3 {
		return OxidationType{s + 1}, true
	}
	return OxidisedOxidation(), false
}

// String ...
func (s oxidation) String() string {
	switch s {
	case 0:
		return ""
	case 1:
		return "exposed"
	case 2:
		return "weathered"
	case 3:
		return "oxidized"
	}
	panic("unknown oxidation type")
}

// OxidationTypes ...
func OxidationTypes() []OxidationType {
	return []OxidationType{UnoxidisedOxidation(), ExposedOxidation(), WeatheredOxidation(), OxidisedOxidation()}
}
