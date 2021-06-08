package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodSign is block that you can write text on.
type WoodSign struct {
	transparent
	bass //might be wrong

	// Wood is the type of wood of the sign. This field must have one of the values found in the material
	// package.
	Wood WoodType
}

// EncodeItem ...
func (s WoodSign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_sign", 0
}

func (s WoodSign) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos = pos.Side(face)
	if face != cube.FaceDown {
		if face == cube.FaceUp {
			if replaceableWith(w, pos, StandingSign{}) {
				block := StandingSign{Wood: s.Wood, Orientation: cube.YawToOrientation(user.Yaw() + 180)}
				place(w, pos, block, user, ctx)
			}
		} else {
			block := WallSign{Wood: s.Wood, Facing: face}
			place(w, pos, block, user, ctx)
		}
	}
	return placed(ctx)
}

func allSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for i := cube.Face(2); i <= 5; i++ {
			signs = append(signs, WallSign{Wood: w, Facing: i})
		}
		for i := cube.Orientation(0); i <= 15; i++ {
			signs = append(signs, StandingSign{Wood: w, Orientation: i})
		}
	}
	return
}
