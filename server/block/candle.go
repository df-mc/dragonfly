package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Candle is a dyeable block that emits light when lit with a flint and steel.
// Up to four of the same color of candle can be placed in one block space, which affects the amount of light produced.
type Candle struct {
	transparent
	sourceWaterDisplacer

	// Colour is the colour of the candle.
	Colour item.OptionalColour
	// Candles is the number of candles.
	Candles int
	// Lit is whether the candles are lit.
	Lit bool
}

// BreakInfo ...
func (c Candle) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, nothingEffective, oneOf(c))
}

// Model ...
func (c Candle) Model() world.BlockModel {
	return model.Candle{Count: c.Candles}
}

// LightEmissionLevel ...
func (c Candle) LightEmissionLevel() uint8 {
	return uint8((c.Candles + 1) * 3)
}

// SideClosed ...
func (Candle) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// UseOnBlock ...
func (c Candle) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	if existing, ok := tx.Block(pos).(Candle); ok {
		if existing.Colour != c.Colour || existing.Candles >= 3 {
			return false
		}

		existing.Candles++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c Candle) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	liquid, _ := tx.Liquid(pos)
	if _, ok := liquid.(Water); ok && c.Lit {
		c.Lit = false
		tx.SetBlock(pos, c, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})
	}
}

// Ignite ...
func (c Candle) Ignite(pos cube.Pos, tx *world.Tx, _ world.Entity) bool {
	if _, ok := tx.Liquid(pos); ok || c.Lit {
		return false
	}

	c.Lit = true
	tx.SetBlock(pos, c, nil)
	tx.PlaySound(pos.Vec3(), sound.Ignite{})
	return true
}

// EncodeItem ...
func (c Candle) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Colour.Prepend("candle"), 0
}

// EncodeBlock ...
func (c Candle) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + c.Colour.Prepend("candle"), map[string]any{"candles": int32(c.Candles), "lit": c.Lit}
}

// allCandles returns candle blocks with all possible colours.
func allCandles() []world.Block {
	b := make([]world.Block, 0)
	for i := 0; i <= 3; i++ {
		for _, c := range item.OptionalColours() {
			b = append(b, Candle{Colour: c, Candles: i})
			b = append(b, Candle{Colour: c, Candles: i, Lit: true})
		}
	}
	return b
}
