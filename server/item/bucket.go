package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// BucketContent is the content of a bucket.
type BucketContent struct {
	liquid world.Liquid
	milk   bool
}

// LiquidBucketContent returns a new BucketContent with the liquid passed in.
func LiquidBucketContent(l world.Liquid) BucketContent {
	return BucketContent{liquid: l}
}

// MilkBucketContent returns a new BucketContent with the milk flag set.
func MilkBucketContent() BucketContent {
	return BucketContent{milk: true}
}

// Liquid returns the world.Liquid that a Bucket with this BucketContent places.
// If this BucketContent does not place a liquid block, false is returned.
func (b BucketContent) Liquid() (world.Liquid, bool) {
	return b.liquid, b.liquid != nil
}

// String converts the BucketContent to a string.
func (b BucketContent) String() string {
	if b.milk {
		return "milk"
	} else if b.liquid != nil {
		return b.liquid.LiquidType()
	}
	return ""
}

// LiquidType returns the type of liquid the bucket contains.
func (b BucketContent) LiquidType() string {
	if b.liquid != nil {
		return b.liquid.LiquidType()
	}
	return "milk"
}

// Bucket is a tool used to carry water, lava and fish.
type Bucket struct {
	// Content is the content that the bucket has. By default, this value resolves to an empty bucket.
	Content BucketContent
}

// MaxCount returns 16.
func (b Bucket) MaxCount() int {
	if b.Empty() {
		return 16
	}
	return 1
}

// AlwaysConsumable ...
func (b Bucket) AlwaysConsumable() bool {
	return b.Content.milk
}

// CanConsume ...
func (b Bucket) CanConsume() bool {
	return b.Content.milk
}

// ConsumeDuration ...
func (b Bucket) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (b Bucket) Consume(_ *world.World, c Consumer) Stack {
	for _, effect := range c.Effects() {
		c.RemoveEffect(effect.Type())
	}
	return NewStack(Bucket{}, 1)
}

// Empty returns true if the bucket is empty.
func (b Bucket) Empty() bool {
	return b.Content.liquid == nil && !b.Content.milk
}

// FuelInfo ...
func (b Bucket) FuelInfo() FuelInfo {
	if liq := b.Content.liquid; liq != nil && liq.LiquidType() == "lava" {
		return newFuelInfo(time.Second * 1000).WithResidue(NewStack(Bucket{}, 1))
	}
	return FuelInfo{}
}

// UseOnBlock handles the bucket filling and emptying logic.
func (b Bucket) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b.Content.milk {
		return false
	}
	if b.Empty() {
		return b.fillFrom(pos, w, ctx)
	}
	liq := b.Content.liquid.WithDepth(8, false)
	if bl := w.Block(pos); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos, liq)
	} else if bl := w.Block(pos.Side(face)); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		w.SetLiquid(pos.Side(face), liq)
	} else {
		return false
	}

	w.PlaySound(pos.Vec3Centre(), sound.BucketEmpty{Liquid: b.Content.liquid})
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

	ctx.NewItem = NewStack(Bucket{Content: LiquidBucketContent(liquid)}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b Bucket) EncodeItem() (name string, meta int16) {
	if !b.Empty() {
		return "minecraft:" + b.Content.String() + "_bucket", 0
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
