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
	liquid     world.Liquid
	milk       bool
	powderSnow bool
}

// LiquidBucketContent returns a new BucketContent with the liquid passed in.
func LiquidBucketContent(l world.Liquid) BucketContent {
	return BucketContent{liquid: l}
}

// MilkBucketContent returns a new BucketContent with the milk flag set.
func MilkBucketContent() BucketContent {
	return BucketContent{milk: true}
}

// PowderSnowBucketContent returns a bucket containing powder snow.
func PowderSnowBucketContent() BucketContent {
	return BucketContent{powderSnow: true}
}

// Liquid returns the world.Liquid that a Bucket with this BucketContent places.
// If this BucketContent does not place a liquid block, false is returned.
func (b BucketContent) Liquid() (world.Liquid, bool) {
	return b.liquid, b.liquid != nil
}

// String converts the BucketContent to a string.
func (b BucketContent) String() string {
	switch {
	case b.powderSnow:
		return "powder_snow"
	case b.milk:
		return "milk"
	case b.liquid != nil:
		return b.liquid.LiquidType()
	}
	return ""
}

// LiquidType returns the type of liquid the bucket contains.
func (b BucketContent) LiquidType() string {
	if b.powderSnow {
		return "powder_snow"
	}
	if b.liquid != nil {
		return b.liquid.LiquidType()
	}
	return "milk"
}

// Bucket is a tool used to carry liquids, milk and powder snow.
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
func (b Bucket) Consume(_ *world.Tx, c Consumer) Stack {
	for _, effect := range c.Effects() {
		c.RemoveEffect(effect.Type())
	}
	return NewStack(Bucket{}, 1)
}

// Empty returns true if the bucket is empty.
func (b Bucket) Empty() bool {
	return b.Content.liquid == nil && !b.Content.milk && !b.Content.powderSnow
}

// FuelInfo ...
func (b Bucket) FuelInfo() FuelInfo {
	if liq := b.Content.liquid; liq != nil && liq.LiquidType() == "lava" {
		return newFuelInfo(time.Second * 1000).WithResidue(NewStack(Bucket{}, 1))
	}
	return FuelInfo{}
}

// UseOnBlock handles the bucket filling and emptying logic.
func (b Bucket) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if b.Content.milk {
		return false
	}
	if b.Empty() {
		return b.fillFrom(pos, tx, ctx)
	}
	if b.Content.powderSnow {
		return b.placePowderSnow(pos, face, tx, user, ctx)
	}
	liq := b.Content.liquid.WithDepth(8, false)
	if bl := tx.Block(pos); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		tx.SetLiquid(pos, liq)
	} else if bl := tx.Block(pos.Side(face)); canDisplace(bl, liq) || replaceableWith(bl, liq) {
		tx.SetLiquid(pos.Side(face), liq)
	} else {
		return false
	}

	tx.PlaySound(pos.Vec3Centre(), sound.BucketEmpty{Liquid: b.Content.liquid})
	ctx.NewItem = NewStack(Bucket{}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// fillFrom fills a bucket from the liquid at the position passed in the world. If there is no liquid or if
// the liquid is no source, fillFrom returns false.
func (b Bucket) fillFrom(pos cube.Pos, tx *world.Tx, ctx *UseContext) bool {
	if _, snow := tx.Block(pos).(interface{ PowderSnow() }); snow {
		tx.SetBlock(pos, nil, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.BucketFill{PowderSnow: true})
		ctx.NewItem = NewStack(Bucket{Content: PowderSnowBucketContent()}, 1)
		ctx.NewItemSurvivalOnly = true
		ctx.SubtractFromCount(1)
		return true
	}
	liquid, ok := tx.Liquid(pos)
	if !ok {
		return false
	}
	if liquid.LiquidDepth() != 8 || liquid.LiquidFalling() {
		// Only allow picking up liquid source blocks.
		return false
	}
	tx.SetLiquid(pos, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.BucketFill{Liquid: liquid})

	ctx.NewItem = NewStack(Bucket{Content: LiquidBucketContent(liquid)}, 1)
	ctx.NewItemSurvivalOnly = true
	ctx.SubtractFromCount(1)
	return true
}

// placePowderSnow uses the placement hook so cancelled placements do not consume the bucket.
func (b Bucket) placePowderSnow(pos cube.Pos, face cube.Face, tx *world.Tx, user User, ctx *UseContext) bool {
	snow, ok := tx.World().BlockRegistry().BlockByName("minecraft:powder_snow", nil)
	if !ok {
		return false
	}
	if !replaceableWith(tx.Block(pos), snow) {
		pos = pos.Side(face)
	}
	if pos.OutOfBounds(tx.Range()) || !replaceableWith(tx.Block(pos), snow) {
		return false
	}
	if placer, ok := user.(interface {
		PlaceBlock(cube.Pos, world.Block, *UseContext)
	}); ok {
		before := ctx.CountSub
		placer.PlaceBlock(pos, snow, ctx)
		if ctx.CountSub == before {
			return false
		}
	} else {
		tx.SetBlock(pos, snow, nil)
		ctx.SubtractFromCount(1)
	}
	tx.PlaySound(pos.Vec3Centre(), sound.BucketEmpty{PowderSnow: true})
	ctx.NewItem = NewStack(Bucket{}, 1)
	ctx.NewItemSurvivalOnly = true
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
