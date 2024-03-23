package item

type TemplateType struct {
	Name string
}

// TemplateNetheriteUpgrade returns the Netherrite Upgrade Template
func TemplateNetheriteUpgrade() TemplateType {
	return TemplateType{Name: "netherite_upgrade"}
}

// TemplateSentry returns the Sentry Template.
func TemplateSentry() TemplateType {
	return TemplateType{"sentry"}
}

// TemplateVex returns the Vex Template.
func TemplateVex() TemplateType {
	return TemplateType{"vex"}
}

// TemplateWild returns the Wild Template.
func TemplateWild() TemplateType {
	return TemplateType{"wild"}
}

// TemplateCoast returns the Coast Template.
func TemplateCoast() TemplateType {
	return TemplateType{"coast"}
}

// TemplateDune returns the Dune Template.
func TemplateDune() TemplateType {
	return TemplateType{"dune"}
}

// TemplateWayFinder returns the WayFinder Template.
func TemplateWayFinder() TemplateType {
	return TemplateType{"wayfinder"}
}

// TemplateRaiser returns the Raiser Template.
func TemplateRaiser() TemplateType {
	return TemplateType{"raiser"}
}

// TemplateShaper returns the Raiser Template.
func TemplateShaper() TemplateType {
	return TemplateType{"shaper"}
}

// TemplateHost returns the Host Template.
func TemplateHost() TemplateType {
	return TemplateType{"host"}
}

// TemplateWard returns the Ward Template.
func TemplateWard() TemplateType {
	return TemplateType{"ward"}
}

// TemplateSilence returns the Silence Template.
func TemplateSilence() TemplateType {
	return TemplateType{"silence"}
}

// TemplateTide returns the Tide Template.
func TemplateTide() TemplateType {
	return TemplateType{"tide"}
}

// TemplateSnout returns the Snout Template.
func TemplateSnout() TemplateType {
	return TemplateType{"snout"}
}

// TemplateRib returns the Rib Template.
func TemplateRib() TemplateType {
	return TemplateType{"rib"}
}

// TemplateEye returns the Eye Template.
func TemplateEye() TemplateType {
	return TemplateType{"eye"}
}

// TemplateSpire returns the Spire Template.
func TemplateSpire() TemplateType {
	return TemplateType{"spire"}
}

// Templates returns all the Templates
func Templates() []TemplateType {
	return []TemplateType{TemplateNetheriteUpgrade(), TemplateSentry(), TemplateVex(), TemplateWild(), TemplateCoast(), TemplateDune(), TemplateWayFinder(), TemplateRaiser(), TemplateShaper(), TemplateHost(), TemplateWard(), TemplateSilence(), TemplateTide(), TemplateSnout(), TemplateRib(), TemplateEye(), TemplateSpire()}
}
