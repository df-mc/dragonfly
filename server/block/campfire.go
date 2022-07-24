package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"strconv"
	"time"
)

// Campfire is a block that can be used to cook food, pacify bees, act as a spread-proof light source, smoke signal or
// damaging trap block.
type Campfire struct {
	// Items represents the items in the campfire that are getting cooked
	Items [4]CampfireItem
	// Facing represents the direction that the campfire is facing.
	Facing cube.Direction
	// Extinguished is true if the campfire was extinguished by a water source.
	Extinguished bool
	// Type represents the type of Campfire, currently there are Normal and Soul campfires
	Type FireType
}

// An object that holds the data about the items in the campfire
type CampfireItem struct {
	Item item.Stack
	// Time is the countdown of ticks until the item is cooked (when 0)
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
	switch c.Type {
	case NormalFire():
		return 15
	case SoulFire():
		return 10
	}
	panic("invalid fire type")
}

// Activate ...
func (c Campfire) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if held, _ := u.HeldItems(); !held.Empty() {
		if _, ok := held.Item().(item.FlintAndSteel); ok {
			c.Extinguished = false
			w.SetBlock(pos, c, nil)
			//TODO: Egnite Sound
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
					w.SetBlock(pos, c, nil)
					return true
				}
			}
		}
	}
	return true
}

// UseOnBlock ...
func (c Campfire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
	//TODO: check to see if ember particles are client side
	if !c.Extinguished {
		// if the campfire is water logged we extinguish it
		if _, ok := w.Liquid(pos); ok {
			c.Extinguished = true
			w.SetBlock(pos, c, nil)
		}
		for i, it := range c.Items {
			if !it.Item.Empty() && it.Time <= 0 {
				itemCooked := it.Item
				if food, ok := itemCooked.Item().(item.Smeltable); ok {
					//TODO: Play Pop Sound
					w.AddEntity(entity.NewItem(food.SmeltInfo().Product, pos.Vec3Middle()))
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

// EntityInside ...
func (c Campfire) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	if flammable, ok := e.(entity.Flammable); ok {
		// Try to egnite the campfire is the entity is on fire and ontop
		if flammable.OnFireDuration() > 0 && c.Extinguished {
			c.Extinguished = false
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
