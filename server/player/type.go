package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"golang.org/x/text/language"
	"math/rand"
	"strings"
	"time"
)

// Type is a world.EntityType implementation for Player.
type Type struct{}

type Config struct {
	Name, XUID, Locale string
	UUID               uuid.UUID
	Skin               skin.Skin
	Data               *Data
	Pos                mgl64.Vec3
	Session            *session.Session
}

func (conf Config) Apply(data *world.EntityData) {
	locale, err := language.Parse(strings.Replace(conf.Locale, "_", "-", 1))
	if err != nil {
		locale = language.BritishEnglish
	}
	data.Name, data.Pos = conf.Name, conf.Pos
	data.Data = &playerData{
		ui:                inventory.New(54, nil),
		inv:               inventory.New(36, nil),
		enderChest:        inventory.New(27, nil),
		offHand:           inventory.New(1, nil),
		armour:            inventory.NewArmour(nil),
		hunger:            newHungerManager(),
		health:            entity.NewHealthManager(20, 20),
		experience:        entity.NewExperienceManager(),
		effects:           entity.NewEffectManager(),
		locale:            locale,
		breathing:         true,
		cooldowns:         make(map[string]time.Time),
		mc:                &entity.MovementComputer{Gravity: 0.08, Drag: 0.02, DragBeforeGravity: true},
		heldSlot:          new(uint32),
		gameMode:          world.GameModeSurvival,
		skin:              conf.Skin,
		airSupplyTicks:    300,
		maxAirSupplyTicks: 300,
		enchantSeed:       rand.Int63(),
		scale:             1.0,
		s:                 conf.Session,
	}
}

func (t Type) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	pd := data.Data.(*playerData)

	// With session:
	//	p := createSession(name, skin, pos)
	//	p.s.Store(s)
	//	p.skin.Store(&skin)
	//	p.uuid, p.xuid = uuid, xuid
	//	p.inv, p.offHand, p.enderChest, p.armour, p.heldSlot = s.HandleInventories()
	//	p.locale, _ = language.Parse(strings.Replace(s.ClientData().LanguageCode, "_", "-", 1))
	//	if data != nil {
	//		p.load(*data)
	//	}
	//	return p
	p := &Player{
		tx:         tx,
		handle:     handle,
		data:       data,
		playerData: pd,
	}

	if pd.s != nil {
		pd.s.HandleInventories(tx, p, pd.inv, pd.offHand, pd.enderChest, pd.ui, pd.armour, pd.heldSlot)
	} else {
		pd.inv.SlotFunc(func(slot int, before, after item.Stack) {
			if slot == int(*p.heldSlot) {
				p.broadcastItems(slot, before, after)
			}
		})
		pd.offHand.SlotFunc(p.broadcastItems)
		pd.armour.Inventory().SlotFunc(p.broadcastArmour)
	}

	return p
}

func (Type) EncodeEntity() string   { return "minecraft:player" }
func (Type) NetworkOffset() float64 { return 1.621 }
func (Type) BBox(e world.Entity) cube.BBox {
	p := e.(*Player)
	s := p.Scale()
	switch {
	case p.Sneaking():
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 1.5*s, 0.3*s)
	case p.Gliding(), p.Swimming():
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 0.6*s, 0.3*s)
	default:
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 1.8*s, 0.3*s)
	}
}
func (t Type) DecodeNBT(map[string]any, *world.EntityData) {}
func (t Type) EncodeNBT(*world.EntityData) map[string]any  { return nil }
