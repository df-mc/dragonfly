package playerdb

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
)

func (p *Provider) fromJson(d jsonData, lookupWorld func(world.Dimension) *world.World) player.Data {
	dim, _ := world.DimensionByID(int(d.Dimension))
	mode, _ := world.GameModeByID(int(d.GameMode))
	data := player.Data{
		UUID:                uuid.MustParse(d.UUID),
		Username:            d.Username,
		Position:            d.Position,
		Velocity:            d.Velocity,
		Yaw:                 d.Yaw,
		Pitch:               d.Pitch,
		Health:              d.Health,
		MaxHealth:           d.MaxHealth,
		Hunger:              d.Hunger,
		FoodTick:            d.FoodTick,
		ExhaustionLevel:     d.ExhaustionLevel,
		SaturationLevel:     d.SaturationLevel,
		AbsorptionLevel:     d.AbsorptionLevel,
		Experience:          d.Experience,
		AirSupply:           d.AirSupply,
		MaxAirSupply:        d.MaxAirSupply,
		EnchantmentSeed:     d.EnchantmentSeed,
		GameMode:            mode,
		Effects:             dataToEffects(d.Effects),
		FireTicks:           d.FireTicks,
		FallDistance:        d.FallDistance,
		Inventory:           dataToInv(d.Inventory),
		EnderChestInventory: make([]item.Stack, 27),
		World:               lookupWorld(dim),
	}
	decodeItems(d.EnderChestInventory, data.EnderChestInventory)
	return data
}

func (p *Provider) toJson(d player.Data) jsonData {
	dim, _ := world.DimensionID(d.World.Dimension())
	mode, _ := world.GameModeID(d.GameMode)
	return jsonData{
		UUID:                d.UUID.String(),
		Username:            d.Username,
		Position:            d.Position,
		Velocity:            d.Velocity,
		Yaw:                 d.Yaw,
		Pitch:               d.Pitch,
		Health:              d.Health,
		MaxHealth:           d.MaxHealth,
		Hunger:              d.Hunger,
		FoodTick:            d.FoodTick,
		ExhaustionLevel:     d.ExhaustionLevel,
		SaturationLevel:     d.SaturationLevel,
		AbsorptionLevel:     d.AbsorptionLevel,
		Experience:          d.Experience,
		AirSupply:           d.AirSupply,
		MaxAirSupply:        d.MaxAirSupply,
		EnchantmentSeed:     d.EnchantmentSeed,
		GameMode:            uint8(mode),
		Effects:             effectsToData(d.Effects),
		FireTicks:           d.FireTicks,
		FallDistance:        d.FallDistance,
		Inventory:           invToData(d.Inventory),
		EnderChestInventory: encodeItems(d.EnderChestInventory),
		Dimension:           uint8(dim),
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
	AbsorptionLevel                  float64
	EnchantmentSeed                  int64
	Experience                       int
	AirSupply, MaxAirSupply          int64
	GameMode                         uint8
	Inventory                        jsonInventoryData
	EnderChestInventory              []jsonSlot
	Effects                          []jsonEffect
	FireTicks                        int64
	FallDistance                     float64
	Dimension                        uint8
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
