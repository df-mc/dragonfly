package item

import "github.com/df-mc/dragonfly/server/world"

// ArmourTrim is a decorative addition to an armour piece. It consists of a
// template that specifies the pattern and a material that specifies the colour.
type ArmourTrim struct {
	Template SmithingTemplateType
	Material ArmourTrimMaterial
}

// Zero checks if an ArmourTrim is considered zero: Either its material is nil
// or its template TemplateNetheriteUpgrade.
func (trim ArmourTrim) Zero() bool {
	return trim.Material == nil || trim.Template == TemplateNetheriteUpgrade()
}

// ArmourTrimMaterial is the material of an ArmourTrim, such as an IronIngot,
// that modifies the colour of an ArmourTrim.
type ArmourTrimMaterial interface {
	// TrimMaterial returns the material name used for reading and writing trim data.
	TrimMaterial() string
	// MaterialColour returns the colour code used for internal text formatting.
	MaterialColour() string
}

// trimMaterialFromString returns a TrimMaterial from a string.
func trimMaterialFromString(name string) (ArmourTrimMaterial, bool) {
	switch name {
	case "amethyst":
		return AmethystShard{}, true
	case "copper":
		return CopperIngot{}, true
	case "diamond":
		return Diamond{}, true
	case "emerald":
		return Emerald{}, true
	case "gold":
		return GoldIngot{}, true
	case "iron":
		return IronIngot{}, true
	case "lapis":
		return LapisLazuli{}, true
	case "netherite":
		return NetheriteIngot{}, true
	case "quartz":
		return NetherQuartz{}, true
	case "resin":
		return ResinBrick{}, true
	}
	// TODO: add redstone material once pr is merged
	return nil, false
}

// ArmourTrimMaterials returns all the items that can be trim materials.
func ArmourTrimMaterials() []world.Item {
	return []world.Item{
		AmethystShard{},
		CopperIngot{},
		Diamond{},
		Emerald{},
		GoldIngot{},
		IronIngot{},
		LapisLazuli{},
		NetheriteIngot{},
		NetherQuartz{},
		ResinBrick{},
	}
}

// Trimmable represents an item, generally Armour, that can have an ArmourTrim
// applied to it in a smithing table.
type Trimmable interface {
	WithTrim(trim ArmourTrim) world.Item
}
