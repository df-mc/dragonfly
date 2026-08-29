package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/google/uuid"
)

// RespawnAnchor is a block that allows players to set their respawn point in the Nether after charging it with
// glowstone.
type RespawnAnchor struct {
	solid
	bassDrum

	// Charges is the amount of glowstone charges stored in the respawn anchor, in the range 0-4.
	Charges int
}

// Activate ...
func (r RespawnAnchor) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Item().(Glowstone); ok && r.Charges < 4 {
		r.Charges++
		tx.SetBlock(pos, r, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorCharge{})
		ctx.SubtractFromCount(1)
		return true
	}
	if r.Charges == 0 {
		return false
	}

	user, ok := u.(interface {
		UUID() uuid.UUID
		Messaget(t chat.Translation, a ...any)
	})
	if !ok {
		return false
	}

	if tx.World().Dimension() != world.Nether {
		tx.SetBlock(pos, nil, nil)
		if tx.World().RespawnBlocksExplode() {
			r.explode(pos, tx)
		}
		return true
	}

	if spawn, ok := tx.World().PlayerSpawnPoint(user.UUID()); ok && spawn.Pos == pos && spawn.Dim == world.Nether {
		return false
	}
	tx.World().SetPlayerSpawn(user.UUID(), pos)
	tx.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorSetSpawn{})
	user.Messaget(chat.MessageRespawnAnchorRespawnPointSet)
	return true
}

// explode creates the Respawn anchor's incendiary explosion.
func (r RespawnAnchor) explode(pos cube.Pos, tx *world.Tx) {
	ExplosionConfig{SpawnFire: true, SuppressUnderwaterImpact: true}.explode(tx, world.BlockExplosionSource{
		Block:         r,
		Pos:           pos,
		ExplosionSize: 5,
	}, respawnAnchorTouchesWater(pos, tx))
}

// respawnAnchorTouchesWater reports whether water touches any face of the anchor. Vanilla treats the resulting
// explosion as underwater in this case.
func respawnAnchorTouchesWater(pos cube.Pos, tx *world.Tx) bool {
	for _, face := range cube.Faces() {
		if liquid, ok := tx.Liquid(pos.Side(face)); ok {
			if _, water := liquid.(Water); water {
				return true
			}
		}
	}
	return false
}

// CanRespawnOn ...
func (r RespawnAnchor) CanRespawnOn() bool {
	return r.Charges > 0
}

// SafeSpawn ...
func (r RespawnAnchor) SafeSpawn(pos cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	if !r.CanRespawnOn() || tx.World().Dimension() != world.Nether {
		return cube.Pos{}, false
	}
	if respawnAnchorSpawnClear(pos, tx) {
		return pos, true
	}
	// Search the surrounding 4x4 area in vanilla order, with X as the outer axis and Z as the inner axis.
	for x := -1; x <= 2; x++ {
		for z := -1; z <= 2; z++ {
			spawn := pos.Add(cube.Pos{x, 0, z})
			if respawnAnchorSpawnClear(spawn, tx) {
				return spawn, true
			}
		}
	}
	return cube.Pos{}, false
}

// UseRespawn consumes one charge after a player respawns at the anchor.
func (r RespawnAnchor) UseRespawn(pos cube.Pos, tx *world.Tx) {
	if r.Charges == 0 {
		return
	}
	r.Charges--
	tx.SetBlock(pos, r, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorDeplete{})
}

// BreakInfo ...
func (r RespawnAnchor) BreakInfo() BreakInfo {
	return newBreakInfo(50, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(RespawnAnchor{})).withBlastResistance(1200)
}

// LightEmissionLevel ...
func (r RespawnAnchor) LightEmissionLevel() uint8 {
	if r.Charges < 1 || r.Charges > 4 {
		return 0
	}
	return [...]uint8{0, 3, 7, 11, 15}[r.Charges]
}

// EncodeItem ...
func (r RespawnAnchor) EncodeItem() (name string, meta int16) {
	return "minecraft:respawn_anchor", 0
}

// EncodeBlock ...
func (r RespawnAnchor) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:respawn_anchor", map[string]any{"respawn_anchor_charge": int32(r.Charges)}
}

// allRespawnAnchors returns all possible respawn anchor block states.
func allRespawnAnchors() (anchors []world.Block) {
	for charges := 0; charges <= 4; charges++ {
		anchors = append(anchors, RespawnAnchor{Charges: charges})
	}
	return
}

func respawnAnchorSpawnClear(pos cube.Pos, tx *world.Tx) bool {
	if pos.OutOfBounds(tx.Range()) || pos.Side(cube.FaceUp).OutOfBounds(tx.Range()) {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if below.OutOfBounds(tx.Range()) || !respawnAnchorSolid(below, tx) {
		return false
	}
	for y := 0; y < 2; y++ {
		if respawnAnchorSolid(pos.Add(cube.Pos{0, y}), tx) {
			return false
		}
	}
	return true
}

func respawnAnchorSolid(pos cube.Pos, tx *world.Tx) bool {
	model := tx.Block(pos).Model()
	for _, face := range cube.Faces() {
		if !model.FaceSolid(pos, face, tx) {
			return false
		}
	}
	return true
}
