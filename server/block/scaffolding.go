package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Scaffolding is a block used for building and climbing, that extends horizontally up to a limited distance from
// a supporting block before it collapses.
//
// Unlike vanilla, Scaffolding deliberately does not implement LiquidDisplacer and can never become waterlogged.
// Real Bedrock scaffolding can be waterlogged, but a waterlogged climbable block cannot be climbed on Bedrock
// (the same is true of ladders; see e.g. Mojang bug reports MC-129479 and MC-127102). This means a scaffolding
// column that passes through water is unclimbable at the exact point where it transitions from the waterlogged
// section to the dry section: the player's hitbox straddles a wet, non-climbable block and a dry, climbable one
// at the same time, and the client never resolves which movement mode to apply, permanently freezing the player
// there until they leave. Placing scaffolding into water instead simply displaces the water, guaranteeing the
// whole column stays climbable at the cost of not preserving vanilla's waterlogging.
type Scaffolding struct {
	transparent

	// Stability is the distance of the scaffolding from a block supporting it, ranging from 0 to 7. A value of 0
	// means it rests on top of a supporting block or another scaffolding column; higher values are reached by
	// extending horizontally. When Stability reaches 7 the scaffolding can no longer support itself.
	Stability int
	// StabilityCheck is true when the scaffolding is the bottom of its column, meaning it does not rest on top of
	// another scaffolding block.
	StabilityCheck bool
}

// NeighbourUpdateTick ...
func (s Scaffolding) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if scaffoldingNextToLavaOrFire(pos, tx) {
		// Bedrock destroys scaffolding placed next to lava or fire outright, without it catching fire and
		// burning down over time the way its Flammability would otherwise cause.
		breakBlockNoDrops(s, pos, tx)
		return
	}
	tx.ScheduleBlockUpdate(pos, s, time.Millisecond*50)
}

// scaffoldingNextToLavaOrFire reports whether any of the six blocks adjacent to pos is Lava or Fire.
func scaffoldingNextToLavaOrFire(pos cube.Pos, tx *world.Tx) bool {
	found := false
	pos.Neighbours(func(neighbour cube.Pos) {
		if found {
			return
		}
		switch tx.Block(neighbour).(type) {
		case Lava, Fire:
			found = true
		}
	}, tx.Range())
	return found
}

// ScheduledTick recalculates the stability of the scaffolding. If it has lost its support, the scaffolding either
// breaks and drops as an item or falls as an entity.
func (s Scaffolding) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	dist := scaffoldingDistance(pos, tx)
	if dist == 7 {
		if s.Stability == 7 {
			// The scaffolding was already unsupported on the previous tick, so it now falls as an entity. This
			// happens when a scaffolding is placed more than the maximum distance from its support.
			tx.SetBlock(pos, nil, nil)
			opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
			tx.AddEntity(tx.World().EntityRegistry().Config().FallingBlock(opts, s))
			return
		}
		// The scaffolding just lost its support, so it breaks and drops as an item. Breaking it triggers the same
		// check on the scaffolding above, cascading up the column.
		breakBlock(s, pos, tx)
		return
	}
	bottom := dist > 0
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Scaffolding); ok {
		bottom = false
	}
	if s.Stability != dist || s.StabilityCheck != bottom {
		s.Stability, s.StabilityCheck = dist, bottom
		tx.SetBlock(pos, s, nil)
	}
}

