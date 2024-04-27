package item

import "github.com/df-mc/dragonfly/server/world"

type ArmourTrim struct {
	Template ArmourSmithingTemplate
	Material ArmourTrimMaterial
}

type ArmourTrimMaterial interface {
	// TrimMaterial returns the material name used for reading and writing trim data.
	TrimMaterial() string
	// MaterialColour returns the colour code used for internal text formatting.
	MaterialColour() string
}

// ArmourTrimMaterialFromString returns a TrimMaterial from a string.
func ArmourTrimMaterialFromString(name string) ArmourTrimMaterial {
	switch name {
	case "amethyst":
		return AmethystShard{}
	case "copper":
		return CopperIngot{}
	case "diamond":
		return Diamond{}
	case "emerald":
		return Emerald{}
	case "gold":
		return GoldIngot{}
	case "iron":
		return IronIngot{}
	case "lapis":
		return LapisLazuli{}
	case "netherite":
		return NetheriteIngot{}
	case "quartz":
		return NetherQuartz{}
	}

	//TODO: add redstone material once pr is merged

	panic("should not happen")
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
	}
}
