package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Bucket is a tool used to carry water, lava, milk and fish.
type Bucket struct {
	// Content is the content that the bucket has. By default, this value resolves to an empty bucket.
	Content world.Liquid
}

// MaxCount returns 16.
func (b Bucket) MaxCount() int {
	if b.Empty() {
		return 16
	}
	return 1
}

// Empty returns true if the bucket is empty.
func (b Bucket) Empty() bool {
	return b.Content == nil
}

// UseOnBlock handles the bucket filling and emptying logic.
func (b Bucket) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b.Empty() {
		return b.fillFrom(pos, w, ctx)
	}
	liq := b.Content.WithDepth(8, false)
	if bl := w.Block(pos); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos, liq)
	} else if bl := w.Block(pos.Side(face)); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos.Side(face), liq)
	} else {
		return false
	}

	w.PlaySound(pos.Vec3Centre(), sound.BucketEmpty{Liquid: b.Content})
	ctx.NewItem = NewStack(Bucket{}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// fillFrom fills a bucket from the liquid at the position passed in the world. If there is no liquid or if
// the liquid is no source, fillFrom returns false.
func (b Bucket) fillFrom(pos cube.Pos, w *world.World, ctx *UseContext) bool {
	liquid, ok := w.Liquid(pos)
	if !ok {
		return false
	}
	if liquid.LiquidDepth() != 8 || liquid.LiquidFalling() {
		// Only allow picking up liquid source blocks.
		return false
	}
	w.SetLiquid(pos, nil)
	w.PlaySound(pos.Vec3Centre(), sound.BucketFill{Liquid: liquid})

	ctx.NewItem = NewStack(Bucket{Content: liquid}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b Bucket) EncodeItem() (name string, meta int16) {
	if !b.Empty() {
		return "minecraft:" + b.Content.LiquidType() + "_bucket", 0
	}
	return "minecraft:bucket", 0
}

type replaceable interface {
	ReplaceableBy(b world.Block) bool
}

func replaceableWith(b world.Block, with world.Block) bool {
	if r, ok := b.(replaceable); ok {
		return r.ReplaceableBy(with)
	}
	return false
}

func canDisplace(b world.Block, liq world.Liquid) bool {
	if d, ok := b.(world.LiquidDisplacer); ok {
		return d.CanDisplace(liq)
	}
	return false
}
