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
// A scaffolding column built through a body of water was found to permanently trap the player at the point
// where it transitions from the waterlogged section to the dry section: a recorded reproduction (climbing up
// through a 3-block-deep water-filled shaft) shows the player's position and camera both freeze completely for
// several seconds right at the water surface, unable to move in either direction, until giving up. The likely
// mechanism is that the player's hitbox straddles a wet block and a dry, climbable one at the same time, and the
// client never resolves which movement mode (swimming or climbing) to apply, but this has not been confirmed
// against Bedrock's own source, only observed and reproduced. Displacing the water on placement instead of
// waterlogging guarantees the whole column stays climbable at the cost of not preserving vanilla's waterlogging.
type Scaffolding struct {
	transparent

	// Stability is the distance of the scaffolding from a block supporting it, ranging from 0 to 7. A value of 0
	// means it rests on top of a supporting block or another scaffolding column; higher values are reached by
	// extending horizontally. When Stability reaches 7 the scaffolding can no longer support itself.
	Stability int
	// StabilityCheck is true only when the scaffolding is both extended away from a support (Stability > 0) and
	// does not rest on top of another scaffolding block, meaning it is floating and held up solely by a
	// horizontal neighbour. It is false when Stability is 0 (resting directly on a supporting block) and also
	// false when resting on top of another scaffolding block, regardless of that column's own stability.
	StabilityCheck bool
}

// NeighbourUpdateTick ...
func (s Scaffolding) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	tx.ScheduleBlockUpdate(pos, s, time.Millisecond*50)
}

// ScheduledTick recalculates the stability of the scaffolding. If it has lost its support, the scaffolding either
// breaks and drops as an item or falls as an entity.
func (s Scaffolding) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	dist, bottom := scaffoldingStability(pos, tx)
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

	s.Stability, s.StabilityCheck = scaffoldingStability(pos, tx)
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
	// Unlike Ladder and Vines, which reset fall distance unconditionally, Scaffolding only does so for entities
	// that are sneaking (sourced from minecraft.wiki's Scaffolding and Tutorial:Breaking a fall pages, both of
	// which specifically say "while sneaking").
	sneaking, ok := e.(interface{ Sneaking() bool })
	if !ok || !sneaking.Sneaking() {
		return
	}
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
	// LavaFlammable is true on Bedrock (unlike Java, where scaffolding is not ignited by lava and can be placed
	// inside it - see the UseOnBlock check below for that part, which still applies on Bedrock too). Confirmed
	// directly in a Bedrock singleplayer world: placing scaffolding above lava visibly catches fire first
	// (Lava.RandomTick igniting a Fire block via the normal ignition mechanic, since scaffolding now counts as
	// lava-flammable), which then consumes it through the ordinary chance-based Fire.burn process - not an
	// instant, fire-less destruction.
	return newFlammabilityInfo(60, 60, true)
}

// FuelInfo ...
func (Scaffolding) FuelInfo() item.FuelInfo {
	// minecraft.wiki: Scaffolding smelts 0.25 items as furnace fuel, a quarter of Planks' fuel value per item
	// smelted (Planks smelts 1.5 items over 15 seconds elsewhere in this package, i.e. 10 seconds per item).
	// Other sources claimed Bedrock's value differs (up to 6x Java's), so this was verified directly in a
	// Bedrock singleplayer world rather than trusted from the wiki alone: each block burned for ~2-3 seconds,
	// confirming 0.25 items and ruling out the longer alternative.
	return newFuelInfo(time.Second * 5 / 2)
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

// scaffoldingStability returns the Stability and StabilityCheck a scaffolding block would have at the position
// passed, based on the block below it and any horizontally adjacent scaffolding. dist is 0 if the block rests on
// a supporting block and 7 if it is unsupported. bottom mirrors the StabilityCheck field: it is true only when
// dist is greater than 0 and the block below is not itself scaffolding, i.e. the block is floating.
func scaffoldingStability(pos cube.Pos, tx *world.Tx) (dist int, bottom bool) {
	below := pos.Side(cube.FaceDown)
	belowBlock := tx.Block(below)
	belowIsScaffolding := false
	dist = 7
	if s, ok := belowBlock.(Scaffolding); ok {
		dist = s.Stability
		belowIsScaffolding = true
	} else if belowBlock.Model().FaceSolid(below, cube.FaceUp, tx) {
		return 0, false
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
	return dist, dist > 0 && !belowIsScaffolding
}

// allScaffolding ...
func allScaffolding() (b []world.Block) {
	for stability := 0; stability <= 7; stability++ {
		b = append(b, Scaffolding{Stability: stability})
		b = append(b, Scaffolding{Stability: stability, StabilityCheck: true})
	}
	return
}
