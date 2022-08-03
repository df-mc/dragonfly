package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"strconv"
	"time"
)

// Campfire is a block that can be used to cook food, pacify bees, act as a spread-proof light source, smoke signal or
// damaging trap block.
type Campfire struct {
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
func (c Campfire) Model() world.BlockModel {
	return model.Campfire{}
}

// SideClosed ...
func (c Campfire) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (c Campfire) BreakInfo() BreakInfo {
	return newBreakInfo(5, alwaysHarvestable, axeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		var drops []item.Stack
		if hasSilkTouch(enchantments) {
			drops = append(drops, item.NewStack(c, 1))
		} else {
			switch c.Type {
			case NormalFire():
				drops = append(drops, item.NewStack(item.Charcoal{}, 2))
			case SoulFire():
				drops = append(drops, item.NewStack(SoulSoil{}, 1))
			}
		}
		for _, v := range c.Items {
			if !v.Item.Empty() {
				drops = append(drops, v.Item)
			}
		}
		return drops
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
func (c Campfire) Ignite(pos cube.Pos, w *world.World) bool {
	w.PlaySound(pos.Vec3(), sound.Ignite{})
	if _, ok := w.Liquid(pos); ok {
		return true
	}

	c.Extinguished = false
	w.SetBlock(pos, c, nil)
	return true
}

// Splash checks to see if the fire was splashed by a bottle and then extinguishes itself.
func (c Campfire) Splash(pos cube.Pos, e world.Entity, t potion.Potion) {
	if t != potion.Water() {
		// Water is the only potion that can extinguish a campfire.
		return
	}

	w := e.World()
	w.PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})

	c.Extinguished = true
	w.SetBlock(pos, c, nil)
}

// Activate ...
func (c Campfire) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if held, _ := u.HeldItems(); !held.Empty() {
		if rawFood, ok := held.Item().(item.Smeltable); ok && rawFood.SmeltInfo().Food {
			for i, it := range c.Items {
				if it.Item.Empty() {
					c.Items[i] = CampfireItem{
						Item: held.Grow(-held.Count() + 1),
						Time: time.Second * 30,
					}

					ctx.SubtractFromCount(1)

					w.PlaySound(pos.Vec3Centre(), sound.ItemAdd{})
					w.SetBlock(pos, c, nil)
					return true
				}
			}
		}
	}
	return false
}

// UseOnBlock ...
func (c Campfire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	if _, ok := w.Block(pos).(Campfire); ok && face == cube.FaceUp {
		return false
	}
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return
	}
	c.Facing = user.Facing().Opposite()
	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// Tick is called to cook the items within the campfire.
func (c Campfire) Tick(_ int64, pos cube.Pos, w *world.World) {
	if c.Extinguished {
		// Extinguished, do nothing.
		return
	}
	if rand.Float64() <= 0.016 { // Every three or so seconds.
		w.PlaySound(pos.Vec3Centre(), sound.CampfireCrackle{})
	}

	for i, it := range c.Items {
		if !it.Item.Empty() && it.Time <= 0 {
			if food, ok := it.Item.Item().(item.Smeltable); ok {
				dropItem(w, food.SmeltInfo().Product, pos.Vec3Middle())
			}

			c.Items[i].Item = item.Stack{}
			w.SetBlock(pos, c, nil)
			continue
		}
		c.Items[i].Time = it.Time - time.Millisecond*50
		w.SetBlock(pos, c, nil)
	}
}

// NeighbourUpdateTick ...
func (c Campfire) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Liquid(pos); ok && !c.Extinguished {
		c.Extinguished = true
		w.PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})
		w.SetBlock(pos, c, nil)
	}
}

// EntityInside ...
func (c Campfire) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	if flammable, ok := e.(flammableEntity); ok {
		if flammable.OnFireDuration() > 0 && c.Extinguished {
			c.Extinguished = false
			w.PlaySound(pos.Vec3(), sound.Ignite{})
			w.SetBlock(pos, c, nil)
		}
		if !c.Extinguished && !w.RainingAt(pos) {
			if l, ok := e.(livingEntity); ok && !l.AttackImmune() {
				l.Hurt(c.Type.Damage(), damage.SourceFire{Campfire: true})
			}
			if flammable.OnFireDuration() < time.Second*8 {
				flammable.SetOnFire(8 * time.Second)
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
			Time: time.Duration(nbtconv.Map[byte](data, "ItemTime"+id)) * time.Millisecond * 50,
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
		"direction":    int32(horizontalDirection(c.Facing)),
		"extinguished": c.Extinguished,
	}
}

// allCampfires ...
func allCampfires() (campfires []world.Block) {
	for _, d := range cube.Directions() {
		for _, e := range []bool{true, false} {
			campfires = append(campfires, Campfire{Facing: d, Extinguished: e, Type: NormalFire()})
			campfires = append(campfires, Campfire{Facing: d, Extinguished: e, Type: SoulFire()})
		}
	}
	return campfires
}
