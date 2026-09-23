package block

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
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
	// Attach describes how the hanging sign is mounted.
	Attach HangingAttachment
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

// Ink inks the hanging sign either glowing or non-glowing.
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

// Wax waxes a hanging sign to prevent it from further editing.
func (h HangingSign) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if h.Waxed {
		return h, false
	}
	h.Waxed = true
	return h, true
}

// UseOnBlock ...
func (h HangingSign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, h)
	if !used {
		return false
	}
	attach, ok := attachment(pos, face, tx, placerOf(user))
	if !ok {
		return false
	}
	h.Attach = attach

	place(tx, pos, h, user, ctx)
	if !placed(ctx) {
		return false
	}
	if editor, ok := user.(SignEditor); ok {
		editor.OpenSign(pos, true)
	}
	return true
}

type placer struct {
	o            cube.Orientation
	dir          cube.Direction
	lookX, lookZ float64
	sneaking     bool
}

func placerOf(user item.User) placer {
	s, ok := user.(interface{ Sneaking() bool })
	return newPlacer(user.Rotation().Yaw(), ok && s.Sneaking())
}

// newPlacer rounds a yaw halfway between two orientations or directions up, as vanilla does.
func newPlacer(yaw float64, sneaking bool) placer {
	rad := mgl64.DegToRad(yaw)
	yaw += 180
	return placer{
		o:        cube.Orientation(int(math.Floor(yaw*16/360+0.5)) & 15),
		dir:      [...]cube.Direction{cube.South, cube.West, cube.North, cube.East}[int(math.Floor(yaw*4/360+0.5))&3],
		lookX:    -math.Sin(rad),
		lookZ:    math.Cos(rad),
		sneaking: sneaking,
	}
}

// attachment only mounts a sign on the block clicked when a side is clicked, never falling back to anything else.
func attachment(pos cube.Pos, face cube.Face, tx *world.Tx, p placer) (HangingAttachment, bool) {
	if face != cube.FaceDown && face != cube.FaceUp {
		return wallAttachmentOn(pos, face.Opposite().Direction(), tx, p)
	}
	if a, ok := ceilingAttachment(pos, tx, p); ok {
		return a, true
	}
	return wallAttachment(pos, tx, p)
}

func ceilingAttachment(pos cube.Pos, tx *world.Tx, p placer) (HangingAttachment, bool) {
	above := pos.Side(cube.FaceUp)
	b := tx.Block(above)
	kind := hangingSupport(b, above, tx)
	if kind == noSupport {
		return HangingAttachment{}, false
	}
	if straightHangs(b, kind, p) {
		return CeilingHangingAttachment(p.dir), true
	}
	return AttachedCeilingHangingAttachment(p.o), true
}

func wallAttachment(pos cube.Pos, tx *world.Tx, p placer) (HangingAttachment, bool) {
	d := p.dir.Opposite()
	for range cube.Directions() {
		if a, ok := wallAttachmentOn(pos, d, tx, p); ok {
			return a, true
		}
		d = d.RotateRight()
	}
	return HangingAttachment{}, false
}

func wallAttachmentOn(pos cube.Pos, d cube.Direction, tx *world.Tx, p placer) (HangingAttachment, bool) {
	if !wallHangs(tx, pos.Side(d.Face()), d.Opposite().Face()) {
		return HangingAttachment{}, false
	}
	// Looking straight along the bar, the panel turns south or east.
	if d.Face().Axis() == cube.X {
		if p.lookZ > lookEpsilon {
			return WallHangingAttachment(cube.North), true
		}
		return WallHangingAttachment(cube.South), true
	}
	if p.lookX > lookEpsilon {
		return WallHangingAttachment(cube.West), true
	}
	return WallHangingAttachment(cube.East), true
}

const lookEpsilon = 1e-6

// Activate ...
func (h HangingSign) Activate(pos cube.Pos, face cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	// Holding a hanging sign only opens the editor where it cannot be hung.
	if held, _ := u.HeldItems(); !held.Empty() {
		if s, ok := held.Item().(HangingSign); ok {
			if s.placeable(pos, face, tx, placerOf(u)) {
				return false
			}
		} else if _, ok := held.Item().(world.Block); ok && !h.textFace(face) {
			return false
		}
	}
	if editor, ok := u.(SignEditor); ok && !h.Waxed {
		editor.OpenSign(pos, h.EditingFrontSide(pos, u.Position()))
	} else if h.Waxed {
		tx.PlaySound(pos.Vec3(), sound.WaxedSignFailedInteraction{})
	}
	return true
}

// NeighbourUpdateTick ...
func (h HangingSign) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	// Vanilla never checks a wall mounted sign again after placing it.
	if !h.Attach.ceiling {
		return
	}
	above := pos.Side(cube.FaceUp)
	if hangingSupport(tx.Block(above), above, tx) == noSupport {
		breakBlock(h, pos, tx)
	}
}

