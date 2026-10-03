package block

import (
	"image/color"
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Cauldron is a hollow vessel that stores water, lava, powder snow or potions.
// TODO: Support comparator output and pointed dripstone filling when those blocks are implemented.
type Cauldron struct {
	transparent

	// Level is the fill level, from 0 (empty) to 6 (full). A bottle holds two levels.
	Level int
	// Content is the material inside the cauldron. Its zero value is water.
	Content CauldronContent
	// Potion is nil for ordinary water, or an item.Potion, item.SplashPotion or
	// item.LingeringPotion describing the stored potion and its bottle type.
	Potion world.Item
	// Colour is the colour of dyed water. Its zero value means the water is undyed.
	Colour color.RGBA
}

// Model ...
func (Cauldron) Model() world.BlockModel { return model.Cauldron{} }

// LightEmissionLevel ...
func (c Cauldron) LightEmissionLevel() uint8 {
	if c.Content == LavaCauldronContent() && c.Level > 0 {
		return 15
	}
	return 0
}

// liquidPosition is the position used by Bedrock's cauldron level events.
func (c Cauldron) liquidPosition(pos cube.Pos) mgl64.Vec3 {
	return pos.Vec3().Add(mgl64.Vec3{0.5, 0.375 + float64(c.Level)*0.125, 0.5})
}

// liquidColour is the particle colour sent with a cauldron interaction.
func (c Cauldron) liquidColour() color.RGBA {
	if c.Colour != (color.RGBA{}) {
		return c.Colour
	}
	if p, _, ok := cauldronPotion(c.Potion); ok {
		colour, _ := effect.ResultingColour(p.Effects())
		return colour
	}
	return color.RGBA{R: 52, G: 81, B: 89, A: 191}
}

// BreakInfo ...
func (Cauldron) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(Cauldron{}))
}

// Pick ...
func (Cauldron) Pick() item.Stack { return item.NewStack(Cauldron{}, 1) }

// UseOnBlock ...
func (Cauldron) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	c := Cauldron{}
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}
	place(tx, pos, c, u, ctx)
	return placed(ctx)
}

// Activate ...
func (c Cauldron) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	switch it := held.Item().(type) {
	case item.Bucket:
		return c.useBucket(pos, tx, it, ctx)
	case item.GlassBottle:
		res, filled, ok := c.FillBottle()
		if !ok {
			return false
		}
		tx.SetBlock(pos, res, nil)
		ctx.SubtractFromCount(1)
		ctx.NewItem = filled
		if c.Potion == nil {
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronTakeWater{Colour: c.liquidColour()})
		} else {
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronTakePotion{Colour: c.liquidColour()})
		}
		return true
	case item.Potion, item.SplashPotion, item.LingeringPotion:
		return c.usePotion(pos, tx, it, ctx)
	case item.Dye:
		if !c.water() {
			return false
		}
		colour := it.Colour.RGBA()
		if c.Colour != (color.RGBA{}) {
			colour.R = uint8((uint16(c.Colour.R) + uint16(colour.R)) / 2)
			colour.G = uint8((uint16(c.Colour.G) + uint16(colour.G)) / 2)
			colour.B = uint8((uint16(c.Colour.B) + uint16(colour.B)) / 2)
		}
		c.Colour = colour
		ctx.SubtractFromCount(1)
		tx.SetBlock(pos, c, nil)
		tx.PlaySound(c.liquidPosition(pos), sound.CauldronAddDye{Colour: colour})
		return true
	case item.Arrow:
		p, _, ok := cauldronPotion(c.Potion)
		if !ok || p.Uint8() <= 4 || it.Tip.Uint8() > 4 || c.Level == 0 {
			return false
		}
		n := min(held.Count(), max(3, c.Level)*16-32)
		c.Level -= (n + 15) / 16
		if c.Level < 3 {
			c = Cauldron{}
		}
		ctx.SubtractFromCount(n)
		ctx.NewItem = held.WithItem(item.Arrow{Tip: p}).Grow(n - held.Count())
		tx.SetBlock(pos, c, nil)
		return true
	case Banner:
		if !c.water() || c.Colour != (color.RGBA{}) {
			return false
		}
		if len(it.Patterns) == 0 {
			return true
		}
		it.Patterns = it.Patterns[:len(it.Patterns)-1]
		ctx.SubtractFromCount(1)
		ctx.NewItem = held.WithItem(it).Grow(1 - held.Count())
		c = c.drain(1)
		tx.SetBlock(pos, c, nil)
		tx.PlaySound(c.liquidPosition(pos), sound.CauldronCleanBanner{Colour: c.liquidColour()})
		return true
	}
	return c.useArmour(pos, tx, held, ctx)
}

