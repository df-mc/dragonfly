package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"image/color"
	"strings"
)

// Sign is a non-solid block that can display text.
type Sign struct {
	transparent
	empty
	bass

	// Wood is the type of wood of the sign. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Attach is the attachment of the Sign. It is either of the type WallAttachment or StandingAttachment.
	Attach Attachment
	// Text is the text displayed on the sign. The text is automatically wrapped if it does not fit on a line.
	Text string
	// BaseColour is the base colour of the text on the sign, changed when using a dye on the sign. The default colour
	// is black.
	BaseColour color.RGBA
	// Glowing specifies if the Sign has glowing text. If set to true, the text will be visible even in the dark, and it
	// will have an outline to improve visibility.
	Glowing bool
	// owner holds the UUID of the player that initially placed the sign.
	owner uuid.UUID
}

// SideClosed ...
func (s Sign) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
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

// EncodeItem ...
func (s Sign) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_sign", 0
}

// BreakInfo ...
func (s Sign) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(s))
}

// CanDisplace ...
func (s Sign) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// Dye dyes the Sign, changing its base colour to that of the colour passed.
func (s Sign) Dye(c item.Colour) (world.Block, bool) {
	if s.BaseColour == c.RGBA() {
		return s, false
	}
	s.BaseColour = c.RGBA()
	return s, true
}

// Ink inks the sign either glowing or non-glowing.
func (s Sign) Ink(glowing bool) (world.Block, bool) {
	if s.Glowing == glowing {
		return s, false
	}
	s.Glowing = glowing
	return s, true
}

// SignEditor represents something that can edit a sign, typically players.
type SignEditor interface {
	UUID() uuid.UUID
}

// EditableBy returns whether a SignEditor can edit the sign or not. This is based on whether the SignEditor
// placed the sign and the sign's chunk has yet to be unloaded.
func (s Sign) EditableBy(editor SignEditor) bool {
	return editor.UUID() == s.owner
}

// UseOnBlock ...
func (s Sign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used || face == cube.FaceDown {
		return false
	}

	if editor, ok := user.(SignEditor); ok {
		s.owner = editor.UUID()
	}

	if face == cube.FaceUp {
		yaw, _ := user.Rotation()
		s.Attach = StandingAttachment(cube.OrientationFromYaw(yaw).Opposite())
		place(w, pos, s, user, ctx)
		return
	}
	s.Attach = WallAttachment(face.Direction())
	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s Sign) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if s.Attach.hanging {
		if _, ok := w.Block(pos.Side(s.Attach.facing.Opposite().Face())).(Air); ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
		}
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
	}
}

// EncodeBlock ...
func (s Sign) EncodeBlock() (name string, properties map[string]any) {
	woodType := strings.Replace(s.Wood.String(), "_", "", 1) + "_"
	if woodType == "oak_" {
		woodType = ""
	}
	if s.Attach.hanging {
		return "minecraft:" + woodType + "wall_sign", map[string]any{"facing_direction": int32(s.Attach.facing + 2)}
	}
	return "minecraft:" + woodType + "standing_sign", map[string]any{"ground_sign_direction": int32(s.Attach.o)}
}

// DecodeNBT ...
func (s Sign) DecodeNBT(data map[string]any) any {
	s.Text = nbtconv.Map[string](data, "Text")
	s.BaseColour = nbtconv.RGBAFromInt32(nbtconv.Map[int32](data, "SignTextColor"))
	s.Glowing = nbtconv.Map[byte](data, "IgnoreLighting") == 1 && nbtconv.Map[byte](data, "TextIgnoreLegacyBugResolved") == 1

	return s
}

// EncodeNBT ...
func (s Sign) EncodeNBT() map[string]any {
	m := map[string]any{
		"id":             "Sign",
		"SignTextColor":  nbtconv.Int32FromRGBA(s.BaseColour),
		"IgnoreLighting": boolByte(s.Glowing),
		// This is some top class Mojang garbage. The client needs it to render the glowing text. Omitting this field
		// will just result in normal text being displayed.
		"TextIgnoreLegacyBugResolved": boolByte(s.Glowing),
	}
	if s.Text != "" {
		// The client does not display the editing GUI if this tag is already set when no text is present, so just don't
		// send it while the text is empty.
		m["Text"] = s.Text
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
