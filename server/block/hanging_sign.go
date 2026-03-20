package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// HangingSign is a decorative sign block that may be attached to the side of a block or hung from its underside.
type HangingSign struct {
	transparent
	empty
	bass
	sourceWaterDisplacer

	// Wood is the type of wood of the hanging sign.
	Wood WoodType
	// Attach describes how the hanging sign is mounted.
	Attach HangingAttachment
	// Waxed specifies if the sign can no longer be edited.
	Waxed bool
	// Front is the text on the front side of the sign.
	Front SignText
	// Back is the text on the back side of the sign.
	Back SignText
}

// SideClosed reports that no face of a hanging sign fully closes off an adjacent block face.
func (h HangingSign) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// MaxCount returns the maximum number of hanging signs that may be stacked in one inventory slot.
func (h HangingSign) MaxCount() int {
	return 16
}

// FlammabilityInfo returns the flammability properties of the hanging sign.
func (h HangingSign) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// FuelInfo returns the furnace fuel properties of the hanging sign.
func (h HangingSign) FuelInfo() item.FuelInfo {
	if !h.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 10)
}

// EncodeItem encodes the hanging sign item name.
func (h HangingSign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + h.Wood.String() + "_hanging_sign", 0
}

// BreakInfo returns the breaking properties of the hanging sign.
func (h HangingSign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(HangingSign{Wood: h.Wood}))
}

// Dye dyes the HangingSign, changing its base colour to that of the colour passed.
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

// Ink inks the sign either glowing or non-glowing.
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

// Wax waxes a sign to prevent it from further editing.
func (h HangingSign) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if h.Waxed {
		return h, false
	}
	h.Waxed = true
	return h, true
}

// Activate opens the sign editor when the hanging sign is editable or plays the waxed interaction sound otherwise.
func (h HangingSign) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if editor, ok := u.(SignEditor); ok && !h.Waxed {
		editor.OpenSign(pos, h.EditingFrontSide(pos, u.Position()))
	} else if h.Waxed {
		tx.PlaySound(pos.Vec3(), sound.WaxedSignFailedInteraction{})
	}
	return true
}

// EditingFrontSide returns if the user is editing the front side of the sign based on their position relative to the sign.
func (h HangingSign) EditingFrontSide(pos cube.Pos, userPos mgl64.Vec3) bool {
	return userPos.Sub(pos.Vec3Centre()).Dot(h.Attach.Rotation().Vec3()) > 0
}

// UseOnBlock places the hanging sign either on the side of a block or underneath a supporting block.
func (h HangingSign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, h)
	if !used || face == cube.FaceUp {
		return false
	}
	switch face {
	case cube.FaceDown:
		supportPos := pos.Side(cube.FaceUp)
		support := tx.Block(supportPos)
		if !supportsCeilingHangingSign(support, supportPos, tx) {
			return false
		}

		rotation := user.Rotation()
		if supportsAttachedCeilingHangingSign(support) || sneaking(user) {
			h.Attach = AttachedCeilingHangingAttachment(rotation.Orientation().Opposite())
		} else {
			h.Attach = CeilingHangingAttachment(rotation.Direction().Opposite())
		}
	default:
		h.Attach = WallHangingAttachment(face.Opposite().Direction().RotateRight())
	}

	place(tx, pos, h, user, ctx)
	if editor, ok := user.(SignEditor); ok {
		editor.OpenSign(pos, true)
	}
	return placed(ctx)
}

// NeighbourUpdateTick breaks the hanging sign when the block supporting it is no longer valid.
func (h HangingSign) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if h.Attach.ceiling {
		supportPos := pos.Side(cube.FaceUp)
		if !supportsCeilingHangingSign(tx.Block(supportPos), supportPos, tx) {
			breakBlock(h, pos, tx)
		}
		return
	}

	supportFace := h.Attach.facing.RotateLeft().Face()
	supportPos := pos.Side(supportFace)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, supportFace.Opposite(), tx) {
		breakBlock(h, pos, tx)
	}
}

// EncodeBlock encodes the Bedrock block state of the hanging sign.
func (h HangingSign) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + h.Wood.String() + "_hanging_sign", map[string]any{
		"attached_bit":          boolByte(h.Attach.attached),
		"facing_direction":      int32(h.Attach.facing.Face()),
		"ground_sign_direction": int32(h.Attach.o),
		"hanging":               boolByte(h.Attach.ceiling),
	}
}

// DecodeNBT decodes block actor data for the hanging sign using the same text format as regular signs.
func (h HangingSign) DecodeNBT(data map[string]any) any {
	s := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}
	s = s.DecodeNBT(data).(Sign)
	h.Front, h.Back, h.Waxed = s.Front, s.Back, s.Waxed
	return h
}

// EncodeNBT encodes block actor data for the hanging sign.
func (h HangingSign) EncodeNBT() map[string]any {
	nbt := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}.EncodeNBT()
	nbt["id"] = "HangingSign"
	return nbt
}

// supportsCeilingHangingSign reports whether the block above can support a ceiling-hanging sign.
func supportsCeilingHangingSign(b world.Block, pos cube.Pos, tx *world.Tx) bool {
	if supportsAttachedCeilingHangingSign(b) {
		return true
	}
	return b.Model().FaceSolid(pos, cube.FaceDown, tx)
}

// supportsAttachedCeilingHangingSign reports whether the block above supports the attached chain variant.
func supportsAttachedCeilingHangingSign(b world.Block) bool {
	switch c := b.(type) {
	case IronChain:
		return c.Axis == cube.Y
	case CopperChain:
		return c.Axis == cube.Y
	default:
		return false
	}
}

// sneaking reports whether the user is currently sneaking.
func sneaking(u item.User) bool {
	s, ok := u.(interface{ Sneaking() bool })
	return ok && s.Sneaking()
}

// allHangingSigns returns all registered hanging sign permutations.
func allHangingSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for _, d := range cube.Directions() {
			signs = append(signs, HangingSign{Wood: w, Attach: WallHangingAttachment(d)})
			signs = append(signs, HangingSign{Wood: w, Attach: CeilingHangingAttachment(d)})
		}
		for o := cube.Orientation(0); o <= 15; o++ {
			signs = append(signs, HangingSign{Wood: w, Attach: AttachedCeilingHangingAttachment(o)})
		}
	}
	return
}
