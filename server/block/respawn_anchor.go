package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// RespawnAnchor is a block , allows players to set there respawn point in the nether.
type RespawnAnchor struct {
	solid
	bassDrum
	charge int32
}

// LightEmissionLevel ...
func (r RespawnAnchor) LightEmissionLevel() uint8 {
	return (uint8(r.charge) + 1) * 3
}

// EncodeItem ...
func (r RespawnAnchor) EncodeItem() (name string, meta int16) {
	return "minecraft:respawn_anchor", int16(r.charge)
}

// EncodeBlock ...
func (r RespawnAnchor) EncodeBlock() (string, map[string]any) {
	return "minecraft:respawn_anchor", map[string]any{"respawn_anchor_charge": r.charge}
}

// BreakInfo ...
func (r RespawnAnchor) BreakInfo() BreakInfo {
	return newBreakInfo(35, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(r)).withBlastResistance(6000)
}

// Activate ...
func (r RespawnAnchor) Activate(pos cube.Pos, clickedFace cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {

	var nether = w.Dimension().WaterEvaporates()

	held, _ := u.HeldItems()
	_, usingGlowStone := held.Item().(Glowstone)

	if r.charge < 4 && usingGlowStone {
		r.charge++
		w.SetBlock(pos, r, nil)
		ctx.SubtractFromCount(1)
		w.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorCharge{Charge: r.charge})
		return true
	}
	if nether {
		if r.charge > 0 {
			u.Messaget(text.Colourf("<grey>%%tile.bed.respawnSet</grey>"))
			u.SetSpawnPos(pos, w)
		}
		return false
	}

	if r.charge > 0 {
		w.SetBlock(pos, nil, nil)
		ExplosionConfig{
			Size:      5,
			SpawnFire: true,
		}.Explode(w, pos.Vec3Centre())
	}

	return false
}

// allRespawnAnchors returns all possible respawn anchors.
func allRespawnAnchors() []world.Block {
	all := make([]world.Block, 0, 5)
	for i := int32(0); i < 5; i++ {
		all = append(all, RespawnAnchor{charge: i})
	}
	return all
}

// allRespawnAnchorsItems returns all possible respawn anchors as items.
func allRespawnAnchorsItems() []world.Item {
	all := make([]world.Item, 0, 5)
	for i := int32(0); i < 5; i++ {
		all = append(all, RespawnAnchor{charge: i})
	}
	return all
}

func (r RespawnAnchor) SpawnBlock() bool {
	return r.charge > 0
}

func (r RespawnAnchor) Update(pos cube.Pos, u item.User, w *world.World) {
	w.SetBlock(pos, RespawnAnchor{charge: r.charge - 1}, nil)
	w.PlaySound(pos.Vec3(), sound.RespawnAnchorDeplete{Charge: r.charge - 1})
}
