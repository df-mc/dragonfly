package playerdb

import "github.com/df-mc/dragonfly/server/world"

const (
	survival = uint8(iota)
	creative
	adventure
	spectator
)

func gameModeToData(mode world.GameMode) uint8 {
	gm := survival
	switch mode.(type) {
	case world.GameModeCreative:
		gm = creative
	case world.GameModeAdventure:
		gm = adventure
	case world.GameModeSpectator:
		gm = spectator
	}
	return gm
}

func dataToGameMode(mode uint8) world.GameMode {
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
