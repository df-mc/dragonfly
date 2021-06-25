package provider

import "github.com/df-mc/dragonfly/server/world"

const (
	survival = uint8(iota)
	creative
	adventure
	spectator
)

func gamemodeToData(mode world.GameMode) uint8 {
	gm := survival
	if _, ok := mode.(world.GameModeCreative); ok {
		gm = creative
	} else if _, ok := mode.(world.GameModeAdventure); ok {
		gm = adventure
	} else if _, ok := mode.(world.GameModeSpectator); ok {
		gm = spectator
	}
	return gm
}

func dataToGamemode(mode uint8) world.GameMode {
	var gm world.GameMode
	switch mode {
	case creative:
		gm = world.GameModeCreative{}
	case adventure:
		gm = world.GameModeAdventure{}
	case spectator:
		gm = world.GameModeSpectator{}
	default:
		gm = world.GameModeSurvival{}
	}
	return gm
}
