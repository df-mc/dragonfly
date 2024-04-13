package item

type ArmourTrimTemplate struct {
	Name string
}

// TemplateNetheriteUpgrade returns the Netherrite Upgrade Template
func TemplateNetheriteUpgrade() ArmourTrimTemplate {
	return ArmourTrimTemplate{"netherite_upgrade"}
}

// TemplateSentry returns the Sentry Template.
func TemplateSentry() ArmourTrimTemplate {
	return ArmourTrimTemplate{"sentry"}
}

// TemplateVex returns the Vex Template.
func TemplateVex() ArmourTrimTemplate {
	return ArmourTrimTemplate{"vex"}
}

// TemplateWild returns the Wild Template.
func TemplateWild() ArmourTrimTemplate {
	return ArmourTrimTemplate{"wild"}
}

// TemplateCoast returns the Coast Template.
func TemplateCoast() ArmourTrimTemplate {
	return ArmourTrimTemplate{"coast"}
}

// TemplateDune returns the Dune Template.
func TemplateDune() ArmourTrimTemplate {
	return ArmourTrimTemplate{"dune"}
}

// TemplateWayFinder returns the WayFinder Template.
func TemplateWayFinder() ArmourTrimTemplate {
	return ArmourTrimTemplate{"wayfinder"}
}

// TemplateRaiser returns the Raiser Template.
func TemplateRaiser() ArmourTrimTemplate {
	return ArmourTrimTemplate{"raiser"}
}

// TemplateShaper returns the Raiser Template.
func TemplateShaper() ArmourTrimTemplate {
	return ArmourTrimTemplate{"shaper"}
}

// TemplateHost returns the Host Template.
func TemplateHost() ArmourTrimTemplate {
	return ArmourTrimTemplate{"host"}
}

// TemplateWard returns the Ward Template.
func TemplateWard() ArmourTrimTemplate {
	return ArmourTrimTemplate{"ward"}
}

// TemplateSilence returns the Silence Template.
func TemplateSilence() ArmourTrimTemplate {
	return ArmourTrimTemplate{"silence"}
}

// TemplateTide returns the Tide Template.
func TemplateTide() ArmourTrimTemplate {
	return ArmourTrimTemplate{"tide"}
}

// TemplateSnout returns the Snout Template.
func TemplateSnout() ArmourTrimTemplate {
	return ArmourTrimTemplate{"snout"}
}

// TemplateRib returns the Rib Template.
func TemplateRib() ArmourTrimTemplate {
	return ArmourTrimTemplate{"rib"}
}

// TemplateEye returns the Eye Template.
func TemplateEye() ArmourTrimTemplate {
	return ArmourTrimTemplate{"eye"}
}

// TemplateSpire returns the Spire Template.
func TemplateSpire() ArmourTrimTemplate {
	return ArmourTrimTemplate{"spire"}
}

// Templates returns all the Templates
func Templates() []ArmourTrimTemplate {
	return []ArmourTrimTemplate{TemplateSentry(), TemplateVex(), TemplateWild(), TemplateCoast(), TemplateDune(), TemplateWayFinder(), TemplateRaiser(), TemplateShaper(), TemplateHost(), TemplateWard(), TemplateSilence(), TemplateTide(), TemplateSnout(), TemplateRib(), TemplateEye(), TemplateSpire()}
}

// TemplateFromString returns a template based on a string.
func TemplateFromString(name string) ArmourTrimTemplate {
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
	case "eye":
		return TemplateEye()
	case "spire":
		return TemplateSpire()
	}

	panic("unknown template type")
}