// useBucket fills or empties a cauldron using a bucket.
func (c Cauldron) useBucket(pos cube.Pos, tx *world.Tx, b item.Bucket, ctx *item.UseContext) bool {
	if b.Empty() {
		if c.Level != 6 || c.Potion != nil {
			return false
		}
		var filled world.Item
		switch c.Content {
		case WaterCauldronContent():
			filled = item.Bucket{Content: item.LiquidBucketContent(Water{Depth: 8})}
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronTakeWater{Colour: c.liquidColour()})
		case LavaCauldronContent():
			filled = item.Bucket{Content: item.LiquidBucketContent(Lava{Depth: 8})}
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronTakeLava{})
		case PowderSnowCauldronContent():
			// Snow bucket item support is registered independently of cauldrons.
			bucket, ok := world.ItemByName("minecraft:powder_snow_bucket", 0)
			if !ok {
				return false
			}
			filled = bucket
			tx.PlaySound(pos.Vec3Centre(), sound.CauldronTakePowderSnow{})
		}
		c = Cauldron{}
		ctx.NewItem = item.NewStack(filled, 1)
	} else {
		var content CauldronContent
		switch b.Content.String() {
		case "water":
			content = WaterCauldronContent()
		case "lava":
			content = LavaCauldronContent()
		case "powder_snow":
			content = PowderSnowCauldronContent()
		default:
			return false
		}
		if c.Level > 0 && (c.Content != content || c.Potion != nil) {
			c = c.explode(pos, tx)
		} else {
			c = Cauldron{Level: 6, Content: content}
			switch content {
			case WaterCauldronContent():
				tx.PlaySound(c.liquidPosition(pos), sound.CauldronFillWater{Colour: c.liquidColour()})
			case LavaCauldronContent():
				tx.PlaySound(c.liquidPosition(pos), sound.CauldronFillLava{})
			case PowderSnowCauldronContent():
				tx.PlaySound(pos.Vec3Centre(), sound.CauldronFillPowderSnow{})
			}
		}
		ctx.NewItem = item.NewStack(item.Bucket{}, 1)
	}
	tx.SetBlock(pos, c, nil)
	ctx.SubtractFromCount(1)
	return true
}

// usePotion pours one bottle into the cauldron, destroying incompatible contents.
func (c Cauldron) usePotion(pos cube.Pos, tx *world.Tx, it world.Item, ctx *item.UseContext) bool {
	p, _, _ := cauldronPotion(it)
	stored, _, _ := cauldronPotion(c.Potion)
	water := p == potion.Water()
	if c.Level > 0 && (c.Content != WaterCauldronContent() || (!water && c.Potion == nil) || (water && c.Potion != nil) || (c.Potion != nil && p != stored)) {
		c = c.explode(pos, tx)
	} else {
		if c.Level == 6 && !water {
			return true
		}
		// A bottle leaves either undyed water or the poured potion behind.
		c = Cauldron{Level: min(6, c.Level+2)}
		if water {
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronFillWater{Colour: c.liquidColour()})
		} else {
			c.Potion = it
			tx.PlaySound(c.liquidPosition(pos), sound.CauldronFillPotion{Colour: c.liquidColour()})
		}
	}
	tx.SetBlock(pos, c, nil)
	ctx.SubtractFromCount(1)
	ctx.NewItem = item.NewStack(item.GlassBottle{}, 1)
	return true
}

// explode empties incompatible contents and plays the native mixing effect.
func (c Cauldron) explode(pos cube.Pos, tx *world.Tx) Cauldron {
	// Bedrock clears the potion before choosing the particle colour, but retains
	// any dye for the effect. The resulting empty cauldron stores neither.
	c = Cauldron{Colour: c.Colour}
	tx.PlaySound(c.liquidPosition(pos), sound.CauldronExplode{Colour: c.liquidColour()})
	return Cauldron{}
}

// FillBottle ...
func (c Cauldron) FillBottle() (world.Block, item.Stack, bool) {
	if c.Level < 2 || c.Content != WaterCauldronContent() {
		return nil, item.Stack{}, false
	}
	it := c.Potion
	if it == nil {
		it = item.Potion{Type: potion.Water()}
	}
	return c.drain(2), item.NewStack(it, 1), true
}

// useArmour dyes or cleans leather armour without discarding its stack metadata.
func (c Cauldron) useArmour(pos cube.Pos, tx *world.Tx, held item.Stack, ctx *item.UseContext) bool {
	if !c.water() {
		return false
	}
	var tier item.ArmourTier
	var withTier func(item.ArmourTier) world.Item
	switch it := held.Item().(type) {
	case item.Helmet:
		tier = it.Tier
		withTier = func(t item.ArmourTier) world.Item { it.Tier = t; return it }
	case item.Chestplate:
		tier = it.Tier
		withTier = func(t item.ArmourTier) world.Item { it.Tier = t; return it }
	case item.Leggings:
		tier = it.Tier
		withTier = func(t item.ArmourTier) world.Item { it.Tier = t; return it }
	case item.Boots:
		tier = it.Tier
		withTier = func(t item.ArmourTier) world.Item { it.Tier = t; return it }
	default:
		return false
	}
	leather, ok := tier.(item.ArmourTierLeather)
	if !ok {
		return false
	}
	if c.Colour == (color.RGBA{}) && leather.Colour == (color.RGBA{}) {
		return true
	}
	leather.Colour = c.Colour
	ctx.NewItem = held.WithItem(withTier(leather))
	ctx.NewItemReplaceHeld = true
	colour := c.Colour
	c = c.drain(1)
	if colour == (color.RGBA{}) {
		tx.PlaySound(c.liquidPosition(pos), sound.CauldronCleanArmour{Colour: c.liquidColour()})
	} else {
		tx.PlaySound(c.liquidPosition(pos), sound.CauldronDyeArmour{Colour: colour})
	}
	tx.SetBlock(pos, c, nil)
	return true
}

