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
	// Waxed specifies if the HangingSign has been waxed. If set to true, the sign can no longer be edited.
	Waxed bool
	// Front is the text of the front side.
	Front SignText
	// Back is the text of the back side.
	Back SignText
	// AttachedBit specifies if the hanging sign's chains are visually attached to the block above.
	AttachedBit bool
	// Hanging specifies if the sign is hanging from the ceiling (true) or mounted on a wall (false).
	Hanging bool
	// FacingDirection is the Minecraft block state facing direction (0-5). Relevant for wall-mounted signs.
	FacingDirection int
	// GroundSignDirection is the 16-step rotation of a ceiling-hung sign (0-15).
	GroundSignDirection int
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
		h.Hanging = true
		h.FacingDirection = int(cube.FaceDown)
		h.GroundSignDirection = int(user.Rotation().Orientation().Opposite())
	case cube.FaceUp:
		return false
	default:
		h.Hanging = false
		h.FacingDirection = int(face)
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
	return "minecraft:" + h.Wood.String() + "_hanging_sign", map[string]any{
		"attached_bit":          boolByte(h.AttachedBit),
		"facing_direction":      int32(h.FacingDirection),
		"ground_sign_direction": int32(h.GroundSignDirection),
		"hanging":               boolByte(h.Hanging),
	}
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
	return userPos.Sub(pos.Vec3Centre()).Dot(h.rotation().Vec3()) > 0
}

// rotation returns the facing rotation of the hanging sign for sign text side detection.
func (h HangingSign) rotation() cube.Rotation {
	if h.Hanging {
		return cube.Rotation{cube.Orientation(h.GroundSignDirection).Yaw()}
	}
	var yaw float64
	switch cube.Face(h.FacingDirection) {
	case cube.FaceWest:
		yaw = 90
	case cube.FaceEast:
		yaw = -90
	case cube.FaceNorth:
		yaw = 180
	}
	return cube.Rotation{yaw}
}

// allHangingSigns returns a list of all hanging sign block states.
// Minecraft registers all 384 combinations per wood type:
// attached_bit (0-1) × facing_direction (0-5) × ground_sign_direction (0-15) × hanging (0-1).
func allHangingSigns() (signs []world.Block) {
	for _, w := range WoodTypes() {
		for _, attached := range []bool{false, true} {
			for _, hanging := range []bool{false, true} {
				for facing := 0; facing <= 5; facing++ {
					for groundDir := 0; groundDir <= 15; groundDir++ {
						signs = append(signs, HangingSign{
							Wood:                w,
							AttachedBit:         attached,
							Hanging:             hanging,
							FacingDirection:     facing,
							GroundSignDirection: groundDir,
						})
					}
				}
			}
		}
	}
	return
}
