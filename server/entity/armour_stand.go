package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

// NewArmourStand creates a new armour stand entity in the world with the given spawn options.
func NewArmourStand(opts world.EntitySpawnOpts) *world.EntityHandle {
	conf := armourStandConf
	conf.Armour = inventory.NewArmour(nil)
	return opts.New(ArmourStandType, conf)
}

var armourStandConf = ArmourStandBehaviourConfig{}

var ArmourStandType armourStandType

// armourStandType is a world.EntityType implementation for armour stands.
type armourStandType struct{}

func (armourStandType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (armourStandType) EncodeEntity() string { return "minecraft:armor_stand" }

func (armourStandType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.25, 0, -0.25, 0.25, 1.975, 0.25)
}

func (armourStandType) DecodeNBT(m map[string]any, data *world.EntityData) {
	c := ArmourStandBehaviourConfig{
		Armour:    inventory.NewArmour(nil),
		PoseIndex: int(nbtconv.Int32(m, "PoseIndex")) % 13,
		MainHand:  nbtconv.MapItem(m, "MainHand"),
		OffHand:   nbtconv.MapItem(m, "Offhand"),
	}
	armours := nbtconv.Slice(m, "Armor")
	for i := 0; i < 4; i++ {
		itemMap, ok := armours[i].(map[string]any)
		if !ok {
			continue
		}
		var it item.Stack
		switch i {
		case 0:
			nbtconv.Item(itemMap, &it)
			c.Armour.SetHelmet(it)
		case 1:
			nbtconv.Item(itemMap, &it)
			c.Armour.SetChestplate(it)
		case 2:
			nbtconv.Item(itemMap, &it)
			c.Armour.SetLeggings(it)
		case 3:
			nbtconv.Item(itemMap, &it)
			c.Armour.SetBoots(it)
		}
	}
	data.Data = c.New()
}

func (armourStandType) EncodeNBT(data *world.EntityData) map[string]any {
	a := data.Data.(*ArmourStandBehaviour)
	return map[string]any{
		"MainHand": nbtconv.WriteItem(a.conf.MainHand, true),
		"Offhand":  nbtconv.WriteItem(a.conf.OffHand, true),
		"Armor": []map[string]any{
			nbtconv.WriteItem(a.Armour().Helmet(), true),
			nbtconv.WriteItem(a.Armour().Chestplate(), true),
			nbtconv.WriteItem(a.Armour().Leggings(), true),
			nbtconv.WriteItem(a.Armour().Boots(), true),
		},
		"PoseIndex": int32(a.conf.PoseIndex),
	}
}