// water reports whether the cauldron contains water rather than a potion.
func (c Cauldron) water() bool {
	return c.Level > 0 && c.Content == WaterCauldronContent() && c.Potion == nil
}

// drain removes content and clears block entity data when the cauldron becomes empty.
func (c Cauldron) drain(n int) Cauldron {
	c.Level -= n
	if c.Level <= 0 {
		return Cauldron{}
	}
	return c
}

// EntityInside ...
func (c Cauldron) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if c.Level == 0 {
		return
	}
	if cube.PosFromVec3(e.Position()) != pos {
		return
	}
	if c.Content == LavaCauldronContent() {
		if f, ok := e.(flammableEntity); ok {
			if living, ok := e.(livingEntity); ok {
				living.Hurt(4, LavaDamageSource{})
			}
			f.SetOnFire(8 * time.Second)
		}
	} else if c.water() || c.Content == PowderSnowCauldronContent() {
		if f, ok := e.(flammableEntity); ok && f.OnFireDuration() > 0 {
			f.Extinguish()
			if fall, ok := e.(fallDistanceEntity); ok {
				fall.ResetFallDistance()
			}
			c.Content = WaterCauldronContent()
			tx.SetBlock(pos, c.drain(1), nil)
		}
	}
}

// RandomTick fills exposed cauldrons during rain or snowfall.
func (c Cauldron) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if c.Level == 6 || c.Potion != nil || c.Colour != (color.RGBA{}) || c.Content == LavaCauldronContent() || r.IntN(20) != 0 {
		return
	}
	above := pos.Side(cube.FaceUp)
	switch {
	case tx.RainingAt(above) && (c.Level == 0 || c.Content == WaterCauldronContent()):
		c.Content = WaterCauldronContent()
	case tx.SnowingAt(above) && (c.Level == 0 || c.Content == PowderSnowCauldronContent()):
		c.Content = PowderSnowCauldronContent()
	default:
		return
	}
	c.Level++
	tx.SetBlock(pos, c, nil)
}

// EncodeItem ...
func (Cauldron) EncodeItem() (string, int16) { return "minecraft:cauldron", 0 }

// EncodeBlock ...
func (c Cauldron) EncodeBlock() (string, map[string]any) {
	return "minecraft:cauldron", map[string]any{"fill_level": int32(c.Level), "cauldron_liquid": c.Content.String()}
}

// EncodeNBT ...
func (c Cauldron) EncodeNBT() map[string]any {
	id, kind := int16(-1), int16(0)
	if p, k, ok := cauldronPotion(c.Potion); ok {
		id, kind = int16(p.Uint8()), k
	}
	m := map[string]any{"id": "Cauldron", "PotionId": id, "PotionType": kind}
	if c.Colour != (color.RGBA{}) {
		m["CustomColor"] = nbtconv.Int32FromRGBA(c.Colour)
	}
	return m
}

// DecodeNBT ...
func (c Cauldron) DecodeNBT(m map[string]any) any {
	c.Potion, c.Colour = nil, color.RGBA{}
	if id, ok := m["PotionId"].(int16); ok && id >= 0 {
		p := potion.From(int32(id))
		kind := nbtconv.Int16(m, "PotionType")
		if nbtconv.Uint8(m, "IsSplash") != 0 {
			kind = 1
		}
		switch kind {
		case 0:
			c.Potion = item.Potion{Type: p}
		case 1:
			c.Potion = item.SplashPotion{Type: p}
		case 2:
			c.Potion = item.LingeringPotion{Type: p}
		}
	}
	if _, ok := m["CustomColor"]; ok {
		c.Colour = nbtconv.RGBAFromInt32(nbtconv.Int32(m, "CustomColor"))
	}
	return c
}

// cauldronPotion extracts a potion's effect and bottle type.
func cauldronPotion(it world.Item) (potion.Potion, int16, bool) {
	switch p := it.(type) {
	case item.Potion:
		return p.Type, 0, true
	case item.SplashPotion:
		return p.Type, 1, true
	case item.LingeringPotion:
		return p.Type, 2, true
	}
	return potion.Potion{}, 0, false
}

// allCauldrons ...
func allCauldrons() (all []world.Block) {
	for _, content := range []CauldronContent{WaterCauldronContent(), LavaCauldronContent(), PowderSnowCauldronContent()} {
		for level := 0; level <= 6; level++ {
			all = append(all, Cauldron{Level: level, Content: content})
		}
	}
	return
}
