package block

import (
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
//
// Placement rules (matching vanilla Bedrock):
//   - FaceDown (click underside of block): ceiling-hung. If the block above is a
//     narrow block (fence, chain, wall, etc.) OR the player is sneaking, the sign
//     uses V-shape chains (attached_bit=true, 16-step ground_sign_direction).
//     Otherwise it uses straight chains (attached_bit=false, 4-step facing_direction).
//   - FaceUp: not allowed.
//   - Side face: wall-mounted, but ONLY when the supporting block is a solid
//     full-width block. Narrow blocks (fences, chains, walls, etc.) cannot be
//     used as wall supports — the sign must be hung from below instead.
func (h HangingSign) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	clickedPos, clickedBlock := pos, tx.Block(pos)
	pos, face, used := firstReplaceable(tx, pos, face, h)
	if !used {
		return false
	}
	// When clicking the front or back face of a wall-mounted hanging sign,
	// place the new sign below it instead of on its side.
	if existing, ok := clickedBlock.(HangingSign); ok && !existing.Attach.ceiling {
		fd := existing.Attach.facing.Face()
		if face == fd || face == fd.Opposite() {
			pos = clickedPos.Side(cube.FaceDown)
			face = cube.FaceDown
		}
	}
	switch face {
	case cube.FaceDown:
		// Hanging from the underside of a block (ceiling-hung).
		supportPos := pos.Side(cube.FaceUp)
		support := tx.Block(supportPos)
		sneaking := false
		if s, ok := user.(interface{ Sneaking() bool }); ok {
			sneaking = s.Sneaking()
		}
		newDir := user.Rotation().Direction().Opposite()
		switch {
		case !sneaking && canStraightHangFrom(support, supportPos, tx, newDir):
			// Straight-chain (CeilingEdges): solid block above, or tip-to-tip with a
			// same-axis straight-chain or wall hanging sign.
			// Matches CeilingEdgesHangingSign.canBeSupportedAt.
			h.Attach = CeilingHangingAttachment(newDir)
		case canCenterHangFrom(support, sneaking):
			// V-shape (CeilingCenter): narrow/post block above, any hanging sign above,
			// or sneaking above any solid block.
			// Matches CeilingCenterHangingSign.canBeSupportedAt + vanilla sneak rule.
			h.Attach = AttachedCeilingHangingAttachment(user.Rotation().Orientation().Opposite())
		default:
			return false
		}
	case cube.FaceUp:
		// Cannot place a hanging sign on top of a block.
		return false
	default:
		// Wall-mounted on the side of a block.
		// Narrow blocks (fences, chains, walls, bars, panes) do not provide a flat
		// wall surface — the sign must be hung from their underside instead.
		support := tx.Block(pos.Side(face.Opposite()))
		if isNarrowHangingBlock(support) {
			// Wall-mounted hanging signs may chain side-to-side.
			if hs, ok := support.(HangingSign); !ok || hs.Attach.ceiling {
				return false
			}
		}
		// The sign panel is perpendicular to the wall attachment axis.
		// Facing = RotateRight from the direction toward the wall.
		wallFacing := face.Direction().RotateRight()
		// Orient the text toward the player when possible, so the front is readable.
		if wallFacing == user.Rotation().Direction() {
			wallFacing = wallFacing.Opposite()
		}
		h.Attach = WallHangingAttachment(wallFacing)
	}
	place(tx, pos, h, user, ctx)
	if editor, ok := user.(SignEditor); ok {
		editor.OpenSign(pos, true)
	}
	return placed(ctx)
}

