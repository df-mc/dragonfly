package item

import "github.com/df-mc/dragonfly/server/world"

type ArmourTrim struct {
	Template ArmourTrimTemplate
	Material TrimMaterial
}

type TrimMaterial interface {
	// TrimMaterial returns the material name used for reading and writing trim data.
	TrimMaterial() string
	// MaterialColor returns the color code used for internal text formatting. Use text.Colourf for proper formatting.
	MaterialColor() string
}

// MaterialFromString returns a TrimMaterial from a string.
func MaterialFromString(name string) TrimMaterial {
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

// TrimMaterials returns all the items that can be trim materials.
func TrimMaterials() []world.Item {
	return []world.Item{AmethystShard{}, CopperIngot{}, Diamond{}, Emerald{}, GoldIngot{}, IronIngot{}, LapisLazuli{}, NetheriteIngot{}, NetherQuartz{}}
}
