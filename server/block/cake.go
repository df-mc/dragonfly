package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Cake is an edible block.
type Cake struct {
	transparent
	sourceWaterDisplacer

	// Colour is the colour of the candle.
	Colour OptionalColour
	// Bites is the amount of bites taken out of the cake.
	Bites int
	// Candle is true if the cake has a candle on top.
	Candle bool
	// Lit is whether the candle is lit.
	Lit bool
}

// LightEmissionLevel ...
func (c Cake) LightEmissionLevel() uint8 {
	if c.Candle && c.Lit {
		return 3
	}
	return 0
}

// SideClosed ...
func (c Cake) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// UseOnBlock ...
func (c Cake) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}

	if _, air := tx.Block(pos.Side(cube.FaceDown)).(Air); air {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c Cake) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, air := tx.Block(pos.Side(cube.FaceDown)).(Air); air {
		breakBlock(c, pos, tx)
	}
}

// Activate ...
func (c Cake) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if c.Bites == 0 && !c.Candle {
		if candle, ok := held.Item().(Candle); ok {
			c.Candle = true
			c.Colour = candle.Colour
			tx.SetBlock(pos, c, nil)
			tx.PlaySound(pos.Vec3Centre(), sound.ItemUseOn{Block: c})
			ctx.SubtractFromCount(1)
			return true
		}
	}

	if _, ok := held.Item().(item.FlintAndSteel); ok {
		return false
	}

	if c.Candle {
		if c.Lit {
			c.Lit = false
			tx.SetBlock(pos, c, nil)
			return true
		}
	}

	if i, ok := u.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(2, 0.4)
		tx.PlaySound(u.Position().Add(mgl64.Vec3{0, 1.5}), sound.Burp{})

		if c.Candle {
			candle := Candle{Colour: c.Colour}
			dropItem(tx, item.NewStack(candle, 1), pos.Vec3Centre())

			c.Candle = false
			c.Colour = c.Colour.Empty()
			c.Lit = false
		}

		c.Bites++
		if c.Bites > 6 {
			tx.SetBlock(pos, nil, nil)
			return true
		}
		tx.SetBlock(pos, c, nil)
		return true
	}
	return false
}

// Ignite ...
func (c Cake) Ignite(pos cube.Pos, tx *world.Tx, _ world.Entity) bool {
	if !c.Candle || c.Lit {
		return false
	}

	c.Lit = true
	tx.SetBlock(pos, c, nil)
	tx.PlaySound(pos.Vec3(), sound.Ignite{})
	return true
}

// BreakInfo ...
func (c Cake) BreakInfo() BreakInfo {
	drops := simpleDrops()
	if c.Candle {
		drops = oneOf(Candle{Colour: c.Colour})
	}
	return newBreakInfo(0.5, neverHarvestable, nothingEffective, drops)
}

// EncodeItem ...
func (c Cake) EncodeItem() (name string, meta int16) {
	if c.Candle {
		color, ok := c.Colour.Colour()
		if ok {
			return "minecraft:" + color.String() + "_candle_cake", 0
		}
		return "minecraft:candle_cake", 0
	}
	return "minecraft:cake", 0
}

// EncodeBlock ...
func (c Cake) EncodeBlock() (name string, properties map[string]any) {
	if c.Candle {
		color, ok := c.Colour.Colour()
		if ok {
			name = "minecraft:" + color.String() + "_candle_cake"
		} else {
			name = "minecraft:candle_cake"
		}
		return name, map[string]any{"lit": c.Lit}
	}
	return "minecraft:cake", map[string]any{"bite_counter": int32(c.Bites)}
}

// Model ...
func (c Cake) Model() world.BlockModel {
	return model.Cake{Bites: c.Bites}
}

// allCake ...
func allCake() (cake []world.Block) {
	for bites := 0; bites < 7; bites++ {
		cake = append(cake, Cake{Bites: bites})
	}
	for _, c := range OptionalColours() {
		cake = append(cake, Cake{Colour: c, Candle: true})
		cake = append(cake, Cake{Colour: c, Candle: true, Lit: true})
	}
	return
}