// EncodeBlock ...
func (h HangingSign) EncodeBlock() (name string, properties map[string]any) {
	var facing, ground int32
	if h.Attach.attached {
		ground = int32(h.Attach.o)
	} else {
		facing = int32(h.Attach.facing.Face())
	}
	return "minecraft:" + h.Wood.String() + "_hanging_sign", map[string]any{
		"attached_bit":          boolByte(h.Attach.attached),
		"facing_direction":      facing,
		"ground_sign_direction": ground,
		"hanging":               boolByte(h.Attach.ceiling),
	}
}

// DecodeNBT ...
func (h HangingSign) DecodeNBT(data map[string]any) any {
	s := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}.DecodeNBT(data).(Sign)
	h.Front, h.Back, h.Waxed = s.Front, s.Back, s.Waxed
	return h
}

// EncodeNBT ...
func (h HangingSign) EncodeNBT() map[string]any {
	nbt := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}.EncodeNBT()
	nbt["id"] = "HangingSign"
	return nbt
}

func (h HangingSign) placeable(pos cube.Pos, face cube.Face, tx *world.Tx, p placer) bool {
	at, face, used := firstReplaceable(tx, pos, face, h)
	if !used {
		return false
	}
	_, ok := attachment(at, face, tx, p)
	return ok
}

func (h HangingSign) textFace(face cube.Face) bool {
	f := h.Attach.Rotation().Direction().Face()
	return face == f || face == f.Opposite()
}

// EditingFrontSide returns whether the user is editing the front side of the sign.
func (h HangingSign) EditingFrontSide(pos cube.Pos, userPos mgl64.Vec3) bool {
	return userPos.Sub(pos.Vec3Centre()).Dot(h.Attach.Rotation().Vec3()) > 0
}

// straightHangs reports straight chains rather than a V. Crouching and blocks carrying only at their centre give
// a V, and below another sign both signs have to run along the same axis.
func straightHangs(b world.Block, kind support, p placer) bool {
	if p.sneaking {
		return false
	}
	hs, ok := b.(HangingSign)
	if !ok {
		return kind == fullSupport
	}
	if !hs.Attach.attached {
		return hs.Attach.facing.Face().Axis() == p.dir.Face().Axis()
	}
	return p.o%4 == 0 && hs.Attach.o%4 == 0 && hs.Attach.o%8 == p.o%8
}

type support uint8

const (
	noSupport support = iota
	centreSupport
	fullSupport
)

func hangingSupport(b world.Block, pos cube.Pos, tx *world.Tx) support {
	if s, ok := classSupport(b); ok {
		return s
	}
	if b.Model().FaceSolid(pos, cube.FaceDown, tx) {
		return fullSupport
	}
	return noSupport
}

// classSupport covers the blocks whose vanilla class overrides canProvideSupport. None of them carry a sign on
// their sides.
func classSupport(b world.Block) (support, bool) {
	switch b := b.(type) {
	case HangingSign, WoodFence, NetherBrickFence, Wall, IronBars, CopperBars, GlassPane,
		Candle, Anvil, EnderChest, Chest, BrewingStand, Lectern:
		return centreSupport, true
	case IronChain:
		return verticalSupport(b.Axis), true
	case CopperChain:
		return verticalSupport(b.Axis), true
	case EndRod:
		return verticalSupport(b.Facing.Axis()), true
	case Skull:
		if b.Attach.hanging {
			return noSupport, true
		}
		return centreSupport, true
	case WoodTrapdoor:
		if b.Top || b.Open {
			return noSupport, true
		}
		return centreSupport, true
	case Stonecutter, EnchantingTable, Campfire, Farmland, DirtPath:
		return fullSupport, true
	case Hopper, Leaves, Torch, CopperTorch, Lantern, CopperLantern:
		return noSupport, true
	}
	return noSupport, false
}

func verticalSupport(axis cube.Axis) support {
	if axis == cube.Y {
		return centreSupport
	}
	return noSupport
}

func wallHangs(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	b := tx.Block(pos)
	if hs, ok := b.(HangingSign); ok {
		// Vanilla checks a list of hanging signs that is missing poplar.
		return !hs.Attach.ceiling && hs.Wood != PoplarWood() && hs.Attach.facing.Face().Axis() != face.Axis()
	}
	if _, ok := classSupport(b); ok {
		return false
	}
	return b.Model().FaceSolid(pos, face, tx)
}

// allHangingSigns ...
func allHangingSigns() []world.Block {
	woods := WoodTypes()
	signs := make([]world.Block, 0, len(woods)*24)
	for _, w := range woods {
		for _, d := range cube.Directions() {
			signs = append(signs, HangingSign{Wood: w, Attach: WallHangingAttachment(d)})
			signs = append(signs, HangingSign{Wood: w, Attach: CeilingHangingAttachment(d)})
		}
		for o := cube.Orientation(0); o <= 15; o++ {
			signs = append(signs, HangingSign{Wood: w, Attach: AttachedCeilingHangingAttachment(o)})
		}
	}
	return signs
}
