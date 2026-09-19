package entity

import (
	"math"
	"math/bits"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// cushionSupportProbe is how far below a cushion its supporting block is looked for.
const cushionSupportProbe = 1.0 / 64

// cushionCanSurvive reports if a cushion is supported and not fully covered by suffocating blocks.
func cushionCanSurvive(pos mgl64.Vec3, tx *world.Tx) bool {
	if !cushionSupported(pos, tx) {
		return false
	}
	box := CushionType.BBox(nil).Translate(pos)
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())
	return !cushionCovers(tx.Block(low)) || !cushionCovers(tx.Block(high))
}

// cushionSupported reports if the non-liquid block below the cushion reaches a cushion sized box centred at the probe.
func cushionSupported(pos mgl64.Vec3, tx *world.Tx) bool {
	probe := pos.Sub(mgl64.Vec3{0, cushionSupportProbe, 0})
	blockPos := cube.PosFromVec3(probe)
	b := tx.Block(blockPos)
	if _, ok := b.(world.Liquid); ok {
		return false
	}
	half := mgl64.Vec3{cushionWidth / 2, cushionHeight / 2, cushionWidth / 2}
	box := cube.Box(probe[0]-half[0], probe[1]-half[1], probe[2]-half[2], probe[0]+half[0], probe[1]+half[1], probe[2]+half[2])
	for _, shape := range cushionSupportShape(b, blockPos, tx) {
		if shape.Translate(blockPos.Vec3()).IntersectsWith(box) {
			return true
		}
	}
	return false
}

// cushionSupportShape returns the collision box of a block, or the vanilla shape of blocks without one, as measured
// on a vanilla server.
func cushionSupportShape(b world.Block, pos cube.Pos, tx *world.Tx) []cube.BBox {
	switch b := b.(type) {
	case block.LilyPad:
		return cushionShape(0, 0.09375)
	case block.Flower:
		return cushionShape(0, 0.6)
	case block.Torch:
		return cushionStandingTorchShape(b.Facing)
	case block.CopperTorch:
		return cushionStandingTorchShape(b.Facing)
	case block.RedstoneTorch:
		return cushionStandingTorchShape(b.Facing)
	case block.Lever:
		switch b.Facing {
		case cube.FaceUp:
			return cushionShape(0, 0.6)
		case cube.FaceDown:
			return cushionShape(0.4, 1)
		}
		return cushionShape(0.25, 0.75)
	case block.ShortGrass, block.Fern:
		return cushionShape(0, 0.8+cushionPlantOffset(pos))
	case block.Sapling, block.BambooSapling:
		return cushionShape(0, 0.8)
	case block.Coral:
		return cushionShape(0, 0.8)
	case block.NetherSprouts:
		return cushionShape(0, 0.3125)
	case block.RedstoneWire:
		return cushionShape(0, 0.0625)
	case block.String:
		return cushionShape(0, 0.5)
	case block.SeaPickle:
		return cushionShape(0, 0.5)
	case block.EndPortal:
		return cushionShape(0, 0.25)
	case block.Sign:
		if cushionStanding(b.Attach) {
			return cushionShape(0, 1)
		}
		return cushionShape(0.28125, 0.78125)
	case block.Banner:
		if cushionStanding(b.Attach) {
			return cushionShape(0, 1)
		}
		return cushionShape(0, 0.78125)
	case block.ItemFrame:
		switch b.Facing {
		case cube.FaceDown:
			return cushionShape(0, 0.0625)
		case cube.FaceUp:
			return cushionShape(0.9375, 1)
		}
		return cushionShape(0, 1)
	case block.SporeBlossom:
		return cushionShape(0.75, 1)
	case block.WheatSeeds:
		return cushionShape(0, min(float64(b.Growth+1)/7, 1))
	case block.Carrot:
		return cushionShape(0, float64(b.Growth+1)/10)
	case block.Potato:
		return cushionShape(0, cushionRootCropHeights[b.Growth&7])
	case block.BeetrootSeeds:
		return cushionShape(0, cushionRootCropHeights[b.Growth&7])
	case block.MelonSeeds:
		return cushionShape(0, float64(b.Growth+1)/8)
	case block.PumpkinSeeds:
		return cushionShape(0, float64(b.Growth+1)/8)
	case block.NetherWart:
		return cushionShape(0, float64(b.Age+1)/4)
	case block.DoubleFlower, block.DoubleTallGrass, block.DeadBush, block.SugarCane, block.Kelp, block.Vines,
		block.Light, block.WoodFenceGate, block.Portal:
		// These reach high enough to support a cushion at any height inside of them.
		if bbs := b.Model().BBox(pos, tx); len(bbs) != 0 {
			return bbs
		}
		return cushionShape(0, 1)
	}
	// Other blocks without a collision box, such as cobwebs and wall torches, do not support cushions.
	return b.Model().BBox(pos, tx)
}

