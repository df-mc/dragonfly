package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

type dispenserBehaviour func(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool

// dispenserBehaviours is keyed by the encoded item name so variants such as tipped arrows and potions use the same
// behaviour while retaining the data stored in their concrete item value.
var dispenserBehaviours = map[string]dispenserBehaviour{
	"minecraft:arrow":             dispenseProjectile,
	"minecraft:bone_meal":         dispenseBoneMeal,
	"minecraft:bucket":            dispenseBucket,
	"minecraft:egg":               dispenseProjectile,
	"minecraft:experience_bottle": dispenseProjectile,
	"minecraft:firework_rocket":   dispenseProjectile,
	"minecraft:flint_and_steel":   dispenseIgnition,
	"minecraft:glass_bottle":      dispenseBottle,
	"minecraft:honeycomb":         dispenseHoneycomb,
	"minecraft:lingering_potion":  dispenseProjectile,
	"minecraft:snowball":          dispenseProjectile,
	"minecraft:splash_potion":     dispenseProjectile,
	"minecraft:tnt":               dispenseTNT,
	"minecraft:lava_bucket":       dispenseBucket,
	"minecraft:water_bucket":      dispenseBucket,
}

func dispenseBoneMeal(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, _ *rand.Rand) bool {
	front := pos.Side(d.Facing)
	affected, ok := tx.Block(front).(item.BoneMealAffected)
	if !ok {
		return false
	}
	result := affected.BoneMeal(front, tx)
	if result == item.BoneMealResultNone || d.inventory.SetItem(slot, stack.Grow(-1)) != nil {
		return false
	}
	tx.AddParticle(front.Vec3(), particle.BoneMeal{Area: result == item.BoneMealResultArea})
	return true
}

// dispenserWaxable is a block that a dispensed honeycomb can wax, mirroring the unexported interface item.Honeycomb
// uses for the same purpose.
type dispenserWaxable interface {
	Wax(cube.Pos, mgl64.Vec3) (world.Block, bool)
}

func dispenseHoneycomb(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, _ *rand.Rand) bool {
	front := pos.Side(d.Facing)
	waxable, ok := tx.Block(front).(dispenserWaxable)
	if !ok {
		return false
	}
	result, ok := waxable.Wax(front, pos.Vec3Centre())
	if !ok || d.inventory.SetItem(slot, stack.Grow(-1)) != nil {
		return false
	}
	tx.SetBlock(front, result, nil)
	tx.PlaySound(front.Vec3Centre(), sound.SignWaxed{})
	return true
}

type dispenserBottleFiller interface {
	FillBottle() (world.Block, item.Stack, bool)
}

func dispenseBottle(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	front := pos.Side(d.Facing)
	b := tx.Block(front)
	filler, fromBlock := b.(dispenserBottleFiller)
	if !fromBlock {
		liquid, ok := tx.Liquid(front)
		if !ok {
			return d.dropDispensedItem(slot, stack, pos, tx, r)
		}
		if filler, ok = liquid.(dispenserBottleFiller); !ok {
			return d.dropDispensedItem(slot, stack, pos, tx, r)
		}
	}
	result, filled, ok := filler.FillBottle()
	if !ok {
		return d.dropDispensedItem(slot, stack, pos, tx, r)
	}
	// Only a block filler leaves a replacement block behind. Bottling a liquid never consumes the source.
	if fromBlock && result != b {
		tx.SetBlock(front, result, nil)
	}
	return d.replaceDispensedItem(slot, stack, filled, pos, tx, r)
}

// dispenserIgnitable is a block that dispensed flint and steel can ignite directly, mirroring the unexported interface
// item.FlintAndSteel uses for the same purpose.
type dispenserIgnitable interface {
	Ignite(cube.Pos, *world.Tx, world.Entity) bool
}

func dispenseIgnition(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, _ *rand.Rand) bool {
	front := pos.Side(d.Facing)
	if target, ok := tx.Block(front).(dispenserIgnitable); ok {
		if !target.Ignite(front, tx, nil) {
			return false
		}
	} else {
		if _, alreadyLit := tx.Block(front).(Fire); alreadyLit {
			return false
		}
		Fire{}.Start(tx, front)
		if _, ok := tx.Block(front).(Fire); !ok {
			return false
		}
		tx.PlaySound(front.Vec3Centre(), sound.Ignite{})
	}
	if err := d.inventory.SetItem(slot, stack.Damage(1)); err != nil {
		return false
	}
	return true
}

func dispenseTNT(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, _ *rand.Rand) bool {
	create := tx.World().EntityRegistry().Config().TNT
	if create == nil {
		return false
	}
	front := pos.Side(d.Facing)
	b := tx.Block(front)
	if _, air := b.(Air); !air {
		replaceable, ok := b.(Replaceable)
		if !ok || !replaceable.ReplaceableBy(TNT{}) {
			return false
		}
	}
	if err := d.inventory.SetItem(slot, stack.Grow(-1)); err != nil {
		return false
	}
	tx.AddEntity(create(world.EntitySpawnOpts{Position: front.Vec3Centre()}, time.Second*4))
	tx.PlaySound(front.Vec3Centre(), sound.TNT{})
	return true
}

func dispenseBucket(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	bucket, ok := stack.Item().(item.Bucket)
	if !ok {
		return false
	}
	front := pos.Side(d.Facing)
	if bucket.Empty() {
		liquid, ok := tx.Liquid(front)
		if !ok || liquid.LiquidDepth() != 8 || liquid.LiquidFalling() {
			return d.dropDispensedItem(slot, stack, pos, tx, r)
		}
		filled := item.NewStack(item.Bucket{Content: item.LiquidBucketContent(liquid)}, 1)
		if !d.replaceDispensedItem(slot, stack, filled, pos, tx, r) {
			return false
		}
		tx.SetLiquid(front, nil)
		tx.PlaySound(front.Vec3Centre(), sound.BucketFill{Liquid: liquid})
		return true
	}

	liquid, ok := bucket.Content.Liquid()
	if !ok {
		return d.dropDispensedItem(slot, stack, pos, tx, r)
	}
	liquid = liquid.WithDepth(8, false)
	if !canDispenserPlaceLiquid(tx.Block(front), liquid) {
		return d.dropDispensedItem(slot, stack, pos, tx, r)
	}
	if err := d.inventory.SetItem(slot, item.NewStack(item.Bucket{}, 1)); err != nil {
		return false
	}
	tx.SetLiquid(front, liquid)
	tx.PlaySound(front.Vec3Centre(), sound.BucketEmpty{Liquid: liquid})
	return true
}

func canDispenserPlaceLiquid(b world.Block, liquid world.Liquid) bool {
	if displacer, ok := b.(world.LiquidDisplacer); ok && displacer.CanDisplace(liquid) {
		return true
	}
	if replaceable, ok := b.(Replaceable); ok {
		return replaceable.ReplaceableBy(liquid)
	}
	return false
}

func dispenseProjectile(d Dispenser, slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	conf := tx.World().EntityRegistry().Config()
	opts := dispenserProjectileOpts(pos, d.Facing, r)
	var handle *world.EntityHandle
	var played world.Sound = sound.ItemThrow{}

	switch it := stack.Item().(type) {
	case item.Arrow:
		if conf.Arrow == nil {
			return false
		}
		handle = conf.Arrow(opts, world.ArrowSpawnConfig{Damage: 2, ObtainArrowOnPickup: true, Tip: it.Tip})
		played = sound.BowShoot{}
	case item.Egg:
		if conf.Egg == nil {
			return false
		}
		handle = conf.Egg(opts, nil)
	case item.Snowball:
		if conf.Snowball == nil {
			return false
		}
		handle = conf.Snowball(opts, nil)
	case item.SplashPotion:
		if conf.SplashPotion == nil {
			return false
		}
		handle = conf.SplashPotion(opts, it.Type, nil)
	case item.LingeringPotion:
		if conf.LingeringPotion == nil {
			return false
		}
		handle = conf.LingeringPotion(opts, it.Type, nil)
	case item.BottleOfEnchanting:
		if conf.BottleOfEnchanting == nil {
			return false
		}
		handle = conf.BottleOfEnchanting(opts, nil)
	case item.Firework:
		if conf.Firework == nil {
			return false
		}
		handle = conf.Firework(opts, it, nil, 1.15, 0.04, false)
		played = sound.FireworkLaunch{}
	default:
		return false
	}
	if err := d.inventory.SetItem(slot, stack.Grow(-1)); err != nil {
		return false
	}
	tx.AddEntity(handle)
	tx.PlaySound(pos.Vec3Centre(), played)
	return true
}

func dispenserProjectileOpts(pos cube.Pos, facing cube.Face, r *rand.Rand) world.EntitySpawnOpts {
	direction := dispenserDirection(facing)
	return world.EntitySpawnOpts{
		Position: pos.Vec3Centre().Add(direction.Mul(0.7)),
		Velocity: direction.Mul(1.1).Add([3]float64{r.Float64()*0.1 - 0.05, r.Float64()*0.1 - 0.05, r.Float64()*0.1 - 0.05}),
	}
}
