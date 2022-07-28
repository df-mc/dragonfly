package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
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
	// Items represents the items in the campfire that are being cooked.
	Items [4]CampfireItem
	// Facing represents the direction that the campfire is facing.
	Facing cube.Direction
	// Extinguished is true if the campfire was extinguished by a water source.
	Extinguished bool
	// Type represents the type of Campfire, currently there are Normal and Soul campfires.
	Type FireType
}

// CampfireItem is an object that holds the data about the items in the campfire.
type CampfireItem struct {
	Item item.Stack
	// Time is the countdown of ticks until the item is cooked (when 0).
	Time int
}

// Model ...
func (c Campfire) Model() world.BlockModel {
	return model.Campfire{}
}

// CanDisplace ...
func (c Campfire) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (c Campfire) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// Splash checks to see if the fire was splashed by a bottle and then extinguishes itself
func (c Campfire) Splash(pos cube.Pos, p *entity.SplashPotion) {
	if p.Type() != potion.Water() {
		return
	}
	w := p.World()
	c.Extinguished = true
	p.World().PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})
	w.SetBlock(pos, c, nil)
}

// BreakInfo ...
func (c Campfire) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		var drops []item.Stack
		if hasSilkTouch(enchantments) {
			drops = append(drops, item.NewStack(c, 1))
		} else {
			switch c.Type {
			case NormalFire():
				drops = append(drops, item.NewStack(item.Charcoal{}, 2))
			case SoulFire():
				drops = append(drops, item.NewStack(SoulSoil{}, 1))
			default:
				panic("invalid fire type")
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
		return false
	}
	c.Extinguished = false
	w.SetBlock(pos, c, nil)
	return true
}

// Activate ...
func (c Campfire) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if held, _ := u.HeldItems(); !held.Empty() {
		if _, ok := held.Enchantment(enchantment.FireAspect{}); ok {
			c.Ignite(pos, w)
			return true
		}
		if rawFood, ok := held.Item().(item.Smeltable); ok && rawFood.SmeltInfo().Food {
			for i, it := range c.Items {
				if it.Item.Empty() {
					c.Items[i] = CampfireItem{
						Item: held.Grow(-held.Count() + 1),
						Time: 600,
					}
					ctx.SubtractFromCount(1)
					w.PlaySound(pos.Vec3Centre(), sound.AddItem{})
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

// Tick is called to cook the items within the campfire
func (c Campfire) Tick(_ int64, pos cube.Pos, w *world.World) {
	if !c.Extinguished {
		if rand.Float64() <= 0.016 { // Every three or so seconds.
			w.PlaySound(pos.Vec3Centre(), sound.CampfireCrackle{})
		}
		for i, it := range c.Items {
			if !it.Item.Empty() && it.Time <= 0 {
				itemCooked := it.Item
				if food, ok := itemCooked.Item().(item.Smeltable); ok {
					ent := entity.NewItem(food.SmeltInfo().Product, pos.Vec3Middle())
					ent.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
					w.AddEntity(ent)
					c.Items[i].Item = item.Stack{}
					w.SetBlock(pos, c, nil)
					continue
				}
			}
			c.Items[i].Time = it.Time - 1
			w.SetBlock(pos, c, nil)
		}
	}
}

// NeighbourUpdateTick ...
func (c Campfire) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	// if the campfire is water logged we extinguish it
	if _, ok := w.Liquid(pos); ok && !c.Extinguished {
		c.Extinguished = true
		w.PlaySound(pos.Vec3Centre(), sound.FireExtinguish{})
		w.SetBlock(pos, c, nil)
	}
}

// EntityInside ...
func (c Campfire) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	if flammable, ok := e.(entity.Flammable); ok {
		// Try to egnite the campfire is the entity is on fire and ontop
		if flammable.OnFireDuration() > 0 && c.Extinguished {
			c.Extinguished = false
			w.PlaySound(pos.Vec3(), sound.Ignite{})
			w.SetBlock(pos, c, nil)
		}
		if !c.Extinguished && !w.RainingAt(pos) {
			if l, ok := e.(entity.Living); ok && !l.AttackImmune() {
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
	m := map[string]any{
		"id": "Campfire",
	}
	for i, v := range c.Items {
		itemNumberIdentifier := strconv.Itoa(i + 1)
		if !v.Item.Empty() {
			m["Item"+itemNumberIdentifier] = nbtconv.WriteItem(v.Item, true)
			m["ItemTime"+itemNumberIdentifier] = uint8(v.Time)
		}
	}
	switch c.Type {
	case NormalFire():
		m["id"] = "Campfire"
		return m
	case SoulFire():
		m["id"] = "SoulFire"
		return m
	}
	panic("invalid fire type")
}

// DecodeNBT ...
func (c Campfire) DecodeNBT(data map[string]any) any {
	for i := 0; i < 4; i++ {
		itemNumberIdentifier := strconv.Itoa(i + 1)
		c.Items[i] = CampfireItem{
			Item: nbtconv.MapItem(data, "Item"+itemNumberIdentifier),
			Time: int(nbtconv.Map[byte](data, "ItemTime"+itemNumberIdentifier)),
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
	default:
		panic("invalid fire type")
	}
	return name, map[string]any{
		"direction":    int32(horizontalDirection(c.Facing)),
		"extinguished": c.Extinguished,
	}
}

func allCampfires() (campfires []world.Block) {
	for _, d := range cube.Directions() {
		for _, e := range []bool{true, false} {
			campfires = append(campfires, Campfire{Facing: d, Extinguished: e, Type: NormalFire()})
			campfires = append(campfires, Campfire{Facing: d, Extinguished: e, Type: SoulFire()})
		}
	}
	return campfires
}
