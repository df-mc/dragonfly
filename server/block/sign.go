package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
	"time"
)

// Sign is a non-solid block that can display text on the front and back of the block.
type Sign struct {
	transparent
	empty
	bass
	sourceWaterDisplacer

	// Wood is the type of wood of the sign. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Attach is the attachment of the Sign. It is either of the type WallAttachment or StandingAttachment.
	Attach Attachment
	// Waxed specifies if the Sign has been waxed by a player. If set to true, the Sign can no longer be edited by
	// anyone and must be destroyed if the text needs to be changed.
	Waxed bool
	// Front is the text of the front side of the sign. Anyone can edit this unless the sign is Waxed.
	Front SignText
	// Back is the text of the back side of the sign. Anyone can edit this unless the sign is Waxed.
	Back SignText
}

// SignText represents the data for a single side of a sign. The sign can be edited on the front and back side.
type SignText struct {
	// Text is the text displayed on this side of the sign. The text is automatically wrapped if it does not fit on a line.
	Text string
	// BaseColour is the base colour of the text on this side of the sign, changed when using a dye on the sign. The default
	// colour is black.
	BaseColour color.RGBA
	// Glowing specifies if the Sign has glowing text on the current side. If set to true, the text will be visible even
	// in the dark, and it will have an outline to improve visibility.
	Glowing bool
	// Owner holds the XUID of the player that most recently edited this side of the sign.
	Owner string
}

// SideClosed ...
func (s Sign) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// MaxCount ...
func (s Sign) MaxCount() int {
	return 16
}

// FlammabilityInfo ...
func (s Sign) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// FuelInfo ...
func (s Sign) FuelInfo() item.FuelInfo {
	if !s.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 10)
}

// EncodeItem ...
func (s Sign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_sign", 0
}

// BreakInfo ...
func (s Sign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(Sign{Wood: s.Wood}))
}

// Dye dyes the Sign, changing its base colour to that of the colour passed.
func (s Sign) Dye(pos cube.Pos, userPos mgl64.Vec3, c item.Colour) (world.Block, bool) {
	if s.EditingFrontSide(pos, userPos) {
		if s.Front.BaseColour == c.SignRGBA() {
			return s, false
		}
		s.Front.BaseColour = c.SignRGBA()
	} else {
		if s.Back.BaseColour == c.SignRGBA() {
			return s, false
		}
		s.Back.BaseColour = c.SignRGBA()
	}
	return s, true
}

// Ink inks the sign either glowing or non-glowing.
func (s Sign) Ink(pos cube.Pos, userPos mgl64.Vec3, glowing bool) (world.Block, bool) {
	if s.EditingFrontSide(pos, userPos) {
		if s.Front.Glowing == glowing {
			return s, false
		}
		s.Front.Glowing = glowing
	} else {
		if s.Back.Glowing == glowing {
			return s, false
		}
		s.Back.Glowing = glowing
	}
	return s, true
}

// Wax waxes a sign to prevent it from further editing.
func (s Sign) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if s.Waxed {
		return s, false
	}
	s.Waxed = true
	return s, true
}

// Activate ...
func (s Sign) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if editor, ok := u.(SignEditor); ok && !s.Waxed {
		editor.OpenSign(pos, s.EditingFrontSide(pos, u.Position()))
	} else if s.Waxed {
		tx.PlaySound(pos.Vec3(), sound.WaxedSignFailedInteraction{})
	}
	return true
}

// EditingFrontSide returns if the user is editing the front side of the sign based on their position relative to the
// position and direction of the sign.
func (s Sign) EditingFrontSide(pos cube.Pos, userPos mgl64.Vec3) bool {
	return userPos.Sub(pos.Vec3Centre()).Dot(s.Attach.Rotation().Vec3()) > 0
}

// SignEditor represents something that can edit a sign, typically players.
type SignEditor interface {
	OpenSign(pos cube.Pos, frontSide bool)
}

// UseOnBlock ...
func (s Sign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, s)
	if !used || face == cube.FaceDown {
		return false
	}

	if face == cube.FaceUp {
		s.Attach = StandingAttachment(user.Rotation().Orientation().Opposite())
	} else {
		s.Attach = WallAttachment(face.Direction())
	}
	place(tx, pos, s, user, ctx)
	if editor, ok := user.(SignEditor); ok {
		editor.OpenSign(pos, true)
	}
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s Sign) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if s.Attach.hanging {
		if _, ok := tx.Block(pos.Side(s.Attach.facing.Opposite().Face())).(Air); ok {
			breakBlock(s, pos, tx)
		}
	} else if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Air); ok {
		breakBlock(s, pos, tx)
	}
}

// EncodeBlock ...
func (s Sign) EncodeBlock() (name string, properties map[string]any) {
	woodType := s.Wood.String() + "_"
	switch s.Wood {
	case OakWood():
		woodType = ""
	case DarkOakWood():
		woodType = "darkoak_"
	}
	if s.Attach.hanging {
		return "minecraft:" + woodType + "wall_sign", map[string]any{"facing_direction": int32(s.Attach.facing + 2)}
	}
	return "minecraft:" + woodType + "standing_sign", map[string]any{"ground_sign_direction": int32(s.Attach.o)}
}

// DecodeNBT ...
func (s Sign) DecodeNBT(data map[string]any) any {
	if nbtconv.String(data, "Text") != "" {
		// The NBT format changed in 1.19.80 to have separate data for each side of the sign. The old format must still
		// be supported for backwards compatibility.
		s.Front.Text = nbtconv.String(data, "Text")
		s.Front.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Int32(data, "SignTextColor"))
		s.Front.Glowing = nbtconv.Bool(data, "IgnoreLighting") && nbtconv.Bool(data, "TextIgnoreLegacyBugResolved")
		return s
	}

	front, ok := data["FrontText"].(map[string]any)
	if ok {
		s.Front.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Int32(front, "Color"))
		s.Front.Glowing = nbtconv.Bool(front, "GlowingText")
		s.Front.Text = nbtconv.String(front, "Text")
		s.Front.Owner = nbtconv.String(front, "Owner")
	}

	back, ok := data["BackText"].(map[string]any)
	if ok {
		s.Back.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Int32(back, "Color"))
		s.Back.Glowing = nbtconv.Bool(back, "GlowingText")
		s.Back.Text = nbtconv.String(back, "Text")
		s.Back.Owner = nbtconv.String(back, "Owner")
	}

	return s
}

// EncodeNBT ...
func (s Sign) EncodeNBT() map[string]any {
	m := map[string]any{
		"id":      "Sign",
		"IsWaxed": boolByte(s.Waxed),
		"FrontText": map[string]any{
			"SignTextColor":  nbtconv.Int32FromRGBA(s.Front.BaseColour),
			"IgnoreLighting": boolByte(s.Front.Glowing),
			"Text":           s.Front.Text,
			"TextOwner":      s.Front.Owner,
		},
		"BackText": map[string]any{
			"SignTextColor":  nbtconv.Int32FromRGBA(s.Back.BaseColour),
			"IgnoreLighting": boolByte(s.Back.Glowing),
			"Text":           s.Back.Text,
			"TextOwner":      s.Back.Owner,
		},
	}
	return m
}

// allSigns ...
func allSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for _, d := range cube.Directions() {
			signs = append(signs, Sign{Wood: w, Attach: WallAttachment(d)})
		}
		for o := cube.Orientation(0); o <= 15; o++ {
			signs = append(signs, Sign{Wood: w, Attach: StandingAttachment(o)})
		}
	}
	return
}
