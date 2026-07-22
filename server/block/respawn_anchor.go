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
	// ExplosionSize is the size of the explosion created when the respawn anchor is used outside the Nether. It
	// defaults to 5.
	ExplosionSize float64
}

// respawnAnchorSpawnOffsets holds the vanilla respawn search priority around an
// anchor: per column (cardinals before diagonals) at the anchor's level and the
// level above it, then the columns one level below, then on top of the anchor.
var respawnAnchorSpawnOffsets = []cube.Pos{
	{0, 0, -1}, {0, 1, -1},
	{-1, 0, 0}, {-1, 1, 0},
	{0, 0, 1}, {0, 1, 1},
	{1, 0, 0}, {1, 1, 0},
	{-1, 0, -1}, {-1, 1, -1},
	{1, 0, -1}, {1, 1, -1},
	{-1, 0, 1}, {-1, 1, 1},
	{1, 0, 1}, {1, 1, 1},
	{0, -1, -1},
	{-1, -1, 0},
	{0, -1, 1},
	{1, -1, 0},
	{-1, -1, -1},
	{1, -1, -1},
	{-1, -1, 1},
	{1, -1, 1},
	{0, 1, 0},
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
		r.explode(pos, tx)
		return true
	}

	if spawn, ok := tx.World().PlayerSpawnPoint(user.UUID()); ok && spawn.Pos == pos && spawn.Dim == world.Nether {
		return true
	}
	tx.World().SetPlayerSpawn(user.UUID(), pos)
	tx.PlaySound(pos.Vec3Centre(), sound.RespawnAnchorSetSpawn{})
	user.Messaget(chat.MessageRespawnAnchorRespawnPointSet)
	return true
}

// explode creates the Respawn anchor's incendiary explosion.
func (r RespawnAnchor) explode(pos cube.Pos, tx *world.Tx) {
	size := r.ExplosionSize
	if size == 0 {
		size = 5
	}
	ExplosionConfig{SpawnFire: true}.Explode(tx, world.BlockExplosionSource{
		Block:         r,
		Pos:           pos,
		ExplosionSize: size,
	})
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
	for _, offset := range respawnAnchorSpawnOffsets {
		spawn := pos.Add(offset)
		if respawnAnchorSpawnClear(spawn, tx) {
			return spawn, true
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
	if below.OutOfBounds(tx.Range()) || !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		return false
	}

	occupied := cube.Box(0, 0, 0, 1, 2, 1).Translate(pos.Vec3())
	for y := 0; y < 2; y++ {
		blockPos := pos.Add(cube.Pos{0, y})
		if _, liquid := tx.Liquid(blockPos); liquid {
			return false
		}
		for _, box := range tx.Block(blockPos).Model().BBox(blockPos, tx) {
			if box.Translate(blockPos.Vec3()).IntersectsWith(occupied) {
				return false
			}
		}
	}
	return true
}
