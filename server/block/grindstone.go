package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Grindstone is a block that repairs items and tools as well as removing enchantments from them. It also serves as a
// weaponsmith's job site block.
type Grindstone struct {
	transparent

	// Attach represents the attachment type of the Grindstone.
	Attach GrindstoneAttachment
	// Facing represents the direction the Grindstone is facing.
	Facing cube.Direction
}

// BreakInfo ...
func (g Grindstone) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(g)).withBlastResistance(30)
}

// Activate ...
func (g Grindstone) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// UseOnBlock ...
func (g Grindstone) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, g)
	if !used {
		return false
	}
	g.Facing = user.Rotation().Direction().Opposite()
	if face == cube.FaceDown {
		g.Attach = HangingGrindstoneAttachment()
	} else if face != cube.FaceUp {
		g.Attach = WallGrindstoneAttachment()
		g.Facing = face.Direction()
	}
	place(tx, pos, g, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (g Grindstone) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportFace := g.Facing.Face().Opposite()
	if g.Attach == HangingGrindstoneAttachment() {
		supportFace = cube.FaceUp
	} else if g.Attach == StandingGrindstoneAttachment() {
		supportFace = cube.FaceDown
	}
	if _, ok := tx.Block(pos.Side(supportFace)).Model().(model.Empty); ok {
		// Grindstone is pickaxeHarvestable, so don't use breakBlock() here.
		breakBlockNoDrops(g, pos, tx)
		dropItem(tx, item.NewStack(g, 1), pos.Vec3Centre())
	}
}

// Model ...
func (g Grindstone) Model() world.BlockModel {
	axis := cube.Y
	if g.Attach == WallGrindstoneAttachment() {
		axis = g.Facing.Face().Axis()
	}
	return model.Grindstone{Axis: axis}
}

// EncodeBlock ...
func (g Grindstone) EncodeBlock() (string, map[string]any) {
	return "minecraft:grindstone", map[string]any{
		"attachment": g.Attach.String(),
		"direction":  int32(horizontalDirection(g.Facing)),
	}
}

// EncodeItem ...
func (g Grindstone) EncodeItem() (name string, meta int16) {
	return "minecraft:grindstone", 0
}

// allGrindstones ...
func allGrindstones() (grindstones []world.Block) {
	for _, a := range GrindstoneAttachments() {
		for _, d := range cube.Directions() {
			grindstones = append(grindstones, Grindstone{Attach: a, Facing: d})
		}
	}
	return
}