// cushionRootCropHeights holds the height of potatoes and beetroots per growth stage.
var cushionRootCropHeights = [8]float64{0.09375, 0.15625, 0.25, 0.34375, 0.4375, 0.5, 0.59375, 0.6875}

// cushionShape returns a full width shape between the heights passed.
func cushionShape(minY, maxY float64) []cube.BBox {
	return []cube.BBox{cube.Box(0, minY, 0, 1, maxY, 1)}
}

// cushionStandingTorchShape returns the shape of a torch. Only standing torches support cushions.
func cushionStandingTorchShape(facing cube.Face) []cube.BBox {
	if facing == cube.FaceDown {
		return cushionShape(0, 0.6)
	}
	return nil
}

// cushionStanding reports if the attachment is to the ground rather than a wall.
func cushionStanding(a block.Attachment) bool {
	return a.FaceUint8() == 1
}

// cushionCovers reports if a block covers a cushion inside it: Suffocating blocks, with a few vanilla exceptions.
func cushionCovers(b world.Block) bool {
	switch b.(type) {
	case block.Glowstone, block.Ice, block.TNT:
		return false
	case block.Slime, block.Barrier, block.CopperGrate, block.Cactus, block.Farmland, block.DirtPath:
		return true
	}
	if _, ok := b.Model().(model.Solid); !ok {
		return false
	}
	if d, ok := b.(block.LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		return false
	}
	if n, ok := b.(block.NonSuffocating); ok && n.PreventsSuffocation() {
		return false
	}
	return true
}

// cushionPlantOffset returns the vanilla vertical offset of short grass and ferns: One of 16 steps between -0.2 and 0,
// picked by the second number of a xoroshiro128++ generator seeded with the X and Z of the position.
func cushionPlantOffset(pos cube.Pos) float64 {
	seed := int64(int32(pos[0]*3129871)) ^ int64(pos[2])*116129781
	seed = int64(int32((seed*seed*42317861 + seed*11) >> 16))

	lo := uint64(seed) ^ 0x6a09e667f3bcc909
	s0, s1 := mixStafford13(lo), mixStafford13(lo+0x9e3779b97f4a7c15)
	if s0 == 0 && s1 == 0 {
		s0, s1 = 0x9e3779b97f4a7c15, 0x6a09e667f3bcc909
	}
	s1 ^= s0
	s0, s1 = bits.RotateLeft64(s0, 49)^s1^(s1<<21), bits.RotateLeft64(s1, 28)

	r := float64((bits.RotateLeft64(s0+s1, 17)+s0)>>40) / (1 << 24)
	return -0.2 + math.Floor(r*16)*(0.2/15)
}

// mixStafford13 mixes a xoroshiro128++ seed.
func mixStafford13(x uint64) uint64 {
	x = (x ^ x>>30) * 0xbf58476d1ce4e5b9
	x = (x ^ x>>27) * 0x94d049bb133111eb
	return x ^ x>>31
}