// UseOnBlock ...
func (s Scaffolding) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, ok := scaffoldingPlacementPos(pos, face, tx, s)
	if !ok {
		return false
	}
	if _, ok := tx.Block(pos).(Lava); ok {
		// Bedrock does not allow scaffolding to be placed inside lava, unlike most other blocks that can
		// simply displace it.
		return false
	}

	s.Stability = scaffoldingDistance(pos, tx)
	// Scaffolding may be placed inside the placing entity, which is required to build a tower upwards while
	// standing on it.
	ctx.IgnoreBBox = true

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// scaffoldingPlacementPos resolves the position a new scaffolding block should be placed at, given the position
// and face clicked. A click against any face of an existing scaffolding block attaches the new block directly
// to that face like a normal block, including sideways, so that scaffolding can be extended horizontally.
//
// Two exceptions redirect the placement to the top of a scaffolding column instead, and both are safe because
// they only ever climb straight up the single column reachable from the clicked block - they can never jump
// sideways into an unrelated structure:
//   - Clicking the underside of a scaffolding block: scaffolding cannot attach below another scaffolding block,
//     so Bedrock redirects this into a "tower up" shortcut.
//   - Clicking the top face of a scaffolding block that already has more scaffolding above it (e.g. clicking a
//     lower block of a tall column instead of the exact tip): the cell directly above is occupied, but it is
//     part of the very same column, so the placement is redirected to the top of it rather than failing. This
//     is what makes towering reliable even when the click does not land on the exact topmost block.
//
// Any other face whose target cell is already occupied simply fails to place, exactly like placing any other
// block against an already-occupied cell. In particular, a sideways click that happens to point at an unrelated,
// separate column is never redirected, since that column has no relation to the block that was clicked.
func scaffoldingPlacementPos(pos cube.Pos, face cube.Face, tx *world.Tx, s Scaffolding) (cube.Pos, bool) {
	if _, ok := tx.Block(pos).(Scaffolding); ok && face == cube.FaceDown {
		pos = scaffoldingColumnTop(pos, tx).Side(cube.FaceUp)
		return pos, replaceableWith(tx, pos, s)
	}
	if resolved, _, used := firstReplaceable(tx, pos, face, s); used {
		return resolved, true
	}
	if target := pos.Side(face); face == cube.FaceUp {
		if _, ok := tx.Block(target).(Scaffolding); ok {
			pos = scaffoldingColumnTop(target, tx).Side(cube.FaceUp)
			return pos, replaceableWith(tx, pos, s)
		}
	}
	return cube.Pos{}, false
}

// scaffoldingColumnTop returns the position of the topmost scaffolding block in the vertical column that pos,
// itself a Scaffolding block, belongs to.
func scaffoldingColumnTop(pos cube.Pos, tx *world.Tx) cube.Pos {
	for {
		above := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(above).(Scaffolding); !ok {
			return pos
		}
		pos = above
	}
}

// EntityInside ...
func (Scaffolding) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
}

// BreakInfo ...
func (s Scaffolding) BreakInfo() BreakInfo {
	// minecraft.wiki lists "Any tool" for Scaffolding, i.e. no tool is effective against it specifically. This
	// has no practical effect on break speed since a hardness of 0 already breaks the block instantly regardless
	// of Effective (see BreaksInstantly), but nothingEffective reflects the sourced behaviour accurately.
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

// FlammabilityInfo ...
func (Scaffolding) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 60, false)
}

// Model ...
func (Scaffolding) Model() world.BlockModel {
	return model.Scaffolding{}
}

// EncodeItem ...
func (Scaffolding) EncodeItem() (name string, meta int16) {
	return "minecraft:scaffolding", 0
}

// EncodeBlock ...
func (s Scaffolding) EncodeBlock() (string, map[string]any) {
	return "minecraft:scaffolding", map[string]any{
		"stability":       int32(s.Stability),
		"stability_check": s.StabilityCheck,
	}
}

// scaffoldingDistance returns the stability a scaffolding block would have at the position passed, based on the
// block below it and any horizontally adjacent scaffolding. It returns 0 if the block rests on a supporting block
// and 7 if it is unsupported.
func scaffoldingDistance(pos cube.Pos, tx *world.Tx) int {
	below := pos.Side(cube.FaceDown)
	belowBlock := tx.Block(below)
	dist := 7
	if s, ok := belowBlock.(Scaffolding); ok {
		dist = s.Stability
	} else if belowBlock.Model().FaceSolid(below, cube.FaceUp, tx) {
		return 0
	}
	for _, face := range []cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceWest, cube.FaceEast} {
		if s, ok := tx.Block(pos.Side(face)).(Scaffolding); ok {
			if d := s.Stability + 1; d < dist {
				dist = d
			}
			if dist == 1 {
				break
			}
		}
	}
	return dist
}

// allScaffolding ...
func allScaffolding() (b []world.Block) {
	for stability := 0; stability <= 7; stability++ {
		b = append(b, Scaffolding{Stability: stability})
		b = append(b, Scaffolding{Stability: stability, StabilityCheck: true})
	}
	return
}
