package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// HangingSign is a non-solid block that can display text and can be hung from the underside of blocks.
type HangingSign struct {
	transparent
	empty
	bass
	sourceWaterDisplacer

	// Wood is the type of wood of the hanging sign.
	Wood WoodType
	// Attach is the attachment of the HangingSign. It uses the same Attachment type as Sign.
	Attach Attachment
	// Waxed specifies if the HangingSign has been waxed. If set to true, the sign can no longer be edited.
	Waxed bool
	// Front is the text of the front side.
	Front SignText
	// Back is the text of the back side.
	Back SignText
}

// SideClosed ...
func (HangingSign) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// MaxCount ...
func (HangingSign) MaxCount() int {
	return 16
}

// FlammabilityInfo ...
func (h HangingSign) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// FuelInfo ...
func (h HangingSign) FuelInfo() item.FuelInfo {
	if !h.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 10)
}

// EncodeItem ...
func (h HangingSign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + h.Wood.String() + "_hanging_sign", 0
}

// BreakInfo ...
func (h HangingSign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(HangingSign{Wood: h.Wood}))
}

// Dye dyes the HangingSign, changing its base colour.
func (h HangingSign) Dye(pos cube.Pos, userPos mgl64.Vec3, c item.Colour) (world.Block, bool) {
	if h.EditingFrontSide(pos, userPos) {
		if h.Front.BaseColour == c.SignRGBA() {
			return h, false
		}
		h.Front.BaseColour = c.SignRGBA()
	} else {
		if h.Back.BaseColour == c.SignRGBA() {
			return h, false
		}
		h.Back.BaseColour = c.SignRGBA()
	}
	return h, true
}

// Ink inks the hanging sign.
func (h HangingSign) Ink(pos cube.Pos, userPos mgl64.Vec3, glowing bool) (world.Block, bool) {
	if h.EditingFrontSide(pos, userPos) {
		if h.Front.Glowing == glowing {
			return h, false
		}
		h.Front.Glowing = glowing
	} else {
		if h.Back.Glowing == glowing {
			return h, false
		}
		h.Back.Glowing = glowing
	}
	return h, true
}

// Wax waxes a hanging sign.
func (h HangingSign) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if h.Waxed {
		return h, false
	}
	h.Waxed = true
	return h, true
}

// UseOnBlock places the hanging sign.
func (h HangingSign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, h)
	if !used {
		return false
	}
	switch face {
	case cube.FaceDown:
		h.Attach = StandingAttachment(user.Rotation().Orientation().Opposite())
	case cube.FaceUp:
		return false
	default:
		h.Attach = WallAttachment(face.Direction())
	}
	place(tx, pos, h, user, ctx)
	if editor, ok := user.(SignEditor); ok {
		editor.OpenSign(pos, true)
	}
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (h HangingSign) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceUp)).(Air); ok {
		breakBlock(h, pos, tx)
	}
}

// EncodeBlock ...
func (h HangingSign) EncodeBlock() (name string, properties map[string]any) {
	woodType := h.Wood.String() + "_"
	if h.Attach.hanging {
		return "minecraft:" + woodType + "wall_hanging_sign", map[string]any{"facing_direction": int32(h.Attach.facing + 2)}
	}
	return "minecraft:" + woodType + "hanging_sign", map[string]any{"ground_sign_direction": int32(h.Attach.o)}
}

// DecodeNBT ...
func (h HangingSign) DecodeNBT(data map[string]any) any {
	front, ok := data["FrontText"].(map[string]any)
	if ok {
		h.Front.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Int32(front, "Color"))
		h.Front.Glowing = nbtconv.Bool(front, "GlowingText")
		h.Front.Text = nbtconv.String(front, "Text")
		h.Front.Owner = nbtconv.String(front, "Owner")
	}
	back, ok := data["BackText"].(map[string]any)
	if ok {
		h.Back.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Int32(back, "Color"))
		h.Back.Glowing = nbtconv.Bool(back, "GlowingText")
		h.Back.Text = nbtconv.String(back, "Text")
		h.Back.Owner = nbtconv.String(back, "Owner")
	}
	h.Waxed = nbtconv.Bool(data, "IsWaxed")
	return h
}

// EncodeNBT ...
func (h HangingSign) EncodeNBT() map[string]any {
	return map[string]any{
		"id":      "HangingSign",
		"IsWaxed": boolByte(h.Waxed),
		"FrontText": map[string]any{
			"SignTextColor":  nbtconv.Int32FromRGBA(h.Front.BaseColour),
			"IgnoreLighting": boolByte(h.Front.Glowing),
			"Text":           h.Front.Text,
			"TextOwner":      h.Front.Owner,
		},
		"BackText": map[string]any{
			"SignTextColor":  nbtconv.Int32FromRGBA(h.Back.BaseColour),
			"IgnoreLighting": boolByte(h.Back.Glowing),
			"Text":           h.Back.Text,
			"TextOwner":      h.Back.Owner,
		},
	}
}

// EditingFrontSide reports whether the user is editing the front side of the sign.
func (h HangingSign) EditingFrontSide(pos cube.Pos, userPos mgl64.Vec3) bool {
	return userPos.Sub(pos.Vec3Centre()).Dot(h.Attach.Rotation().Vec3()) > 0
}

// allHangingSigns returns a list of all hanging sign types.
func allHangingSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for o := cube.Orientation(0); o <= 15; o++ {
			signs = append(signs, HangingSign{Wood: w, Attach: StandingAttachment(o)})
		}
		for _, d := range cube.Directions() {
			signs = append(signs, HangingSign{Wood: w, Attach: WallAttachment(d)})
		}
	}
	return
}
