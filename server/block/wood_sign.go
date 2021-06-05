package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"strings"
)

// WoodSign is block that you can write text on.
type WoodSign struct {
	transparent
	bass //might be wrong

	// Wood is the type of wood of the door. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Facing is the direction the sign is facing.
	Facing cube.Direction
	// Standing is whether the sign is on a block or against
	Standing bool
	// Text is an array with the text
	Text string
	// SignTextColor Color the sign is dyed witth
	SignTextColor int32
	// TextOwner XUID of the player who placed the sign
	TextOwner string
}

// FlammabilityInfo ...
func (s WoodSign) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, false)
}

// Model ...
func (s WoodSign) Model() world.BlockModel {
	return model.Empty{}
}

// BreakInfo ...
func (s WoodSign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(s))
}

// CanDisplace ...
func (s WoodSign) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

func (s WoodSign) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos = pos.Side(face)
	if face != cube.FaceDown {

		if face == cube.FaceUp {
			s.Standing = true
			s.Facing = cube.Direction(int16(math.Floor((((user.Yaw()+180)*16/360)+0.5)*100)/100) & 15)
		} else {
			s.Facing = cube.Direction(face)
		}
		place(w, pos, s, user, ctx)
	}
	return placed(ctx)
}

// DecodeNBT ...
func (s WoodSign) DecodeNBT(data map[string]interface{}) interface{} {
	return WoodSign{Wood: s.Wood, Facing: s.Facing, Standing: s.Standing,
		Text:          data["Text"].(string),
		SignTextColor: data["SignTextColor"].(int32),
		TextOwner:     data["TextOwner"].(string),
	}
}

// EncodeNBT ...
func (s WoodSign) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{
		"id":            "Sign",
		"Text":          s.Text,
		"SignTextColor": s.SignTextColor,
		"TextOwner":     s.TextOwner,
	}
}

// EncodeItem ...
func (s WoodSign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_sign", 0
}

// EncodeBlock ...
func (s WoodSign) EncodeBlock() (name string, properties map[string]interface{}) {
	woodType := strings.Replace(s.Wood.String(), "_", "", 1)
	if woodType == "oak" {
		woodType = "" // oak signs = wall sign|standing_sign
	} else {
		woodType = woodType + "_"
	}

	if s.Standing {
		return "minecraft:" + woodType + "standing_sign", map[string]interface{}{"ground_sign_direction": int32(s.Facing)}
	}

	return "minecraft:" + woodType + "wall_sign", map[string]interface{}{"facing_direction": int32(s.Facing)}
}

func allSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for i := cube.Direction(0); i <= 15; i++ {
			signs = append(signs, WoodSign{Wood: w, Facing: i, Standing: true})
		}
		for i := cube.Direction(0); i <= 5; i++ {
			signs = append(signs, WoodSign{Wood: w, Facing: i, Standing: false})
		}
	}
	return
}
