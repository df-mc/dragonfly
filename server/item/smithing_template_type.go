package item

type ArmourSmithingTemplate struct {
	smithingTemplateType
}

// TemplateNetheriteUpgrade returns the Netherite Upgrade Template
func TemplateNetheriteUpgrade() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{0}
}

// TemplateSentry returns the Sentry Template.
func TemplateSentry() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{1}
}

// TemplateVex returns the Vex Template.
func TemplateVex() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{2}
}

// TemplateWild returns the Wild Template.
func TemplateWild() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{3}
}

// TemplateCoast returns the Coast Template.
func TemplateCoast() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{4}
}

// TemplateDune returns the Dune Template.
func TemplateDune() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{5}
}

// TemplateWayFinder returns the WayFinder Template.
func TemplateWayFinder() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{6}
}

// TemplateRaiser returns the Raiser Template.
func TemplateRaiser() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{7}
}

// TemplateShaper returns the Shaper Template.
func TemplateShaper() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{8}
}

// TemplateHost returns the Host Template.
func TemplateHost() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{9}
}

// TemplateWard returns the Ward Template.
func TemplateWard() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{10}
}

// TemplateSilence returns the Silence Template.
func TemplateSilence() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{11}
}

// TemplateTide returns the Tide Template.
func TemplateTide() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{12}
}

// TemplateSnout returns the Snout Template.
func TemplateSnout() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{13}
}

// TemplateRib returns the Rib Template.
func TemplateRib() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{14}
}

// TemplateEye returns the Eye Template.
func TemplateEye() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{15}
}

// TemplateSpire returns the Spire Template.
func TemplateSpire() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{16}
}

// TemplateFlow returns the Flow Template.
func TemplateFlow() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{17}
}

// TemplateBolt returns the Bolt Template.
func TemplateBolt() ArmourSmithingTemplate {
	return ArmourSmithingTemplate{18}
}

// SmithingTemplates returns all the ArmourSmithingTemplates
func SmithingTemplates() []ArmourSmithingTemplate {
	return []ArmourSmithingTemplate{
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

// Uint8 returns the uint8 value of the smithing template type.
func (s smithingTemplateType) Uint8() uint8 {
	return uint8(s)
}

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

// ArmourSmithingTemplateFromString returns an armour smithing template based on a string.
func ArmourSmithingTemplateFromString(name string) ArmourSmithingTemplate {
	switch name {
	case "netherite_upgrade":
		return TemplateNetheriteUpgrade()
	case "sentry":
		return TemplateSentry()
	case "vex":
		return TemplateVex()
	case "wild":
		return TemplateWild()
	case "coast":
		return TemplateCoast()
	case "dune":
		return TemplateDune()
	case "wayfinder":
		return TemplateWayFinder()
	case "raiser":
		return TemplateRaiser()
	case "shaper":
		return TemplateShaper()
	case "host":
		return TemplateHost()
	case "ward":
		return TemplateWard()
	case "silence":
		return TemplateSilence()
	case "tide":
		return TemplateTide()
	case "snout":
		return TemplateSnout()
	case "rib":
		return TemplateRib()
	case "eye":
		return TemplateEye()
	case "spire":
		return TemplateSpire()
	case "flow":
		return TemplateFlow()
	case "bolt":
		return TemplateBolt()
	}

	panic("unknown template type")
}
