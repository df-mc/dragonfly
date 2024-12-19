package item

type SmithingTemplateType struct {
	smithingTemplateType
}

// TemplateNetheriteUpgrade returns the Netherite Upgrade Template
func TemplateNetheriteUpgrade() SmithingTemplateType {
	return SmithingTemplateType{0}
}

// TemplateSentry returns the Sentry Template.
func TemplateSentry() SmithingTemplateType {
	return SmithingTemplateType{1}
}

// TemplateVex returns the Vex Template.
func TemplateVex() SmithingTemplateType {
	return SmithingTemplateType{2}
}

// TemplateWild returns the Wild Template.
func TemplateWild() SmithingTemplateType {
	return SmithingTemplateType{3}
}

// TemplateCoast returns the Coast Template.
func TemplateCoast() SmithingTemplateType {
	return SmithingTemplateType{4}
}

// TemplateDune returns the Dune Template.
func TemplateDune() SmithingTemplateType {
	return SmithingTemplateType{5}
}

// TemplateWayFinder returns the WayFinder Template.
func TemplateWayFinder() SmithingTemplateType {
	return SmithingTemplateType{6}
}

// TemplateRaiser returns the Raiser Template.
func TemplateRaiser() SmithingTemplateType {
	return SmithingTemplateType{7}
}

// TemplateShaper returns the Shaper Template.
func TemplateShaper() SmithingTemplateType {
	return SmithingTemplateType{8}
}

// TemplateHost returns the Host Template.
func TemplateHost() SmithingTemplateType {
	return SmithingTemplateType{9}
}

// TemplateWard returns the Ward Template.
func TemplateWard() SmithingTemplateType {
	return SmithingTemplateType{10}
}

// TemplateSilence returns the Silence Template.
func TemplateSilence() SmithingTemplateType {
	return SmithingTemplateType{11}
}

// TemplateTide returns the Tide Template.
func TemplateTide() SmithingTemplateType {
	return SmithingTemplateType{12}
}

// TemplateSnout returns the Snout Template.
func TemplateSnout() SmithingTemplateType {
	return SmithingTemplateType{13}
}

// TemplateRib returns the Rib Template.
func TemplateRib() SmithingTemplateType {
	return SmithingTemplateType{14}
}

// TemplateEye returns the Eye Template.
func TemplateEye() SmithingTemplateType {
	return SmithingTemplateType{15}
}

// TemplateSpire returns the Spire Template.
func TemplateSpire() SmithingTemplateType {
	return SmithingTemplateType{16}
}

// TemplateFlow returns the Flow Template.
func TemplateFlow() SmithingTemplateType {
	return SmithingTemplateType{17}
}

// TemplateBolt returns the Bolt Template.
func TemplateBolt() SmithingTemplateType {
	return SmithingTemplateType{18}
}

// SmithingTemplates returns all the ArmourSmithingTemplates
func SmithingTemplates() []SmithingTemplateType {
	return []SmithingTemplateType{
		TemplateNetheriteUpgrade(),
		TemplateSentry(),
		TemplateVex(),
		TemplateWild(),
		TemplateCoast(),
		TemplateDune(),
		TemplateWayFinder(),
		TemplateRaiser(),
		TemplateShaper(),
		TemplateHost(),
		TemplateWard(),
		TemplateSilence(),
		TemplateTide(),
		TemplateSnout(),
		TemplateRib(),
		TemplateEye(),
		TemplateSpire(),
		TemplateFlow(),
		TemplateBolt(),
	}
}

type smithingTemplateType uint8

// String ...
func (s smithingTemplateType) String() string {
	switch s {
	case 0:
		return "netherite_upgrade"
	case 1:
		return "sentry"
	case 2:
		return "vex"
	case 3:
		return "wild"
	case 4:
		return "coast"
	case 5:
		return "dune"
	case 6:
		return "wayfinder"
	case 7:
		return "raiser"
	case 8:
		return "shaper"
	case 9:
		return "host"
	case 10:
		return "ward"
	case 11:
		return "silence"
	case 12:
		return "tide"
	case 13:
		return "snout"
	case 14:
		return "rib"
	case 15:
		return "eye"
	case 16:
		return "spire"
	case 17:
		return "flow"
	case 18:
		return "bolt"
	}

	panic("should never happen")
}

// smithingTemplateFromString returns an armour smithing template based on a string.
func smithingTemplateFromString(name string) (SmithingTemplateType, bool) {
	switch name {
	case "netherite_upgrade":
		return TemplateNetheriteUpgrade(), true
	case "sentry":
		return TemplateSentry(), true
	case "vex":
		return TemplateVex(), true
	case "wild":
		return TemplateWild(), true
	case "coast":
		return TemplateCoast(), true
	case "dune":
		return TemplateDune(), true
	case "wayfinder":
		return TemplateWayFinder(), true
	case "raiser":
		return TemplateRaiser(), true
	case "shaper":
		return TemplateShaper(), true
	case "host":
		return TemplateHost(), true
	case "ward":
		return TemplateWard(), true
	case "silence":
		return TemplateSilence(), true
	case "tide":
		return TemplateTide(), true
	case "snout":
		return TemplateSnout(), true
	case "rib":
		return TemplateRib(), true
	case "eye":
		return TemplateEye(), true
	case "spire":
		return TemplateSpire(), true
	case "flow":
		return TemplateFlow(), true
	case "bolt":
		return TemplateBolt(), true
	default:
		return SmithingTemplateType{}, false
	}
}