// Activate opens the sign editor on right-click if the sign is not waxed.
// Returns false when the player is holding a hanging sign item so that
// placement (chaining signs) is not blocked by the editor.
func (h HangingSign) Activate(pos cube.Pos, face cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if carrier, ok := u.(interface {
		HeldItems() (mainHand, offHand item.Stack)
	}); ok {
		held, _ := carrier.HeldItems()
		if _, isHangingSign := held.Item().(HangingSign); isHangingSign {
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
	if !h.Attach.ceiling {
		// Wall-mounted signs may chain side-to-side, but the chain must eventually
		// connect to a real wall support. A segment with only signs is invalid.
		fd := h.Attach.facing.Face()
		perp1, perp2, ok := wallSupportFaces(fd)
		if !ok {
			return
		}
		if !hasWallAnchor(tx, pos, perp1, perp2, map[cube.Pos]struct{}{}) {
			breakBlock(h, pos, tx)
		}
		return
	}
	supportPos := pos.Side(cube.FaceUp)
	support := tx.Block(supportPos)
	if h.Attach.attached {
		// V-shape (attached_bit=true): only break if the block above is completely gone.
		if _, ok := support.(Air); ok {
			breakBlock(h, pos, tx)
		}
	} else {
		// Straight-chain (attached_bit=false): break if support can no longer hold it.
		// Any HangingSign above (regardless of its facing direction) is a valid support —
		// axis checks are placement-time only, not break-condition checks.
		if _, ok := support.(Air); ok {
			breakBlock(h, pos, tx)
			return
		}
		switch support.(type) {
		case WoodFence, NetherBrickFence, Wall,
			IronChain, CopperChain,
			IronBars, CopperBars,
			GlassPane:
			breakBlock(h, pos, tx)
			return
		}
		if _, ok := support.(HangingSign); ok {
			return // any hanging sign above is a valid support
		}
		if !support.Model().FaceSolid(supportPos, cube.FaceDown, tx) {
			breakBlock(h, pos, tx)
		}
	}
}

// wallSupportFaces returns the two faces that may support a wall-mounted hanging
// sign for a given panel-facing direction.
func wallSupportFaces(fd cube.Face) (cube.Face, cube.Face, bool) {
	switch fd {
	case cube.FaceNorth, cube.FaceSouth:
		return cube.FaceEast, cube.FaceWest, true
	case cube.FaceEast, cube.FaceWest:
		return cube.FaceNorth, cube.FaceSouth, true
	default:
		return 0, 0, false
	}
}

// hasWallAnchor reports whether the wall-mounted hanging sign at pos has a path
// through side-connected wall-mounted signs to at least one real wall support.
func hasWallAnchor(tx *world.Tx, pos cube.Pos, perp1, perp2 cube.Face, visited map[cube.Pos]struct{}) bool {
	if _, ok := visited[pos]; ok {
		return false
	}
	visited[pos] = struct{}{}

	for _, side := range []cube.Face{perp1, perp2} {
		nPos := pos.Side(side)
		nb := tx.Block(nPos)

		if _, ok := nb.(Air); ok {
			continue
		}
		// Any non-narrow block is a real wall anchor.
		if !isNarrowHangingBlock(nb) {
			return true
		}
		// Continue through wall-mounted hanging signs on the same support axis.
		if hs, ok := nb.(HangingSign); ok && !hs.Attach.ceiling {
			np1, np2, valid := wallSupportFaces(hs.Attach.facing.Face())
			if !valid {
				continue
			}
			if (np1 == perp1 && np2 == perp2) || (np1 == perp2 && np2 == perp1) {
				if hasWallAnchor(tx, nPos, perp1, perp2, visited) {
					return true
				}
			}
		}
	}
	return false
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
	s := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}
	s = s.DecodeNBT(data).(Sign)
	h.Front, h.Back, h.Waxed = s.Front, s.Back, s.Waxed
	return h
}

// EncodeNBT ...
func (h HangingSign) EncodeNBT() map[string]any {
	nbt := Sign{Front: h.Front, Back: h.Back, Waxed: h.Waxed}.EncodeNBT()
	nbt["id"] = "HangingSign"
	return nbt
}

// EditingFrontSide reports whether the user is editing the front side of the sign.
func (h HangingSign) EditingFrontSide(pos cube.Pos, userPos mgl64.Vec3) bool {
	return userPos.Sub(pos.Vec3Centre()).Dot(h.Attach.Rotation().Vec3()) > 0
}

// canStraightHangFrom reports whether a straight-chain (AttachedBit=false) hanging sign
// can be placed below block b. supportPos is b's position, tx is the world, and newDir is
// the facing direction the new sign would use.
//
//   - SupportType::FULL → block whose bottom face is solid (full blocks, bottom slabs, etc.).
//   - WallHangingSign or CeilingEdgesHangingSign with the same facing axis → tip-to-tip.
func canStraightHangFrom(b world.Block, supportPos cube.Pos, tx *world.Tx, newDir cube.Direction) bool {
	if _, ok := b.(Air); ok {
		return false
	}
	switch b.(type) {
	case WoodFence, NetherBrickFence, Wall,
		IronChain, CopperChain,
		IronBars, CopperBars,
		GlassPane:
		return false
	}
	if hs, ok := b.(HangingSign); ok {
		if hs.Attach.ceiling && hs.Attach.attached {
			return false
		}
		return hs.Attach.facing == newDir || hs.Attach.facing == newDir.Opposite()
	}
	return b.Model().FaceSolid(supportPos, cube.FaceDown, tx)
}

// canCenterHangFrom reports whether a V-shape (AttachedBit=true) hanging sign can be
// placed below block b.
//
//   - hasCenterSupport() → fence, wall, chain, iron bars, glass pane.
//   - hasTypeTag(HANGING_SIGN) → any hanging sign variant.
//
// Sneaking also allows placement below any solid (non-air) block (vanilla Bedrock
// forced-V-shape behaviour).
func canCenterHangFrom(b world.Block, sneaking bool) bool {
	if _, ok := b.(Air); ok {
		return false
	}
	// isNarrowHangingBlock covers fences, walls, chains, bars, panes, and hanging signs.
	return isNarrowHangingBlock(b) || sneaking
}

// canWallHangFrom reports whether block b can support a wall-mounted hanging sign.
// Any non-narrow, non-air block is a valid wall support.
func canWallHangFrom(b world.Block) bool {
	if _, ok := b.(Air); ok {
		return false
	}
	return !isNarrowHangingBlock(b)
}

// isNarrowHangingBlock reports whether b is a narrow/thin block. Narrow blocks:
//   - Trigger V-shape (attached_bit=true) chains when a hanging sign is hung from below.
//   - Cannot be used as a wall surface for wall-mounted hanging signs.
//
// This matches the BlockThin / BlockIronChain / BlockHangingSign check.
func isNarrowHangingBlock(b world.Block) bool {
	switch b.(type) {
	case WoodFence, NetherBrickFence, Wall,
		IronChain, CopperChain,
		IronBars, CopperBars,
		GlassPane,
		HangingSign:
		return true
	}
	return false
}

// allHangingSigns returns a list of all hanging sign block states.
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
