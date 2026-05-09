package player

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestPortalTravelInstantaneousUsesLiveGameMode(t *testing.T) {
	var data world.EntityData
	Config{GameMode: world.GameModeSurvival}.Apply(&data)

	pdata := data.Data.(*playerData)
	if pdata.portalTravel.Instantaneous() {
		t.Fatal("survival player had instantaneous portal travel")
	}

	pdata.gameMode = world.GameModeCreative
	if !pdata.portalTravel.Instantaneous() {
		t.Fatal("creative player did not have instantaneous portal travel after game mode change")
	}

	pdata.gameMode = world.GameModeSurvival
	if pdata.portalTravel.Instantaneous() {
		t.Fatal("survival player still had instantaneous portal travel after game mode change")
	}
}
