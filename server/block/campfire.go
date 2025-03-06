package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"strconv"
	"time"
)

// Campfire is a block that can be used to cook food, pacify bees, act as a spread-proof light source, smoke signal or
// damaging trap block.
type Campfire struct {
	transparent
	bass
	sourceWaterDisplacer

	// Items represents the items in the campfire that are being cooked.
	Items [4]CampfireItem
	// Facing represents the direction that the campfire is facing.
	Facing cube.Direction
	// Extinguished is true if the campfire was extinguished by a water source.
	Extinguished bool
	// Type represents the type of Campfire, currently there are Normal and Soul campfires.
	Type FireType
}

// CampfireItem holds data about the items in the campfire.
type CampfireItem struct {
	// Item is a specific item being cooked on top of the campfire.
	Item item.Stack
	// Time is the countdown of ticks until the food item is cooked (when 0).
	Time time.Duration
}

// Model ...
func (Campfire) Model() world.BlockModel {
	return model.Campfire{}
}

// SideClosed ...
func (Campfire) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (c Campfire) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(Campfire{Type: c.Type}, 1)}
		}
		switch c.Type {
		case NormalFire():
			return []item.Stack{item.NewStack(item.Charcoal{}, 2)}
		case SoulFire():
			return []item.Stack{item.NewStack(SoulSoil{}, 1)}
		}
		panic("should never happen")
	}).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		for _, v := range c.Items {
			if !v.Item.Empty() {
				dropItem(tx, v.Item, pos.Vec3Centre())
			}
		}
	})
}

// LightEmissionLevel ...
func (c Campfire) LightEmissionLevel() uint8 {
	if c.Extinguished {
		return 0
	}
	return c.Type.LightLevel()
}

// Ignite ...
func (c Campfire) Ignite(pos cube.Pos, tx *world.Tx, _ world.Entity) bool {
	tx.PlaySound(pos.Vec3(), sound.Ignite{})
	if !c.Extinguished {
		return false
	}
	if _, ok := tx.Liquid(pos); ok {
		return false
	}

	c.Extinguished = false
	tx.SetBlock(pos, c, nil)
	return true
}

// Splash ...
func (c Campfire) Splash(tx *world.Tx, pos cube.Pos) {
	if c.Extinguished {
		return
	}

	c.extinguish(pos, tx)
}

// extinguish extinguishes the campfire.
func (c Campfire) extinguish(pos cube.Pos, tx *world.Tx) {
	tx.PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})
	c.Extinguished = true

	for i := range c.Items {
		c.Items[i].Time = time.Second * 30
	}

	tx.SetBlock(pos, c, nil)
}

// Activate ...
func (c Campfire) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if held.Empty() {
		return false
	}

	if _, ok := held.Item().(item.Shovel); ok && !c.Extinguished {
		c.extinguish(pos, tx)
		ctx.DamageItem(1)
		return true
	}

	rawFood, ok := held.Item().(item.Smeltable)
	if !ok || !rawFood.SmeltInfo().Food {
		return false
	}

	if _, ok = tx.Liquid(pos); ok {
		return false
	}

	for i, it := range c.Items {
		if it.Item.Empty() {
			c.Items[i] = CampfireItem{
				Item: held.Grow(-held.Count() + 1),
				Time: time.Second * 30,
			}

			ctx.SubtractFromCount(1)

			tx.PlaySound(pos.Vec3Centre(), sound.ItemAdd{})
			tx.SetBlock(pos, c, nil)
			return true
		}
	}
	return false
}

// UseOnBlock ...
func (c Campfire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Campfire); ok {
		return false
	}
	c.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// Tick is called to cook the items within the campfire.
func (c Campfire) Tick(_ int64, pos cube.Pos, tx *world.Tx) {
	if c.Extinguished {
		// Extinguished, do nothing.
		return
	}
	if rand.Float64() <= 0.016 { // Every three or so seconds.
		tx.PlaySound(pos.Vec3Centre(), sound.CampfireCrackle{})
	}

	updated := false
	for i, it := range c.Items {
		if it.Item.Empty() {
			continue
		}

		updated = true
		if it.Time > 0 {
			c.Items[i].Time = it.Time - time.Millisecond*50
			continue
		}

		if food, ok := it.Item.Item().(item.Smeltable); ok {
			dropItem(tx, food.SmeltInfo().Product, pos.Vec3Middle())
		}
		c.Items[i].Item = item.Stack{}
	}
	if updated {
		tx.SetBlock(pos, c, nil)
	}
}

// NeighbourUpdateTick ...
func (c Campfire) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Liquid(pos); ok {
		var updated bool
		for i, it := range c.Items {
			if !it.Item.Empty() {
				dropItem(tx, it.Item, pos.Vec3Centre())
				c.Items[i].Item, updated = item.Stack{}, true
			}
		}
		if !c.Extinguished {
			c.extinguish(pos, tx)
		} else if updated {
			tx.SetBlock(pos, c, nil)
		}
		return
	}
	if liquid, ok := tx.Liquid(pos.Side(cube.FaceUp)); ok && liquid.LiquidType() == "water" && !c.Extinguished {
		c.extinguish(pos, tx)
	}
}

// EntityInside ...
func (c Campfire) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if flammable, ok := e.(flammableEntity); ok {
		if flammable.OnFireDuration() > 0 && c.Extinguished {
			c.Extinguished = false
			tx.PlaySound(pos.Vec3(), sound.Ignite{})
			tx.SetBlock(pos, c, nil)
		}
		if !c.Extinguished {
			if l, ok := e.(livingEntity); ok {
				l.Hurt(c.Type.Damage(), FireDamageSource{})
			}
		}
	}
}

// EncodeNBT ...
func (c Campfire) EncodeNBT() map[string]any {
	m := map[string]any{"id": "Campfire"}
	for i, v := range c.Items {
		id := strconv.Itoa(i + 1)
		if !v.Item.Empty() {
			m["Item"+id] = nbtconv.WriteItem(v.Item, true)
			m["ItemTime"+id] = uint8(v.Time.Milliseconds() / 50)
		}
	}
	return m
}

// DecodeNBT ...
func (c Campfire) DecodeNBT(data map[string]any) any {
	for i := 0; i < 4; i++ {
		id := strconv.Itoa(i + 1)
		c.Items[i] = CampfireItem{
			Item: nbtconv.MapItem(data, "Item"+id),
			Time: time.Duration(nbtconv.Int16(data, "ItemTime"+id)) * time.Millisecond * 50,
		}
	}
	return c
}

// EncodeItem ...
func (c Campfire) EncodeItem() (name string, meta int16) {
	switch c.Type {
	case NormalFire():
		return "minecraft:campfire", 0
	case SoulFire():
		return "minecraft:soul_campfire", 0
	}
	panic("invalid fire type")
}

// EncodeBlock ...
func (c Campfire) EncodeBlock() (name string, properties map[string]any) {
	switch c.Type {
	case NormalFire():
		name = "minecraft:campfire"
	case SoulFire():
		name = "minecraft:soul_campfire"
	}
	return name, map[string]any{
		"minecraft:cardinal_direction": c.Facing.String(),
		"extinguished":                 c.Extinguished,
	}
}

// allCampfires ...
func allCampfires() (campfires []world.Block) {
	for _, d := range cube.Directions() {
		for _, f := range FireTypes() {
			campfires = append(campfires, Campfire{Facing: d, Type: f, Extinguished: true})
			campfires = append(campfires, Campfire{Facing: d, Type: f})
		}
	}
	return campfires
}
