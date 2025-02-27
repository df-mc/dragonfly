package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"golang.org/x/text/language"
	"math/rand/v2"
	"time"
)

// Config holds options that a Player can be created with.
type Config struct {
	Session  *session.Session
	Skin     skin.Skin
	XUID     string
	UUID     uuid.UUID
	Name     string
	Locale   language.Tag
	GameMode world.GameMode

	Position               mgl64.Vec3
	Rotation               cube.Rotation
	Velocity               mgl64.Vec3
	Health                 float64
	MaxHealth              float64
	FoodTick               int
	Food                   int
	Exhaustion, Saturation float64
	AirSupply              int
	MaxAirSupply           int
	EnchantmentSeed        int64
	Experience             int
	HeldSlot               int
	Inventory              *inventory.Inventory
	OffHand                *inventory.Inventory
	Armour                 *inventory.Armour
	EnderChestInventory    *inventory.Inventory
	FireTicks              int64
	FallDistance           float64
	Effects                []effect.Effect
}

// Apply applies fields from a Config to a world.EntityData, filling out empty
// fields with reasonable defaults.
func (cfg Config) Apply(data *world.EntityData) {
	conf := fillDefaults(cfg)

	data.Name, data.Pos, data.Rot = conf.Name, conf.Position, conf.Rotation
	slot := uint32(conf.HeldSlot)
	pdata := &playerData{
		xuid:                conf.XUID,
		ui:                  inventory.New(54, nil),
		inv:                 conf.Inventory,
		enderChest:          conf.EnderChestInventory,
		offHand:             conf.OffHand,
		armour:              conf.Armour,
		hunger:              newHungerManager(),
		health:              entity.NewHealthManager(conf.Health, conf.MaxHealth), // 20, 20
		experience:          entity.NewExperienceManager(),
		effects:             entity.NewEffectManager(conf.Effects...),
		locale:              conf.Locale,
		cooldowns:           make(map[string]time.Time),
		mc:                  &entity.MovementComputer{Gravity: 0.08, Drag: 0.02, DragBeforeGravity: true},
		heldSlot:            &slot,
		gameMode:            conf.GameMode,
		skin:                conf.Skin,
		enchantSeed:         conf.EnchantmentSeed,
		s:                   conf.Session,
		h:                   NopHandler{},
		speed:               0.1,
		flightSpeed:         0.05,
		verticalFlightSpeed: 1.0,
		scale:               1.0,
		airSupplyTicks:      conf.AirSupply,
		maxAirSupplyTicks:   conf.MaxAirSupply,
		breathing:           true,
		nameTag:             conf.Name,
		fireTicks:           conf.FireTicks,
		fallDistance:        conf.FallDistance,
	}
	pdata.hunger.foodLevel, pdata.hunger.foodTick, pdata.hunger.exhaustionLevel, pdata.hunger.saturationLevel = conf.Food, conf.FoodTick, conf.Exhaustion, conf.Saturation
	pdata.experience.Add(conf.Experience)
	data.Data = pdata
}

// fillDefaults fills empty fields in a Config with reasonable default values.
func fillDefaults(conf Config) Config {
	if (conf.Locale == language.Tag{}) {
		conf.Locale = language.BritishEnglish
	}
	if conf.Inventory == nil {
		conf.Inventory = inventory.New(36, nil)
	}
	if conf.EnderChestInventory == nil {
		conf.EnderChestInventory = inventory.New(27, nil)
	}
	if conf.OffHand == nil {
		conf.OffHand = inventory.New(1, nil)
	}
	if conf.Armour == nil {
		conf.Armour = inventory.NewArmour(nil)
	}
	if conf.Food == 0 && conf.FoodTick == 0 && conf.Exhaustion == 0 && conf.Saturation == 0 {
		conf.Food, conf.Saturation = 20, 5
	}
	if conf.EnchantmentSeed == 0 {
		conf.EnchantmentSeed = rand.Int64()
	}
	if conf.MaxAirSupply == 0 {
		conf.AirSupply, conf.MaxAirSupply = 300, 300
	}
	if conf.MaxHealth == 0 {
		conf.MaxHealth, conf.Health = 20, 20
	}
	if conf.GameMode == nil {
		conf.GameMode = world.GameModeSurvival
	}
	return conf
}
