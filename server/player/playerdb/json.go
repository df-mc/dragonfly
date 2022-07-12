package playerdb

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
)

func fromJson(d jsonData) player.Data {
	return player.Data{
		UUID:            uuid.MustParse(d.UUID),
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
		Experience:      d.Experience,
		AirSupply:       d.AirSupply,
		MaxAirSupply:    d.MaxAirSupply,
		GameMode:        dataToGameMode(d.GameMode),
		Effects:         dataToEffects(d.Effects),
		FireTicks:       d.FireTicks,
		FallDistance:    d.FallDistance,
		Inventory:       dataToInv(d.Inventory),
		Dimension:       d.Dimension,
	}
}

func toJson(d player.Data) jsonData {
	return jsonData{
		UUID:            d.UUID.String(),
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
		Experience:      d.Experience,
		AirSupply:       d.AirSupply,
		MaxAirSupply:    d.MaxAirSupply,
		GameMode:        gameModeToData(d.GameMode),
		Effects:         effectsToData(d.Effects),
		FireTicks:       d.FireTicks,
		FallDistance:    d.FallDistance,
		Inventory:       invToData(d.Inventory),
		Dimension:       d.Dimension,
	}
}

type jsonData struct {
	UUID                             string
	Username                         string
	Position, Velocity               mgl64.Vec3
	Yaw, Pitch                       float64
	Health, MaxHealth                float64
	Hunger                           int
	FoodTick                         int
	ExhaustionLevel, SaturationLevel float64
	Experience                       int
	AirSupply, MaxAirSupply          int64
	GameMode                         uint8
	Inventory                        jsonInventoryData
	Effects                          []jsonEffect
	FireTicks                        int64
	FallDistance                     float64
	Dimension                        int
}

type jsonInventoryData struct {
	Items        []jsonSlot
	Boots        []byte
	Leggings     []byte
	Chestplate   []byte
	Helmet       []byte
	OffHand      []byte
	MainHandSlot uint32
}

type jsonSlot struct {
	Item []byte
	Slot int
}

type jsonEffect struct {
	ID       int
	Level    int
	Duration time.Duration
	Ambient  bool
}
