package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"strings"
)

type StandingSign struct {
	transparent
	bass //might be wrong

	// Wood is the type of wood of the sign. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Facing is the direction the sign is facing.
	Orientation cube.Orientation
	// Text is a string which is displayed on the sign
	Text string
	// SignTextColor Color the sign is dyed with
	SignTextColor int32
	// TextOwner XUID of the player who placed the sign
	TextOwner string
}

// FlammabilityInfo ...
func (s StandingSign) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, false)
}

// Model ...
func (s StandingSign) Model() world.BlockModel {
	return model.Empty{}
}

// EncodeItem ...
func (s StandingSign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_sign", 0
}

// BreakInfo ...
func (s StandingSign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(s))
}

// CanDisplace ...
func (s StandingSign) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// EncodeBlock ...
func (s StandingSign) EncodeBlock() (name string, properties map[string]interface{}) {
	woodType := strings.Replace(s.Wood.String(), "_", "", 1)
	if woodType == "oak" {
		woodType = "" // oak signs = wall sign|standing_sign
	} else {
		woodType = woodType + "_"
	}

	return "minecraft:" + woodType + "standing_sign", map[string]interface{}{"ground_sign_direction": int32(s.Orientation)}
}

// DecodeNBT ...
func (s StandingSign) DecodeNBT(data map[string]interface{}) interface{} {
	s.Text = readString(data, "Text")
	s.SignTextColor = readInt32(data, "SignTextColor")
	s.TextOwner = readString(data, "TextOwner")

	return s
}

// EncodeNBT ...
func (s StandingSign) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{
		"id":            "Sign",
		"Text":          s.Text,
		"SignTextColor": s.SignTextColor,
		"TextOwner":     s.TextOwner,
	}
}

func (s StandingSign) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, w *world.World) {
	if _, ok := w.Block(pos).(StandingSign); ok {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
			w.BreakBlock(pos)
		}
	}
}
