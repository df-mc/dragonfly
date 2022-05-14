package playerdb

import "github.com/df-mc/dragonfly/server/world"

const (
	survival = uint8(iota)
	creative
	adventure
	spectator
)

func gameModeToData(mode world.GameMode) uint8 {
	switch mode {
	case world.GameModeCreative:
		return creative
	case world.GameModeAdventure:
		return adventure
	case world.GameModeSpectator:
		return spectator
	default:
		return survival
	}
}

func dataToGameMode(mode uint8) world.GameMode {
	switch mode {
	case creative:
		return world.GameModeCreative
	case adventure:
		return world.GameModeAdventure
	case spectator:
		return world.GameModeSpectator
	default:
		return world.GameModeSurvival
	}
}
