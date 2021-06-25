package provider

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

func fromJson(d jsonData) player.Data {
	return player.Data{
		XUID:            d.XUID,
		Username:        d.Username,
		Position:        d.Position,
		Velocity:        d.Velocity,
		Yaw:             d.Yaw,
		Pitch:           d.Pitch,
		Health:          d.Health,
		MaxHealth:       d.MaxHealth,
		Hunger:          d.Hunger,
		FoodTick:        d.FoodTick,
		ExhaustionLevel: d.ExhaustionLevel,
		SaturationLevel: d.SaturationLevel,
		Gamemode:        dataToGamemode(d.Gamemode),
		Effects:         dataToEffects(d.Effects),
		FireTicks:       d.FireTicks,
		FallDistance:    d.FallDistance,
		Inventory:       dataToInv(d.Inventory),
	}
}

func toJson(d player.Data) jsonData {
	return jsonData{
		XUID:            d.XUID,
		Username:        d.Username,
		Position:        d.Position,
		Velocity:        d.Velocity,
		Yaw:             d.Yaw,
		Pitch:           d.Pitch,
		Health:          d.Health,
		MaxHealth:       d.MaxHealth,
		Hunger:          d.Hunger,
		FoodTick:        d.FoodTick,
		ExhaustionLevel: d.ExhaustionLevel,
		SaturationLevel: d.SaturationLevel,
		Gamemode:        gamemodeToData(d.Gamemode),
		Effects:         effectsToData(d.Effects),
		FireTicks:       d.FireTicks,
		FallDistance:    d.FallDistance,
		Inventory:       invToData(d.Inventory),
	}
}

type jsonData struct {
	XUID                             string
	Username                         string
	Position, Velocity               mgl64.Vec3
	Yaw, Pitch                       float64
	Health, MaxHealth                float64
	Hunger                           int
	FoodTick                         int
	ExhaustionLevel, SaturationLevel float64
	Gamemode                         uint8
	Inventory                        jsonInventoryData
	Effects                          []jsonEffect
	FireTicks                        int64
	FallDistance                     float64
}

type jsonInventoryData struct {
	Items    []map[string]interface{}
	Armor    [4]map[string]interface{}
	Offhand  map[string]interface{}
	Mainhand uint32
}

type jsonEffect struct {
	ID       int
	Level    int
	Duration time.Duration
	Ambient  bool
}
