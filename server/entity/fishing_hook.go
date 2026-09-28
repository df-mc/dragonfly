package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// NewFishingHook creates a fishing hook entity cast by the owner passed. The
// lure level decreases the time it takes for a fish to bite, while the luck
// level increases the chance of catching treasure.
func NewFishingHook(opts world.EntitySpawnOpts, owner world.Entity, lure, luck int) *world.EntityHandle {
	conf := fishingHookConf
	conf.Owner = owner.H()
	conf.Lure, conf.Luck = lure, luck
	return opts.New(FishingHookType, conf)
}

var fishingHookConf = FishingHookBehaviourConfig{}

// FishingHookType is a world.EntityType implementation for fishing hooks.
var FishingHookType fishingHookType

type fishingHookType struct{}

func (t fishingHookType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (fishingHookType) EncodeEntity() string { return "minecraft:fishing_hook" }
func (fishingHookType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// DecodeNBT creates a fishing hook without an owner. Fishing hooks are not
// persisted in vanilla, so a hook loaded from disk removes itself on its first
// tick.
func (fishingHookType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = fishingHookConf.New()
}
func (fishingHookType) EncodeNBT(*world.EntityData) map[string]any { return nil }
