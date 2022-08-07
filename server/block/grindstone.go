package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Grindstone is a block that repairs items and tools as well as removing enchantments from them. It also serves as a
// weaponsmith's job site block.
type Grindstone struct {
	transparent

	// Attach represents the attachment type of the Grindstone.
	Attach GrindstoneAttachment
	// Direction represents the direction the Grindstone is facing.
	Direction cube.Direction
}

// BreakInfo ...
func (g Grindstone) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(g)).withBlastResistance(30)
}

// Activate ...
func (g Grindstone) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// UseOnBlock ...
func (g Grindstone) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, g)
	if !used {
		return false
	}
	if face == cube.FaceUp {
		g.Direction = user.Facing().Opposite()
	} else if face == cube.FaceDown {
		g.Direction = user.Facing().Opposite()
		g.Attach = HangingGrindstoneAttachment()
	} else {
		g.Attach = WallGrindstoneAttachment()
		g.Direction = face.Direction()
	}
	place(w, pos, g, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (g Grindstone) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	supportFace := g.Direction.Face().Opposite()
	if g.Attach == HangingGrindstoneAttachment() {
		supportFace = cube.FaceUp
	} else if g.Attach == StandingGrindstoneAttachment() {
		supportFace = cube.FaceDown
	}
	if _, ok := w.Block(pos.Side(supportFace)).(Air); ok {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: g})
	}
}

// Model ...
func (g Grindstone) Model() world.BlockModel {
	var axis cube.Axis
	if g.Attach == WallGrindstoneAttachment() {
		axis = g.Direction.Face().Axis()
	} else {
		axis = cube.Y
	}
	return model.Grindstone{Axis: axis}
}

// EncodeBlock ...
func (g Grindstone) EncodeBlock() (string, map[string]any) {
	return "minecraft:grindstone", map[string]any{
		"attachment": g.Attach.String(),
		"direction":  int32(horizontalDirection(g.Direction)),
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
			grindstones = append(grindstones, Grindstone{Attach: a, Direction: d})
		}
	}
	return
}
