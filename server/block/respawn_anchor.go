package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// RespawnAnchor is a block that allows the player to set their spawn point in the Nether.
type RespawnAnchor struct {
	solid
	bassDrum

	// Charge is the Glowstone charge of the RespawnAnchor.
	Charge int
}

// LightEmissionLevel ...
func (r RespawnAnchor) LightEmissionLevel() uint8 {
	return uint8(max(0, 3+4*(r.Charge-1)))
}

// EncodeItem ...
func (r RespawnAnchor) EncodeItem() (name string, meta int16) {
	return "minecraft:respawn_anchor", 0
}

// EncodeBlock ...
func (r RespawnAnchor) EncodeBlock() (string, map[string]any) {
	return "minecraft:respawn_anchor", map[string]any{"respawn_anchor_charge": int32(r.Charge)}
}

// BreakInfo ...
func (r RespawnAnchor) BreakInfo() BreakInfo {
	return newBreakInfo(50, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(r)).withBlastResistance(6000)
}

// Activate ...
func (r RespawnAnchor) Activate(pos cube.Pos, clickedFace cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	_, usingGlowstone := held.Item().(Glowstone)

	sleeper, ok := u.(world.Sleeper)
	if !ok {
		return false
	}

	if r.Charge < 4 && usingGlowstone {
		r.Charge++
		tx.SetBlock(pos, r, nil)
		ctx.SubtractFromCount(1)
		tx.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorCharge{Charge: r.Charge})
		return true
	}

	w := tx.World()

	if r.Charge > 0 {
		if w.Dimension() == world.Nether {
			previousSpawn := w.PlayerSpawn(sleeper.UUID())
			if previousSpawn == pos {
				return false
			}
			sleeper.Messaget(chat.MessageRespawnPointSet)
			w.SetPlayerSpawn(sleeper.UUID(), pos)
			return true
		}
		tx.SetBlock(pos, nil, nil)
		ExplosionConfig{
			Size:      5,
			SpawnFire: true,
		}.Explode(tx, pos.Vec3Centre())
	}

	return false
}

// allRespawnAnchors returns all possible respawn anchors.
func allRespawnAnchors() []world.Block {
	all := make([]world.Block, 0, 5)
	for i := 0; i < 5; i++ {
		all = append(all, RespawnAnchor{Charge: i})
	}
	return all
}

func (r RespawnAnchor) CanRespawnOn() bool {
	return r.Charge > 0
}

func (r RespawnAnchor) RespawnOn(pos cube.Pos, u item.User, w *world.Tx) {
	w.SetBlock(pos, RespawnAnchor{Charge: r.Charge - 1}, nil)
	w.PlaySound(pos.Vec3(), sound.RespawnAnchorDeplete{Charge: r.Charge - 1})
}
